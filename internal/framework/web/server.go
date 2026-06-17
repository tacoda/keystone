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
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
	"time"

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
		Handler: srv.mux,
		// ReadHeaderTimeout caps how long a slow client can take to
		// finish the request headers. IdleTimeout closes keep-alive
		// connections that go quiet — bounded recovery from leaked
		// connections without affecting in-progress requests. We
		// deliberately leave WriteTimeout unset because /events is a
		// long-lived SSE stream; per-handler caps live in the
		// TimeoutHandler middleware in routes().
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(shutdown)
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
	mux        *http.ServeMux
	tmplFS     fs.FS
	funcs      template.FuncMap
	hub        *sseHub
	watcher    *fsWatcher

	primitiveCache *primitiveCache
	healthCache    *healthCache
}

func newServer(projectDir string) (*server, error) {
	tmplFS, err := fs.Sub(embedded, "templates")
	if err != nil {
		return nil, err
	}
	hub := newSSEHub()

	s := &server{
		projectDir: projectDir,
		mux:        http.NewServeMux(),
		tmplFS:     tmplFS,
		funcs: template.FuncMap{
			"join": strings.Join,
		},
		hub:            hub,
		primitiveCache: newPrimitiveCache(projectDir),
		healthCache:    newHealthCache(projectDir),
	}

	// Watcher rebuilds the primitive cache on every debounced file
	// change. Wired before newFSWatcher returns so the first event
	// drives a refresh.
	watcher, err := newFSWatcher(projectDir, hub, s.primitiveCache.refresh)
	if err != nil {
		return nil, err
	}
	s.watcher = watcher

	s.routes()
	return s, nil
}

func (s *server) routes() {
	// Static assets.
	assetFS, _ := fs.Sub(embedded, "assets")
	s.mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assetFS))))

	// SSE hub — bare, no TimeoutHandler. /events is a long-lived
	// stream; wrapping it would kill the connection at the 30s mark.
	s.mux.HandleFunc("/events", s.hub.ServeHTTP)

	// HTMX dashboard pages. Wrapped via s.handle so a slow handler
	// returns 503 instead of holding the connection forever.
	s.handle("/", s.handleHome)
	s.handle("/metrics", s.handleMetrics)
	s.handle("/primitives", s.handlePrimitivesList)
	s.handle("/primitives/new", s.handlePrimitivesNew)
	s.handle("/primitives/", s.handlePrimitivesDetail)
	s.handle("/policies", s.handlePolicies)
	s.handle("/policies/investigate", s.handleInvestigator)
	s.handle("/sources", s.handleSources)
	s.handle("/sources/new", s.handleSourcesNew)
	s.handle("/sources/", s.handleSourceDetail)
	s.handle("/verify", s.handleVerifyPage)
	s.handle("/prune", s.handlePrune)
	s.handle("/inbox", s.handleInbox)
	s.handle("/flywheels", s.handleFlywheels)
	s.handle("/evals", s.handleEvals)
	s.handle("/search", s.handleSearch)
	s.handle("/graph", s.handleGraph)
	s.handle("/web/fragments/search", s.handleSearchFragment)
	s.handle("/api/search", s.apiSearch)
	s.handle("/api/evals", s.apiEvals)
	s.handle("/api/evals/run", s.apiEvalRun)
	s.handle("/web/actions/eval/run", s.handleActionEvalRun)
	s.handle("/insights", s.handleInsights)
	s.handle("/api/insights", s.apiInsights)

	// REST API (read-only).
	s.handle("/api/index", s.apiIndex)
	s.handle("/api/primitives", s.apiPrimitives)
	s.handle("/api/primitives/", s.apiPrimitiveDetail)
	s.handle("/api/sources", s.apiSources)
	s.handle("/api/sources/", s.apiSourceDetail)
	s.handle("/api/harness/status", s.apiHarnessStatus)
	s.handle("/api/metrics", s.apiMetrics)

	// HTMX fragment endpoints + write actions.
	s.handle("/web/fragments/primitives", s.handlePrimitivesFragment)
	s.handle("/web/fragments/investigator", s.handleInvestigatorFragment)
	s.handle("/web/actions/primitives/new", s.handleActionNewPrimitive)
	s.handle("/web/actions/primitives/delete", s.handleActionDeletePrimitive)
	s.handle("/web/actions/policy/add", s.handleActionPolicyAdd)
	s.handle("/web/actions/policy/remove", s.handleActionPolicyRemove)
	s.handle("/web/actions/verify", s.handleActionVerify)
	s.handle("/web/actions/sources/add", s.handleActionSourceAdd)
	s.handle("/web/actions/sources/remove", s.handleActionSourceRemove)
	s.handle("/web/actions/sources/query", s.handleActionSourceQuery)
	s.handle("/web/actions/sources/health", s.handleActionSourceHealth)
	s.handle("/web/actions/sources/verify-all", s.handleActionSourceVerifyAll)
	s.handle("/web/actions/inbox/accept", s.handleActionInboxAccept)
	s.handle("/web/actions/inbox/reject", s.handleActionInboxReject)
}

// handlerTimeout is the per-request ceiling enforced by the
// TimeoutHandler middleware wrapping every non-SSE route. Generous
// enough to cover an evicted-cache refill, tight enough that no
// handler can pin a client connection forever. Adjust if write
// actions that shell out to the keystone CLI ever need longer.
const handlerTimeout = 30 * time.Second

// handle registers an http.HandlerFunc wrapped in TimeoutHandler so
// a slow handler returns 503 instead of holding the client
// connection forever. Use this for every route except /events
// (SSE — wrapping would kill the long-lived stream) and /assets/
// (static file server, already non-blocking).
func (s *server) handle(pattern string, h http.HandlerFunc) {
	s.mux.Handle(pattern, http.TimeoutHandler(h, handlerTimeout, "request timed out"))
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
