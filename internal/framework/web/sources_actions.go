package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tacoda/keystone/internal/framework/mcp"
)

// handleActionSourceAdd appends a new entry to .keystone/context.json.
// Form fields: name, type, plus a single free-form `settings` text
// area carrying JSON for the adapter-specific keys (path, base, etc).
// The settings JSON is merged into the new source object.
//
// Re-adding a name that already exists overwrites the prior entry —
// makes the dashboard's "add or update" UX a single form.
func (s *server) handleActionSourceAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		htmxFragment(w, errorFragment(err.Error()))
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	srcType := strings.TrimSpace(r.FormValue("type"))
	settingsRaw := strings.TrimSpace(r.FormValue("settings"))
	if name == "" || srcType == "" {
		htmxFragment(w, errorFragment("name and type are required"))
		return
	}

	settings := map[string]any{}
	if settingsRaw != "" {
		if err := json.Unmarshal([]byte(settingsRaw), &settings); err != nil {
			htmxFragment(w, errorFragment("settings JSON: "+err.Error()))
			return
		}
	}

	if err := writeSourceEntry(s.projectDir, name, srcType, settings); err != nil {
		htmxFragment(w, errorFragment(err.Error()))
		return
	}
	// Mutation invalidates the cached health snapshot — refresh now
	// in the background so the redirect target sees fresh data
	// without re-blocking the handler.
	go s.healthCache.refresh(context.Background())
	htmxRedirect(w, r, "/sources")
}

// handleActionSourceRemove deletes the named entry from
// .keystone/context.json. Idempotent — missing entries return OK.
func (s *server) handleActionSourceRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		htmxFragment(w, errorFragment(err.Error()))
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		htmxFragment(w, errorFragment("name is required"))
		return
	}
	if err := removeSourceEntry(s.projectDir, name); err != nil {
		htmxFragment(w, errorFragment(err.Error()))
		return
	}
	go s.healthCache.refresh(context.Background())
	htmxRedirect(w, r, "/sources")
}

// handleActionSourceQuery runs a query against the named source and
// renders the result as an inline fragment swapped under the form.
func (s *server) handleActionSourceQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		htmxFragment(w, errorFragment(err.Error()))
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	q := strings.TrimSpace(r.FormValue("query"))
	if name == "" || q == "" {
		htmxFragment(w, errorFragment("name and query are required"))
		return
	}
	cfg, err := mcp.LoadContextConfig(s.projectDir)
	if err != nil {
		htmxFragment(w, errorFragment(err.Error()))
		return
	}
	if cfg == nil {
		htmxFragment(w, errorFragment("no sources configured"))
		return
	}
	var entry *mcp.SourceConfig
	for i := range cfg.Sources {
		if cfg.Sources[i].Name == name {
			entry = &cfg.Sources[i]
			break
		}
	}
	if entry == nil {
		htmxFragment(w, errorFragment("no source named "+name))
		return
	}
	a := mcp.BuildAdapter(*entry)
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	res, err := a.Query(ctx, q)
	if err != nil {
		htmxFragment(w, errorFragment(err.Error()))
		return
	}
	s.render(w, "_source_query_result.html", map[string]any{
		"Result": res,
		"Source": name,
		"Query":  q,
	})
}

// handleActionSourceHealth re-probes one source's health and renders
// the updated badge as a fragment so the source detail page swaps it
// in place.
func (s *server) handleActionSourceHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		htmxFragment(w, errorFragment(err.Error()))
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		htmxFragment(w, errorFragment("name is required"))
		return
	}
	cfg, err := mcp.LoadContextConfig(s.projectDir)
	if err != nil || cfg == nil {
		htmxFragment(w, errorFragment("no sources configured"))
		return
	}
	for _, src := range cfg.Sources {
		if src.Name == name {
			a := mcp.BuildAdapter(src)
			ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
			defer cancel()
			h := a.Health(ctx)
			s.render(w, "_health_badge.html", map[string]any{
				"Health": h,
				"Probed": time.Now().Format(time.RFC3339),
			})
			return
		}
	}
	htmxFragment(w, errorFragment("no source named "+name))
}

// handleActionSourceVerifyAll probes every configured source's
// health and renders an aggregate table the dashboard swaps in.
// "Verify integrations" button.
func (s *server) handleActionSourceVerifyAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	entries, err := s.sourceList(r.Context())
	if err != nil {
		htmxFragment(w, errorFragment(err.Error()))
		return
	}
	s.render(w, "_sources_table.html", map[string]any{
		"Sources": entries,
		"Probed":  time.Now().Format(time.RFC3339),
	})
}

// -- context.json read/write helpers --------------------------------

// writeSourceEntry adds or replaces a single source in
// .keystone/context.json. Creates the file when missing.
func writeSourceEntry(projectDir, name, srcType string, settings map[string]any) error {
	doc, err := readContextDoc(projectDir)
	if err != nil {
		return err
	}
	srcArray := toSourceArray(doc["sources"])
	// Build the new entry preserving any extra settings keys.
	entry := map[string]any{"name": name, "type": srcType}
	for k, v := range settings {
		if k == "name" || k == "type" {
			continue
		}
		entry[k] = v
	}
	// Replace by name, or append.
	replaced := false
	for i, s := range srcArray {
		if asString(s, "name") == name {
			srcArray[i] = entry
			replaced = true
			break
		}
	}
	if !replaced {
		srcArray = append(srcArray, entry)
	}
	doc["sources"] = srcArray
	if _, ok := doc["version"]; !ok {
		doc["version"] = 2
	}
	return writeContextDoc(projectDir, doc)
}

func removeSourceEntry(projectDir, name string) error {
	doc, err := readContextDoc(projectDir)
	if err != nil {
		return err
	}
	srcArray := toSourceArray(doc["sources"])
	filtered := srcArray[:0]
	for _, s := range srcArray {
		if asString(s, "name") == name {
			continue
		}
		filtered = append(filtered, s)
	}
	doc["sources"] = filtered
	return writeContextDoc(projectDir, doc)
}

func readContextDoc(projectDir string) (map[string]any, error) {
	path := filepath.Join(projectDir, ".keystone", "context.json")
	doc := map[string]any{"version": 2, "sources": []any{}}
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return doc, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if doc["sources"] == nil {
		doc["sources"] = []any{}
	}
	return doc, nil
}

func writeContextDoc(projectDir string, doc map[string]any) error {
	dir := filepath.Join(projectDir, ".keystone")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, "context.json")
	body, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	return os.WriteFile(path, body, 0o644)
}

func toSourceArray(v any) []any {
	if v == nil {
		return []any{}
	}
	arr, ok := v.([]any)
	if !ok {
		return []any{}
	}
	return arr
}

func asString(v any, key string) string {
	m, ok := v.(map[string]any)
	if !ok {
		return ""
	}
	s, _ := m[key].(string)
	return s
}
