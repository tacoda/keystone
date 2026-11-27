package web

import (
	"net/http"
	"strings"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/drift"
)

// driftReport computes the context-drift report for the project, the
// same data `keystone drift --json` prints. files= scopes it.
func (s *server) driftReport(r *http.Request) (drift.Report, error) {
	days := config.DefaultDriftDays
	if cfg, err := config.ReadProjectConfig(s.projectDir); err == nil {
		days = cfg.DriftWindow()
	}
	var files []string
	if f := r.URL.Query().Get("files"); f != "" {
		for _, p := range strings.Split(f, ",") {
			if p = strings.TrimSpace(p); p != "" {
				files = append(files, p)
			}
		}
	}
	return drift.Scan(s.projectDir, drift.Options{Files: files, Days: days})
}

// apiDrift serves /api/drift — the drift report as JSON (same data the
// CLI and the MCP keystone_drift_report return).
func (s *server) apiDrift(w http.ResponseWriter, r *http.Request) {
	rep, err := s.driftReport(r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rep)
}

// handleDrift renders the drift pane on the flywheels page section.
func (s *server) handleDrift(w http.ResponseWriter, r *http.Request) {
	rep, err := s.driftReport(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	counts := map[string]int{}
	for _, wn := range rep.Warnings {
		counts[wn.Kind]++
	}
	s.render(w, "drift.html", map[string]any{
		"ProjectDir": s.projectDir,
		"Checked":    rep.Checked,
		"Warnings":   rep.Warnings,
		"Counts":     counts,
	})
}
