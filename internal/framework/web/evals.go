package web

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/tacoda/keystone/internal/framework/eval"
)

// handleEvals lists every eval discovered in the charter + offers a
// "run all" button.
func (s *server) handleEvals(w http.ResponseWriter, r *http.Request) {
	specs, err := eval.LoadAll(s.projectDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.renderPage(w, r, "evals.html", map[string]any{
		"ProjectDir": s.projectDir,
		"Evals":      specs,
	})
}

// handleActionEvalRun executes every eval (or a filtered subset),
// renders the markdown report as an inline fragment.
func (s *server) handleActionEvalRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		htmxFragment(w, errorFragment(err.Error()))
		return
	}
	filter := strings.TrimSpace(r.FormValue("filter"))
	specs, err := eval.LoadAll(s.projectDir)
	if err != nil {
		htmxFragment(w, errorFragment(err.Error()))
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()
	rep := eval.Run(ctx, s.projectDir, specs, filter)
	s.render(w, "_eval_report.html", map[string]any{
		"Report": rep,
	})
}

// apiEvals — JSON endpoint mirroring keystone_eval_list.
func (s *server) apiEvals(w http.ResponseWriter, r *http.Request) {
	specs, err := eval.LoadAll(s.projectDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	type entry struct {
		ID          string   `json:"id"`
		Level       string   `json:"level"`
		Levels      []string `json:"levels,omitempty"`
		Description string   `json:"description"`
	}
	out := make([]entry, 0, len(specs))
	for _, sp := range specs {
		out = append(out, entry{ID: sp.ID, Level: sp.Level, Levels: sp.Levels, Description: sp.Description})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"count": len(out),
		"evals": out,
	})
}

// apiEvalRun — JSON endpoint mirroring keystone_eval_run.
func (s *server) apiEvalRun(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter")
	specs, err := eval.LoadAll(s.projectDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()
	rep := eval.Run(ctx, s.projectDir, specs, filter)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(rep)
}
