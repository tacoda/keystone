package web

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	kconfig "github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// MetricsSnapshot is the per-request view of harness health. Cheap
// to compute — walks the harness once, no caching.
type MetricsSnapshot struct {
	ProjectDir     string             `json:"project_dir"`
	Generated      string             `json:"generated"`
	PrimitiveTotal int                `json:"primitive_total"`
	ByKind         map[string]int     `json:"by_kind"`
	LintErrors     int                `json:"lint_errors"`
	LintWarnings   int                `json:"lint_warnings"`
	Freshness      []FreshnessEntry   `json:"freshness"`
	SourceHealths  []sourceEntry      `json:"source_healths"`
	IndexFresh     bool               `json:"index_fresh"`
	IndexFile      string             `json:"index_file"`
	IndexAge       string             `json:"index_age,omitempty"`
}

// FreshnessEntry reports the most-recently-modified primitive per
// kind. Useful for spotting work-in-progress areas of the harness
// at a glance.
type FreshnessEntry struct {
	Kind     string `json:"kind"`
	ID       string `json:"id"`
	Path     string `json:"path"`
	Modified string `json:"modified"`
}

// collectMetrics is the read-only snapshot the /metrics page + the
// /api/metrics endpoint share.
func (s *server) collectMetrics(ctx context.Context) (*MetricsSnapshot, error) {
	primitives, err := s.loadPrimitives()
	if err != nil {
		return nil, err
	}

	byKind := map[string]int{}
	latestByKind := map[string]FreshnessEntry{}
	for _, p := range primitives {
		byKind[p.Kind]++
		info, statErr := os.Stat(filepath.Join(s.projectDir, p.Path))
		if statErr != nil {
			continue
		}
		modified := info.ModTime()
		prev, ok := latestByKind[p.Kind]
		if !ok {
			latestByKind[p.Kind] = FreshnessEntry{
				Kind: p.Kind, ID: p.ID, Path: p.Path,
				Modified: modified.UTC().Format(time.RFC3339),
			}
			continue
		}
		prevTime, _ := time.Parse(time.RFC3339, prev.Modified)
		if modified.After(prevTime) {
			latestByKind[p.Kind] = FreshnessEntry{
				Kind: p.Kind, ID: p.ID, Path: p.Path,
				Modified: modified.UTC().Format(time.RFC3339),
			}
		}
	}

	freshness := make([]FreshnessEntry, 0, len(latestByKind))
	for _, e := range latestByKind {
		freshness = append(freshness, e)
	}
	sort.Slice(freshness, func(i, j int) bool {
		return freshness[i].Modified > freshness[j].Modified
	})

	findings := primitive.Lint(primitives)
	errs, warns := 0, 0
	for _, f := range findings {
		switch f.Severity {
		case primitive.FindingError:
			errs++
		case primitive.FindingWarning:
			warns++
		}
	}

	indexPath := filepath.Join(s.projectDir, kconfig.KeystoneDir(kconfig.DefaultHarnessRoot), kconfig.IndexName)
	indexFresh := false
	indexAge := ""
	if info, err := os.Stat(indexPath); err == nil {
		age := time.Since(info.ModTime())
		indexAge = humanizeDuration(age)
		// "fresh" = modified within the last 5 minutes AND newer than
		// every primitive file.
		indexFresh = age < 5*time.Minute
		for _, p := range primitives {
			if pi, err := os.Stat(filepath.Join(s.projectDir, p.Path)); err == nil {
				if pi.ModTime().After(info.ModTime()) {
					indexFresh = false
					break
				}
			}
		}
	}

	sources, _ := s.sourceList(ctx)

	return &MetricsSnapshot{
		ProjectDir:     s.projectDir,
		Generated:      time.Now().UTC().Format(time.RFC3339),
		PrimitiveTotal: len(primitives),
		ByKind:         byKind,
		LintErrors:     errs,
		LintWarnings:   warns,
		Freshness:      freshness,
		SourceHealths:  sources,
		IndexFresh:     indexFresh,
		IndexFile:      indexPath,
		IndexAge:       indexAge,
	}, nil
}

// /metrics HTMX page.
func (s *server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	snap, err := s.collectMetrics(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.renderPage(w, r, "metrics.html", map[string]any{
		"ProjectDir": s.projectDir,
		"Metrics":    snap,
		"KPINames":   kpiNames(),
	})
}

// /api/metrics JSON endpoint. Same shape as the dashboard view.
func (s *server) apiMetrics(w http.ResponseWriter, r *http.Request) {
	snap, err := s.collectMetrics(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, snap)
}

// humanizeDuration renders a short, human-friendly span.
func humanizeDuration(d time.Duration) string {
	switch {
	case d < time.Minute:
		return strings.TrimSuffix(strings.TrimSuffix((d / time.Second).String()+"s", "0s"), "s") + "s"
	case d < time.Hour:
		return (d / time.Minute).String() + "m"
	case d < 24*time.Hour:
		return (d / time.Hour).String() + "h"
	default:
		return (d / (24 * time.Hour)).String() + "d"
	}
}

// JSON helper for tests / debugging.
func (s *server) metricsJSON(ctx context.Context) ([]byte, error) {
	snap, err := s.collectMetrics(ctx)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(snap, "", "  ")
}
