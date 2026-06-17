package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Adapter is the contract for an external-source backend. Adapters
// satisfy stage 3 of the runtime resolution flow — fetching
// information from outside the harness when in-harness rules and
// corpus aren't enough.
//
// Implementations must be context-aware and fast-failing: the MCP
// server returns tool results inline, and a hung adapter blocks the
// agent.
type Adapter interface {
	// Name returns the user-facing source name (matches the entry in
	// .keystone/context.json).
	Name() string

	// Type returns the adapter type ("folder", "url", "linear", …).
	Type() string

	// Query runs a free-form query against the source. The result is
	// a markdown document the agent reads; empty body + no error
	// means "no hits."
	Query(ctx context.Context, query string) (QueryResult, error)

	// Health probes the adapter's reachability and auth state. Cheap;
	// callable on every health request.
	Health(ctx context.Context) Health
}

// QueryResult is what an adapter returns. Body is markdown; Title and
// Source URI help the agent attribute the finding when it asks the
// user how to apply it.
type QueryResult struct {
	Title    string `json:"title,omitempty"`
	SourceID string `json:"source_id,omitempty"`
	URI      string `json:"uri,omitempty"`
	Body     string `json:"body"`
}

// Health reports adapter reachability + auth state.
type Health struct {
	OK      bool   `json:"ok"`
	Status  string `json:"status,omitempty"`  // "ready" | "auth-missing" | "unreachable" | …
	Message string `json:"message,omitempty"`
}

// SourceConfig is one entry in .keystone/context.json's sources list.
// Free-form `Settings` keeps the adapter-specific fields portable
// without locking the schema.
type SourceConfig struct {
	Name     string         `json:"name"`
	Type     string         `json:"type"`
	Settings map[string]any `json:"-"`
}

// UnmarshalJSON pulls name + type into typed fields and lifts everything
// else into Settings so each adapter can read what it needs.
func (s *SourceConfig) UnmarshalJSON(data []byte) error {
	raw := map[string]any{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if v, ok := raw["name"].(string); ok {
		s.Name = v
	}
	if v, ok := raw["type"].(string); ok {
		s.Type = v
	}
	delete(raw, "name")
	delete(raw, "type")
	s.Settings = raw
	return nil
}

// ContextConfig is the shape of .keystone/context.json (the MCP
// runtime config; distinct from keystone.json).
type ContextConfig struct {
	Version int            `json:"version"`
	Sources []SourceConfig `json:"sources"`
}

// LoadContextConfig reads .keystone/context.json from projectDir.
// Returns (nil, nil) if the file does not exist — running without
// configured sources is valid; the agent just can't escalate to
// stage 3.
func LoadContextConfig(projectDir string) (*ContextConfig, error) {
	path := filepath.Join(projectDir, ".keystone", "context.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var cfg ContextConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &cfg, nil
}

// BuildAdapter is the registry — given a SourceConfig, returns the
// right Adapter implementation. Unknown types are not fatal at load
// time; the adapter slot just reports an unreachable health and
// returns errors on query.
//
// Exported so the web package (and any future host) can build
// adapters from a parsed config without going through the full
// loadAdapters path.
func BuildAdapter(s SourceConfig) Adapter {
	switch s.Type {
	case "folder":
		return newFolderAdapter(s)
	case "url":
		return newURLAdapter(s)
	default:
		return &unknownAdapter{name: s.Name, kind: s.Type}
	}
}

// loadAdapters reads the config and builds every declared adapter.
// Returns an empty slice if no config file exists.
func loadAdapters(projectDir string) ([]Adapter, error) {
	cfg, err := LoadContextConfig(projectDir)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, nil
	}
	out := make([]Adapter, 0, len(cfg.Sources))
	for _, s := range cfg.Sources {
		out = append(out, BuildAdapter(s))
	}
	return out, nil
}

// findAdapter returns the named adapter or an error if none match.
func findAdapter(adapters []Adapter, name string) (Adapter, error) {
	for _, a := range adapters {
		if a.Name() == name {
			return a, nil
		}
	}
	return nil, fmt.Errorf("no source named %q (configured: %s)", name, strings.Join(adapterNames(adapters), ", "))
}

func adapterNames(adapters []Adapter) []string {
	out := make([]string, 0, len(adapters))
	for _, a := range adapters {
		out = append(out, a.Name())
	}
	return out
}

// -- folder adapter ----------------------------------------------------

// folderAdapter reads markdown from a local directory. Useful for
// org wikis already exported to disk, or for testing the adapter
// surface without external services.
type folderAdapter struct {
	name string
	path string
}

func newFolderAdapter(s SourceConfig) Adapter {
	p, _ := s.Settings["path"].(string)
	return &folderAdapter{name: s.Name, path: p}
}

func (a *folderAdapter) Name() string { return a.name }
func (a *folderAdapter) Type() string { return "folder" }

func (a *folderAdapter) Health(ctx context.Context) Health {
	if a.path == "" {
		return Health{Status: "config-missing", Message: "no `path` set in source config"}
	}
	info, err := os.Stat(a.path)
	if err != nil {
		return Health{Status: "unreachable", Message: err.Error()}
	}
	if !info.IsDir() {
		return Health{Status: "unreachable", Message: a.path + " is not a directory"}
	}
	return Health{OK: true, Status: "ready"}
}

// Query walks the configured folder and returns the contents of every
// markdown file whose body or filename contains a case-insensitive
// substring of query. Cheap; not full-text-indexed.
func (a *folderAdapter) Query(ctx context.Context, query string) (QueryResult, error) {
	if h := a.Health(ctx); !h.OK {
		return QueryResult{}, fmt.Errorf("%s: %s", a.name, h.Message)
	}
	needle := strings.ToLower(strings.TrimSpace(query))
	var hits []string
	err := filepath.Walk(a.path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(p), ".md") {
			return nil
		}
		body, err := os.ReadFile(p)
		if err != nil {
			return nil
		}
		if needle == "" || strings.Contains(strings.ToLower(string(body)), needle) ||
			strings.Contains(strings.ToLower(info.Name()), needle) {
			rel, _ := filepath.Rel(a.path, p)
			hits = append(hits, fmt.Sprintf("## %s\n\n%s", rel, string(body)))
		}
		return nil
	})
	if err != nil {
		return QueryResult{}, err
	}
	if len(hits) == 0 {
		return QueryResult{Title: fmt.Sprintf("no hits for %q in %s", query, a.name), Body: ""}, nil
	}
	return QueryResult{
		Title:    fmt.Sprintf("%d hit(s) for %q in %s", len(hits), query, a.name),
		SourceID: a.name,
		URI:      "file://" + a.path,
		Body:     strings.Join(hits, "\n\n---\n\n"),
	}, nil
}

