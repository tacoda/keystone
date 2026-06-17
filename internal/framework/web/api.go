package web

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/tacoda/keystone/internal/framework/mcp"
)

// writeJSON serializes v as indented JSON and writes it with
// Content-Type: application/json. Errors during marshal/write surface
// as a 500 — these only happen on I/O failure.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{"error": msg})
}

// -- /api/index --------------------------------------------------------

func (s *server) apiIndex(w http.ResponseWriter, r *http.Request) {
	idx, err := s.buildIndex()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, idx)
}

// -- /api/primitives ---------------------------------------------------

func (s *server) apiPrimitives(w http.ResponseWriter, r *http.Request) {
	primitives, err := s.loadPrimitives()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	kind := r.URL.Query().Get("kind")
	glob := r.URL.Query().Get("glob")
	out := primitives[:0:0]
	for _, p := range primitives {
		if kind != "" && p.Kind != kind {
			continue
		}
		if glob != "" {
			match := false
			for _, g := range p.Globs {
				if g == glob {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}
		out = append(out, p)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"count":      len(out),
		"primitives": out,
	})
}

// /api/primitives/{kind}/{id...}
func (s *server) apiPrimitiveDetail(w http.ResponseWriter, r *http.Request) {
	kind, id := splitPrimitivePath(r.URL.Path, "/api/primitives/")
	if kind == "" || id == "" {
		writeError(w, http.StatusBadRequest, "URL must be /api/primitives/<kind>/<id>")
		return
	}
	primitives, err := s.loadPrimitives()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for _, p := range primitives {
		if p.Kind == kind && p.ID == id {
			body, err := os.ReadFile(filepath.Join(s.projectDir, p.Path))
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"kind":        p.Kind,
				"id":          p.ID,
				"path":        p.Path,
				"description": p.Description,
				"body":        string(body),
			})
			return
		}
	}
	writeError(w, http.StatusNotFound, "no primitive with kind="+kind+" id="+id)
}

// -- /api/sources ------------------------------------------------------

func (s *server) apiSources(w http.ResponseWriter, r *http.Request) {
	entries, err := sourceList(r.Context(), s.projectDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"count":   len(entries),
		"sources": entries,
	})
}

// /api/sources/{name}/health
func (s *server) apiSourceDetail(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/api/sources/")
	rest = strings.TrimSuffix(rest, "/health")
	rest = strings.TrimSuffix(rest, "/")
	if rest == "" {
		writeError(w, http.StatusBadRequest, "URL must be /api/sources/<name>/health")
		return
	}
	entries, err := sourceList(r.Context(), s.projectDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for _, e := range entries {
		if e.Name == rest {
			writeJSON(w, http.StatusOK, e)
			return
		}
	}
	writeError(w, http.StatusNotFound, "no source named "+rest)
}

// -- /api/harness/status -----------------------------------------------

func (s *server) apiHarnessStatus(w http.ResponseWriter, r *http.Request) {
	idx, err := s.buildIndex()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"version":         idx.Version,
		"generated":       idx.Generated,
		"primitive_count": len(idx.Primitives),
		"by_kind":         idx.ByKind,
		"project_dir":     s.projectDir,
	})
}

// -- helpers -----------------------------------------------------------

// splitPrimitivePath pulls (kind, id) out of `<prefix><kind>/<id...>`.
// ids may contain slashes (guides use <topic>/<name>), so id is the
// remainder after the first slash following the kind.
func splitPrimitivePath(urlPath, prefix string) (kind, id string) {
	rest := strings.TrimPrefix(urlPath, prefix)
	slash := strings.IndexByte(rest, '/')
	if slash < 0 {
		return rest, ""
	}
	return rest[:slash], rest[slash+1:]
}

// sourceEntry is the small shape the API + dashboard share for a
// single external source.
type sourceEntry struct {
	Name   string     `json:"name"`
	Type   string     `json:"type"`
	Health mcp.Health `json:"health"`
}

// sourceList reads .keystone/context.json and probes each adapter's
// health. Returns an empty slice when no config file exists.
func sourceList(ctx context.Context, projectDir string) ([]sourceEntry, error) {
	cfg, err := mcp.LoadContextConfig(projectDir)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return []sourceEntry{}, nil
	}
	out := make([]sourceEntry, 0, len(cfg.Sources))
	for _, src := range cfg.Sources {
		a := mcp.BuildAdapter(src)
		out = append(out, sourceEntry{
			Name:   a.Name(),
			Type:   a.Type(),
			Health: a.Health(ctx),
		})
	}
	return out, nil
}
