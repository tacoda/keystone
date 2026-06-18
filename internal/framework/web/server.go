// Package web hosts the localhost HTMX dashboard + read-only REST
// API for a keystone install. Single port (default 4773), single
// origin: REST under /api/, HTML dashboard under /, SSE push at
// /events. Writes happen via HTMX form posts under /web/actions/
// (deferred — read-only in v1).
//
// Same process as the keystone CLI; reuses
// `internal/framework/primitive` for the data model and
// `internal/framework/mcp/adapter` for source health.
package web

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

// DefaultPort is the localhost port `keystone web serve` binds to
// when no --port flag is supplied. 4773 = "KEYS" on a phone keypad
// — easy to recall, low conflict probability in the registered
// range.
const DefaultPort = 4773

//go:embed assets/* templates/*.html
var embedded embed.FS

// Options configures the server.
type Options struct {
	// ProjectDir is the consumer project root. Defaults to cwd at
	// Serve time when empty.
	ProjectDir string

	// Port is the localhost port to bind. Zero falls back to
	// DefaultPort.
	Port int
}

// Serve runs the dashboard + REST API + SSE hub on localhost.
// Blocks until the context is cancelled.
func Serve(ctx context.Context, opts Options) error {
	if opts.ProjectDir == "" {
		opts.ProjectDir = "."
	}
	abs, err := filepath.Abs(opts.ProjectDir)
	if err != nil {
		return fmt.Errorf("abs project dir: %w", err)
	}
	if opts.Port == 0 {
		opts.Port = DefaultPort
	}

	srv, err := newServer(abs)
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("127.0.0.1:%d", opts.Port)
	httpSrv := &http.Server{
		Addr:    addr,
		Handler: withTimeoutExceptSSE(srv.engine),
		// ReadHeaderTimeout caps how long a slow client can take to
		// finish the request headers. IdleTimeout closes keep-alive
		// connections that go quiet — bounded recovery from leaked
		// connections without affecting in-progress requests. We
		// deliberately leave WriteTimeout unset because /events is a
		// long-lived SSE stream; per-handler caps live in the
		// timeoutMiddleware applied in routes().
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(shutdown)
		// Close the audit log only after the HTTP server has drained,
		// so any in-flight publish-driven Append still lands.
		_ = srv.audit.Close()
	}()

	srv.startCacheRefreshers(ctx)
	srv.watcher.Start(ctx)

	fmt.Printf("keystone web — http://%s  (project: %s)\n", addr, abs)
	err = httpSrv.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

type server struct {
	projectDir string
	engine     *gin.Engine
	tmplFS     fs.FS
	funcs      template.FuncMap
	hub        *sseHub
	watcher    *fsWatcher
	audit      *auditLog

	primitiveCache *primitiveCache
	healthCache    *healthCache
}

func newServer(projectDir string) (*server, error) {
	tmplFS, err := fs.Sub(embedded, "templates")
	if err != nil {
		return nil, err
	}
	hub := newSSEHub()

	// gin.ReleaseMode suppresses the debug banner + per-route
	// registration log; default writer is io.Discard so gin doesn't
	// interleave its own request logs with ours.
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard
	engine := gin.New()
	engine.Use(gin.Recovery())

	// Open the per-session audit log. Best-effort: if the audit
	// directory can't be created (read-only project, permissions),
	// the dashboard still runs — audit features render an empty
	// state instead.
	audit, auditErr := openAuditLog(projectDir)
	if auditErr != nil {
		fmt.Fprintf(os.Stderr, "keystone web: audit disabled: %v\n", auditErr)
		audit = nil
	}

	s := &server{
		projectDir: projectDir,
		engine:     engine,
		tmplFS:     tmplFS,
		funcs: template.FuncMap{
			"join": strings.Join,
		},
		hub:            hub,
		audit:          audit,
		primitiveCache: newPrimitiveCache(projectDir),
		healthCache:    newHealthCache(projectDir),
	}

	// Watcher rebuilds every cached harness layer on each debounced
	// file-change burst. Per-layer dirty matching isn't worth the
	// complexity: a debounce already collapses bursts, and walks of
	// the small per-layer trees are cheap relative to per-request
	// walks.
	watcher, err := newFSWatcher(projectDir, hub, s.refreshAllLayers)
	if err != nil {
		return nil, err
	}
	// Audit log lives at the watcher boundary: every published burst
	// produces one JSONL line. The hook is set after construction so
	// the watcher type doesn't gain a constructor parameter that's
	// only ever wired from one call site.
	if s.audit != nil {
		watcher.onPublish = func(paths []string, topics []sseTopic) {
			rels := make([]string, 0, len(paths))
			for _, p := range paths {
				rels = append(rels, relPath(projectDir, p))
			}
			ts := make([]string, 0, len(topics))
			for _, t := range topics {
				ts = append(ts, string(t))
			}
			s.audit.Append(auditEntry{
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Topics:    ts,
				Paths:     rels,
				Summary:   summarizeAudit(rels, ts),
			})
		}
	}
	s.watcher = watcher

	s.routes()
	return s, nil
}

func (s *server) routes() {
	// Static assets via gin's StaticFS, wired to the embedded FS sub.
	// Assets ship inside the binary (embed.FS) and are version-pinned
	// to the release tag — safe to mark immutable so repeat hits skip
	// the network entirely. Critical for "snappy" — devs notice every
	// 30ms of latency on a tool they use all day.
	if assetFS, err := fs.Sub(embedded, "assets"); err == nil {
		s.engine.Use(func(c *gin.Context) {
			if strings.HasPrefix(c.Request.URL.Path, "/assets/") {
				c.Header("Cache-Control", "public, max-age=31536000, immutable")
			}
			c.Next()
		})
		s.engine.StaticFS("/assets", http.FS(assetFS))
	}

	// SSE hub. The per-request timeout lives in the
	// http.Server.Handler wrapper in Serve() — that wrapper exempts
	// /events by path so the long-lived stream is never wrapped.
	s.engine.GET("/events", gin.WrapF(s.hub.ServeHTTP))

	// Exact-match routes. Every handler keeps its (w, r) signature;
	// gin is a routing surface, not a handler-shape rewrite.
	// Page routes — 5 sections × tabs. Old single-purpose URLs are
	// retired in 2.1.1; nothing redirects, nothing aliases.
	//   / + /observability/*  → metrics, insights, audit, live
	//   /harness/*            → primitives, policies, investigator, graph
	//   /sources, /sources/new, /sources/<name>
	//   /flywheels/*          → overview, inbox, prune
	//   /quality/*            → verify, evals
	exact := []routeBinding{
		// Observability — `/` is the dashboard landing.
		{"/", s.handleMetrics},
		{"/observability", s.handleMetrics},
		{"/observability/metrics", s.handleMetrics},
		{"/observability/insights", s.handleInsights},

		// Harness.
		{"/harness", s.handlePrimitivesList},
		{"/harness/primitives", s.handlePrimitivesList},
		{"/harness/primitives/new", s.handlePrimitivesNew},
		{"/harness/policies", s.handlePolicies},
		{"/harness/investigator", s.handleInvestigator},
		{"/harness/graph", s.handleGraph},

		// Sources.
		{"/sources", s.handleSources},
		{"/sources/new", s.handleSourcesNew},

		// Flywheels.
		{"/flywheels", s.handleFlywheels},
		{"/flywheels/inbox", s.handleInbox},
		{"/flywheels/prune", s.handlePrune},

		// Quality.
		{"/quality", s.handleVerifyPage},
		{"/quality/verify", s.handleVerifyPage},
		{"/quality/evals", s.handleEvals},

		// Topbar search (popover backed by /web/fragments/search; the
		// page URL kept as a bookmark-friendly fallback).
		{"/search", s.handleSearch},

		// On-demand widget fragments. KPI strip lives under
		// /web/widgets/kpi/<name> via the prefix dispatch below; the
		// graph fragment is a single exact route because it has no
		// per-instance variant.
		{"/web/widgets/graph", s.handleGraphWidget},
		{"/web/widgets/audit", s.handleAuditWidget},

		// REST API (read-only) exact routes.
		{"/api/index", s.apiIndex},
		{"/api/primitives", s.apiPrimitives},
		{"/api/sources", s.apiSources},
		{"/api/harness/status", s.apiHarnessStatus},
		{"/api/metrics", s.apiMetrics},
		{"/api/search", s.apiSearch},
		{"/api/evals", s.apiEvals},
		{"/api/evals/run", s.apiEvalRun},
		{"/api/insights", s.apiInsights},

		// HTMX fragments + write actions.
		{"/web/fragments/search", s.handleSearchFragment},
		{"/web/fragments/primitives", s.handlePrimitivesFragment},
		{"/web/fragments/investigator", s.handleInvestigatorFragment},
		{"/web/actions/eval/run", s.handleActionEvalRun},
		{"/web/actions/primitives/new", s.handleActionNewPrimitive},
		{"/web/actions/primitives/delete", s.handleActionDeletePrimitive},
		{"/web/actions/policy/add", s.handleActionPolicyAdd},
		{"/web/actions/policy/remove", s.handleActionPolicyRemove},
		{"/web/actions/verify", s.handleActionVerify},
		{"/web/actions/sources/add", s.handleActionSourceAdd},
		{"/web/actions/sources/remove", s.handleActionSourceRemove},
		{"/web/actions/sources/query", s.handleActionSourceQuery},
		{"/web/actions/sources/health", s.handleActionSourceHealth},
		{"/web/actions/sources/verify-all", s.handleActionSourceVerifyAll},
		{"/web/actions/inbox/accept", s.handleActionInboxAccept},
		{"/web/actions/inbox/reject", s.handleActionInboxReject},
	}
	for _, b := range exact {
		s.engine.Any(b.pattern, wrap(b.h))
	}

	// Prefix routes (ServeMux trailing-slash semantics). Gin's tree
	// rejects a catch-all sibling next to a static segment, so we
	// register prefixes through a NoRoute fallback instead. The
	// static sibling routes already registered above take precedence;
	// only requests that miss every static match land here.
	prefixes := []struct {
		prefix string
		h      http.HandlerFunc
	}{
		{"/harness/primitives/", s.handlePrimitivesDetail},
		{"/sources/", s.handleSourceDetail},
		{"/api/primitives/", s.apiPrimitiveDetail},
		{"/api/sources/", s.apiSourceDetail},
		{"/web/widgets/kpi/", s.handleKPIWidget},
	}
	s.engine.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		for _, p := range prefixes {
			if strings.HasPrefix(path, p.prefix) {
				// Gin's NoRoute defaults to 404; we matched a real
				// route via prefix, so flip the default before
				// handing off. The handler is still free to write a
				// different status (e.g. 404 on missing resource).
				c.Status(http.StatusOK)
				p.h(c.Writer, c.Request)
				return
			}
		}
		http.NotFound(c.Writer, c.Request)
	})
}

