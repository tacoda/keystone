package web

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// runKeystone shells out to the running keystone binary the same way
// the MCP write tools do. Reuses CLI arg parsing; no duplicated
// authoring logic. Returns combined stdout+stderr.
func runKeystone(ctx context.Context, projectDir string, args ...string) (string, error) {
	self, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locate self: %w", err)
	}
	fullArgs := append([]string{}, args...)
	cmd := exec.CommandContext(ctx, self, fullArgs...)
	cmd.Dir = projectDir
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err = cmd.Run()
	return buf.String(), err
}

// htmxRedirect writes an HX-Redirect header so the dashboard
// performs a full page navigation after a form post. Falls back to
// 303-See-Other for non-HTMX requests.
func htmxRedirect(w http.ResponseWriter, r *http.Request, url string) {
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", url)
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, url, http.StatusSeeOther)
}

// htmxFragment writes a small HTML fragment as the response body —
// used for inline form feedback ("✓ created", "✗ error: ...").
func htmxFragment(w http.ResponseWriter, html string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(html))
}

// -- new primitive --------------------------------------------------

// handleActionNewPrimitive accepts a POST from the "scaffold new
// primitive" form. Body fields: kind, id, optional kind-specific
// extras. Shells out to `keystone new <kind> <id>` and redirects to
// the new primitive's detail page on success.
func (s *server) handleActionNewPrimitive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		htmxFragment(w, errorFragment(err.Error()))
		return
	}
	kind := strings.TrimSpace(r.FormValue("kind"))
	id := strings.TrimSpace(r.FormValue("id"))
	if kind == "" || id == "" {
		htmxFragment(w, errorFragment("kind and id are required"))
		return
	}

	args := []string{"new", kind, id}
	if sensorKind := strings.TrimSpace(r.FormValue("sensor_kind")); kind == "sensor" && sensorKind != "" {
		args = append(args, "--kind", sensorKind)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	out, err := runKeystone(ctx, s.projectDir, args...)
	if err != nil {
		htmxFragment(w, errorFragment(fmt.Sprintf("keystone new failed: %v\n%s", err, out)))
		return
	}

	// Refresh INDEX.json so the new primitive shows up.
	_, _ = runKeystone(ctx, s.projectDir, "index", "--dir", s.projectDir)

	htmxRedirect(w, r, "/primitives/"+kind+"/"+id)
}

// -- delete primitive ----------------------------------------------

// handleActionDeletePrimitive removes the underlying file for a
// primitive and refreshes the index. The dashboard fires this from
// a small "delete" button on the primitive detail page.
//
// 2.0 keeps the user as the source of truth: the file is unlinked
// (no recursive prune of empty dirs). Re-running `keystone index`
// removes the entry from INDEX.json on the next walk.
func (s *server) handleActionDeletePrimitive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		htmxFragment(w, errorFragment(err.Error()))
		return
	}
	kind := r.FormValue("kind")
	id := r.FormValue("id")
	if kind == "" || id == "" {
		htmxFragment(w, errorFragment("kind and id are required"))
		return
	}
	primitives, err := s.loadPrimitives()
	if err != nil {
		htmxFragment(w, errorFragment(err.Error()))
		return
	}
	var target string
	for _, p := range primitives {
		if p.Kind == kind && p.ID == id {
			target = filepath.Join(s.projectDir, p.Path)
			break
		}
	}
	if target == "" {
		htmxFragment(w, errorFragment("no primitive with kind="+kind+" id="+id))
		return
	}
	if err := os.Remove(target); err != nil {
		htmxFragment(w, errorFragment(err.Error()))
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	_, _ = runKeystone(ctx, s.projectDir, "index", "--dir", s.projectDir)
	htmxRedirect(w, r, "/primitives")
}

// -- policy actions -------------------------------------------------

// handleActionPolicyAdd shells out to `keystone policy add` with the
// shorthand+version the form supplied. On success, redirects to
// /policies.
func (s *server) handleActionPolicyAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		htmxFragment(w, errorFragment(err.Error()))
		return
	}
	spec := strings.TrimSpace(r.FormValue("spec"))
	if spec == "" {
		htmxFragment(w, errorFragment("spec is required (e.g. acme/policies@v1.0.0)"))
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	args := []string{"policy", "add", spec}
	if name != "" {
		args = append(args, "--name", name)
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	out, err := runKeystone(ctx, s.projectDir, args...)
	if err != nil {
		htmxFragment(w, errorFragment(fmt.Sprintf("policy add failed: %v\n%s", err, out)))
		return
	}
	htmxRedirect(w, r, "/policies")
}

func (s *server) handleActionPolicyRemove(w http.ResponseWriter, r *http.Request) {
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
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	out, err := runKeystone(ctx, s.projectDir, "policy", "remove", name)
	if err != nil {
		htmxFragment(w, errorFragment(fmt.Sprintf("policy remove failed: %v\n%s", err, out)))
		return
	}
	htmxRedirect(w, r, "/policies")
}

// -- verify ---------------------------------------------------------

// handleActionVerify runs `keystone verify` and renders the captured
// output as an HTMX fragment swapped into the verify page's result
// pane.
func (s *server) handleActionVerify(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	out, err := runKeystone(ctx, s.projectDir, "verify", "--dir", s.projectDir)
	pass := err == nil
	s.render(w, "_verify_result.html", map[string]any{
		"Pass":   pass,
		"Output": out,
		"Ran":    time.Now().Format(time.RFC3339),
	})
}

// errorFragment renders a small HTML error banner.
func errorFragment(msg string) string {
	return `<div class="bad">✗ ` + htmlEscape(msg) + `</div>`
}

func htmlEscape(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&#39;",
	)
	return r.Replace(s)
}
