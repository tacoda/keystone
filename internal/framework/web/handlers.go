package web

import (
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"
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

// renderPage renders a full page template. When the request is an
// HTMX swap (`HX-Request: true`), only the page's `main` block is
// returned so the SPA shell stays put. Otherwise the full layout
// is rendered so deep-links and reloads still produce a complete
// document.
//
// History-restore swaps (HX-History-Restore-Request) get the
// fragment too — htmx replaces the swap target, not the whole doc.
func (s *server) renderPage(w http.ResponseWriter, r *http.Request, name string, data any) {
	if r != nil && r.Header.Get("HX-Request") == "true" {
		s.renderFragment(w, name, data)
		return
	}
	s.render(w, name, data)
}

// renderFragment parses the page (without layout.html) and executes
// its `main` block directly — the body that would normally be
// injected into the layout's `{{block "main"}}` slot. Used for HTMX
// in-app navigation.
func (s *server) renderFragment(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	t := template.New("").Funcs(s.funcs)
	files := []string{name}
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
	if err := t.ExecuteTemplate(w, "main", data); err != nil {
		http.Error(w, "render exec: "+err.Error(), http.StatusInternalServerError)
	}
}

// -- pages -----------------------------------------------------------

func (s *server) handlePrimitivesList(w http.ResponseWriter, r *http.Request) {
	primitives, err := s.loadPrimitives()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	kind := r.URL.Query().Get("kind")
	glob := r.URL.Query().Get("glob")
	tags := r.URL.Query()["tag"] // multi-value query param: ?tag=a&tag=b
	filtered := filterPrimitives(primitives, kind, glob, tags)
	s.renderPage(w, r, "primitives.html", map[string]any{
		"ProjectDir": s.projectDir,
		"Primitives": filtered,
		"Total":      len(primitives),
		"Filter":     map[string]any{"kind": kind, "glob": glob, "tags": tags},
		"Kinds":      uniqueKinds(primitives),
		"AllTags":    uniqueTags(primitives),
	})
}

// /primitives/<kind>/<id...>
func (s *server) handlePrimitivesDetail(w http.ResponseWriter, r *http.Request) {
	kind, id := splitPrimitivePath(r.URL.Path, "/harness/primitives/")
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
			s.renderPage(w, r, "primitive_detail.html", map[string]any{
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
	entries, err := s.sourceList(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.renderPage(w, r, "sources.html", map[string]any{
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
	entries, err := s.sourceList(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, e := range entries {
		if e.Name == name {
			s.renderPage(w, r, "source_detail.html", map[string]any{
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
	tags := r.URL.Query()["tag"]
	filtered := filterPrimitives(primitives, kind, glob, tags)
	s.render(w, "_primitives_table.html", map[string]any{
		"Primitives": filtered,
	})
}

// -- helpers ----------------------------------------------------------

func filterPrimitives(in []primitive.Primitive, kind, glob string, tags []string) []primitive.Primitive {
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
		if len(tags) > 0 && !hasAllTagsWeb(p.Tags, tags) {
			continue
		}
		out = append(out, p)
	}
	return out
}

// hasAllTagsWeb mirrors the AND semantics used by `keystone list
// --tag X --tag Y` — every requested tag must appear on the primitive
// (concern contributions already merged by Compose).
func hasAllTagsWeb(have, want []string) bool {
	if len(want) == 0 {
		return true
	}
	set := make(map[string]struct{}, len(have))
	for _, h := range have {
		set[h] = struct{}{}
	}
	for _, w := range want {
		if _, ok := set[w]; !ok {
			return false
		}
	}
	return true
}

// uniqueTags returns the sorted union of every primitive's Tags. Used
// by the dashboard's tag-filter chip strip.
func uniqueTags(in []primitive.Primitive) []string {
	seen := map[string]struct{}{}
	for _, p := range in {
		for _, t := range p.Tags {
			seen[t] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for t := range seen {
		out = append(out, t)
	}
	sort.Strings(out)
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