// routeBinding pairs a URL pattern with an http.HandlerFunc — the
// gin-side analogue of one old mux.Handle call.
type routeBinding struct {
	pattern string
	h       http.HandlerFunc
}

// wrap adapts a stdlib http.HandlerFunc into a gin.HandlerFunc so
// existing handler bodies stay untouched.
func wrap(h http.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		h(c.Writer, c.Request)
	}
}

// handlerTimeout caps every non-SSE request. Generous enough to
// cover an evicted-cache refill, tight enough that no handler can
// pin a client connection forever.
const handlerTimeout = 30 * time.Second

// withTimeoutExceptSSE wraps the engine in http.TimeoutHandler for
// every path except /events. /events is a long-lived SSE stream;
// wrapping it would kill the connection at the timeout mark.
func withTimeoutExceptSSE(h http.Handler) http.Handler {
	timed := http.TimeoutHandler(h, handlerTimeout, "request timed out")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/events" {
			h.ServeHTTP(w, r)
			return
		}
		timed.ServeHTTP(w, r)
	})
}

// loadPrimitives is the shared read path for both API + dashboard
// handlers. Reads from primitiveCache — the fsWatcher rebuilds the
// cache on every debounced file-change event, plus a slow ticker
// fallback in cache.go.
func (s *server) loadPrimitives() ([]primitive.Primitive, error) {
	prims, _, err := s.primitiveCache.get()
	return prims, err
}

// buildIndex is the same envelope `keystone index` emits, served
// from primitiveCache. Build cost is paid by the cache refresh, not
// the request goroutine.
func (s *server) buildIndex() (primitive.Index, error) {
	_, idx, err := s.primitiveCache.get()
	return idx, err
}
