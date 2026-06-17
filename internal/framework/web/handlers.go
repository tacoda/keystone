package web

import (
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

// render is the shared template entry. Parses a fresh template set
// per call so per-page `{{define "main"}}` blocks don't collide
// across the embedded set. The page name is the .html filename
// (e.g. "home.html"); partials in the same dir are pulled in too.
func (s *server) render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Fragment templates (filenames starting with `_`) don't need the
	// layout — they render raw.
	pageOnly := strings.HasPrefix(name, "_")

	t := template.New("").Funcs(s.funcs)
	files := []string{name}
	if !pageOnly {
		files = append(files, "layout.html")
	}
	// Also pull in any same-page partials referenced by the page.
	// We do this by parsing every `_*.html` file the page might
	// invoke — cheap given the small embedded set.
	if entries, err := fs.ReadDir(s.tmplFS, "."); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasPrefix(e.Name(), "_") && e.Name() != name {
				files = append(files, e.Name())
			}
		}
	}
	t, err := t.ParseFS(s.tmplFS, files...)
	if err != nil {
		http.Error(w, "render parse: "+err.Error(), http.StatusInternalServerError)
		return
	}

	root := "layout"
	if pageOnly {
		root = name
	}
	if err := t.ExecuteTemplate(w, root, data); err != nil {
		http.Error(w, "render exec: "+err.Error(), http.StatusInternalServerError)
	}
}

// -- pages -----------------------------------------------------------

func (s *server) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	idx, err := s.buildIndex()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sources, _ := sourceList(r.Context(), s.projectDir)
	s.render(w, "home.html", map[string]any{
		"ProjectDir":     s.projectDir,
		"PrimitiveCount": len(idx.Primitives),
		"ByKind":         idx.ByKind,
		"Generated":      idx.Generated,
		"Sources":        sources,
	})
}

func (s *server) handlePrimitivesList(w http.ResponseWriter, r *http.Request) {
	primitives, err := s.loadPrimitives()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	kind := r.URL.Query().Get("kind")
	glob := r.URL.Query().Get("glob")
	filtered := filterPrimitives(primitives, kind, glob)
	s.render(w, "primitives.html", map[string]any{
		"ProjectDir": s.projectDir,
		"Primitives": filtered,
		"Total":      len(primitives),
		"Filter":     map[string]string{"kind": kind, "glob": glob},
		"Kinds":      uniqueKinds(primitives),
	})
}

// /primitives/<kind>/<id...>
func (s *server) handlePrimitivesDetail(w http.ResponseWriter, r *http.Request) {
	kind, id := splitPrimitivePath(r.URL.Path, "/primitives/")
	if kind == "" || id == "" {
		http.NotFound(w, r)
		return
	}
	primitives, err := s.loadPrimitives()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, p := range primitives {
		if p.Kind == kind && p.ID == id {
			body, err := os.ReadFile(filepath.Join(s.projectDir, p.Path))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			incoming := primitive.IncomingRefs(primitives, p)
			s.render(w, "primitive_detail.html", map[string]any{
				"ProjectDir": s.projectDir,
				"Primitive":  p,
				"Body":       string(body),
				"Incoming":   incoming,
			})
			return
		}
	}
	http.NotFound(w, r)
}

func (s *server) handleSources(w http.ResponseWriter, r *http.Request) {
	entries, err := sourceList(r.Context(), s.projectDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "sources.html", map[string]any{
		"ProjectDir": s.projectDir,
		"Sources":    entries,
	})
}

func (s *server) handleSourceDetail(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/sources/")
	if name == "" {
		http.NotFound(w, r)
		return
	}
	entries, err := sourceList(r.Context(), s.projectDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, e := range entries {
		if e.Name == name {
			s.render(w, "source_detail.html", map[string]any{
				"ProjectDir": s.projectDir,
				"Source":     e,
			})
			return
		}
	}
	http.NotFound(w, r)
}

// -- HTMX fragments ---------------------------------------------------

// handlePrimitivesFragment returns just the table-body partial.
// Wired to the kind/glob filter form via hx-get on input change.
func (s *server) handlePrimitivesFragment(w http.ResponseWriter, r *http.Request) {
	primitives, err := s.loadPrimitives()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	kind := r.URL.Query().Get("kind")
	glob := r.URL.Query().Get("glob")
	filtered := filterPrimitives(primitives, kind, glob)
	s.render(w, "_primitives_table.html", map[string]any{
		"Primitives": filtered,
	})
}

// -- helpers ----------------------------------------------------------

func filterPrimitives(in []primitive.Primitive, kind, glob string) []primitive.Primitive {
	out := in[:0:0]
	for _, p := range in {
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
	return out
}

func uniqueKinds(in []primitive.Primitive) []string {
	seen := map[string]struct{}{}
	out := []string{}
	for _, p := range in {
		if _, ok := seen[p.Kind]; ok {
			continue
		}
		seen[p.Kind] = struct{}{}
		out = append(out, p.Kind)
	}
	return out
}
