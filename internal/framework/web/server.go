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

	"github.com/tacoda/keystone/internal/framework/config"
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
		Addr:              addr,
		Handler:           srv.mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(shutdown)
	}()

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
}

func newServer(projectDir string) (*server, error) {
	tmplFS, err := fs.Sub(embedded, "templates")
	if err != nil {
		return nil, err
	}
	hub := newSSEHub()
	watcher, err := newFSWatcher(projectDir, hub)
	if err != nil {
		return nil, err
	}

	s := &server{
		projectDir: projectDir,
		mux:        http.NewServeMux(),
		tmplFS:     tmplFS,
		funcs: template.FuncMap{
			"join": strings.Join,
		},
		hub:     hub,
		watcher: watcher,
	}
	s.routes()
	return s, nil
}

func (s *server) routes() {
	// Static assets.
	assetFS, _ := fs.Sub(embedded, "assets")
	s.mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assetFS))))

	// SSE hub.
	s.mux.HandleFunc("/events", s.hub.ServeHTTP)

	// HTMX dashboard pages.
	s.mux.HandleFunc("/", s.handleHome)
	s.mux.HandleFunc("/metrics", s.handleMetrics)
	s.mux.HandleFunc("/primitives", s.handlePrimitivesList)
	s.mux.HandleFunc("/primitives/new", s.handlePrimitivesNew)
	s.mux.HandleFunc("/primitives/", s.handlePrimitivesDetail)
	s.mux.HandleFunc("/policies", s.handlePolicies)
	s.mux.HandleFunc("/policies/investigate", s.handleInvestigator)
	s.mux.HandleFunc("/sources", s.handleSources)
	s.mux.HandleFunc("/sources/new", s.handleSourcesNew)
	s.mux.HandleFunc("/sources/", s.handleSourceDetail)
	s.mux.HandleFunc("/verify", s.handleVerifyPage)
	s.mux.HandleFunc("/prune", s.handlePrune)
	s.mux.HandleFunc("/inbox", s.handleInbox)
	s.mux.HandleFunc("/flywheels", s.handleFlywheels)
	s.mux.HandleFunc("/evals", s.handleEvals)
	s.mux.HandleFunc("/search", s.handleSearch)
	s.mux.HandleFunc("/graph", s.handleGraph)
	s.mux.HandleFunc("/web/fragments/search", s.handleSearchFragment)
	s.mux.HandleFunc("/api/search", s.apiSearch)
	s.mux.HandleFunc("/api/evals", s.apiEvals)
	s.mux.HandleFunc("/api/evals/run", s.apiEvalRun)
	s.mux.HandleFunc("/web/actions/eval/run", s.handleActionEvalRun)
	s.mux.HandleFunc("/insights", s.handleInsights)
	s.mux.HandleFunc("/api/insights", s.apiInsights)

	// REST API (read-only).
	s.mux.HandleFunc("/api/index", s.apiIndex)
	s.mux.HandleFunc("/api/primitives", s.apiPrimitives)
	s.mux.HandleFunc("/api/primitives/", s.apiPrimitiveDetail)
	s.mux.HandleFunc("/api/sources", s.apiSources)
	s.mux.HandleFunc("/api/sources/", s.apiSourceDetail)
	s.mux.HandleFunc("/api/harness/status", s.apiHarnessStatus)
	s.mux.HandleFunc("/api/metrics", s.apiMetrics)

	// HTMX fragment endpoints + write actions.
	s.mux.HandleFunc("/web/fragments/primitives", s.handlePrimitivesFragment)
	s.mux.HandleFunc("/web/fragments/investigator", s.handleInvestigatorFragment)
	s.mux.HandleFunc("/web/actions/primitives/new", s.handleActionNewPrimitive)
	s.mux.HandleFunc("/web/actions/primitives/delete", s.handleActionDeletePrimitive)
	s.mux.HandleFunc("/web/actions/policy/add", s.handleActionPolicyAdd)
	s.mux.HandleFunc("/web/actions/policy/remove", s.handleActionPolicyRemove)
	s.mux.HandleFunc("/web/actions/verify", s.handleActionVerify)
	s.mux.HandleFunc("/web/actions/sources/add", s.handleActionSourceAdd)
	s.mux.HandleFunc("/web/actions/sources/remove", s.handleActionSourceRemove)
	s.mux.HandleFunc("/web/actions/sources/query", s.handleActionSourceQuery)
	s.mux.HandleFunc("/web/actions/sources/health", s.handleActionSourceHealth)
	s.mux.HandleFunc("/web/actions/sources/verify-all", s.handleActionSourceVerifyAll)
	s.mux.HandleFunc("/web/actions/inbox/accept", s.handleActionInboxAccept)
	s.mux.HandleFunc("/web/actions/inbox/reject", s.handleActionInboxReject)
}

// loadPrimitives is the shared read path for both API + dashboard
// handlers. Walks the harness on every request — cheap and avoids
// stale state.
func (s *server) loadPrimitives() ([]primitive.Primitive, error) {
	primitives, _, err := primitive.Walk(s.projectDir, config.DefaultHarnessRoot)
	return primitives, err
}

// buildIndex is the same envelope `keystone index` emits, computed
// fresh per request.
func (s *server) buildIndex() (primitive.Index, error) {
	primitives, err := s.loadPrimitives()
	if err != nil {
		return primitive.Index{}, err
	}
	return primitive.Build(primitives, time.Now()), nil
}