// -- url adapter -------------------------------------------------------

// urlAdapter is a generic HTTPS fetch backend. The query string is
// appended to the configured base URL; the response body is returned
// verbatim. Useful for REST endpoints that return markdown or plain
// text directly.
type urlAdapter struct {
	name string
	base string
}

func newURLAdapter(s SourceConfig) Adapter {
	b, _ := s.Settings["base"].(string)
	return &urlAdapter{name: s.Name, base: b}
}

func (a *urlAdapter) Name() string { return a.name }
func (a *urlAdapter) Type() string { return "url" }

func (a *urlAdapter) Health(ctx context.Context) Health {
	if a.base == "" {
		return Health{Status: "config-missing", Message: "no `base` URL set"}
	}
	if _, err := url.Parse(a.base); err != nil {
		return Health{Status: "config-invalid", Message: err.Error()}
	}
	return Health{OK: true, Status: "ready"}
}

func (a *urlAdapter) Query(ctx context.Context, query string) (QueryResult, error) {
	if h := a.Health(ctx); !h.OK {
		return QueryResult{}, fmt.Errorf("%s: %s", a.name, h.Message)
	}
	u, err := url.Parse(a.base)
	if err != nil {
		return QueryResult{}, fmt.Errorf("parse base url: %w", err)
	}
	q := u.Query()
	q.Set("q", query)
	u.RawQuery = q.Encode()

	httpCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(httpCtx, http.MethodGet, u.String(), nil)
	if err != nil {
		return QueryResult{}, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return QueryResult{}, fmt.Errorf("fetch %s: %w", u.String(), err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1 MiB cap
	if err != nil {
		return QueryResult{}, err
	}
	if resp.StatusCode >= 400 {
		return QueryResult{}, fmt.Errorf("fetch %s: status %d: %s", u.String(), resp.StatusCode, string(body))
	}
	return QueryResult{
		Title:    fmt.Sprintf("response from %s", a.name),
		SourceID: a.name,
		URI:      u.String(),
		Body:     string(body),
	}, nil
}

// -- unknown adapter ---------------------------------------------------

// unknownAdapter is what we return for a context.json entry whose
// `type:` doesn't match any registered backend. It surfaces a clear
// error on query and an unreachable health.
type unknownAdapter struct {
	name string
	kind string
}

func (a *unknownAdapter) Name() string { return a.name }
func (a *unknownAdapter) Type() string { return a.kind }
func (a *unknownAdapter) Health(ctx context.Context) Health {
	return Health{Status: "unknown-type", Message: fmt.Sprintf("adapter type %q not registered", a.kind)}
}
func (a *unknownAdapter) Query(ctx context.Context, query string) (QueryResult, error) {
	return QueryResult{}, fmt.Errorf("adapter type %q not registered (source %q)", a.kind, a.name)
}
