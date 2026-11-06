package web

import (
	"net/http"
	"sort"
	"strings"

	"github.com/tacoda/keystone/internal/framework/budget"
	"github.com/tacoda/keystone/internal/framework/charter"
	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// region is one uncharted top-level directory with its file count.
type region struct {
	Name  string
	Count int
}

// handleBudget renders the context-budget page: each host's always-loaded
// tier, the on-demand tier per kind, and any keystone.json cap exceeded.
func (s *server) handleBudget(w http.ResponseWriter, r *http.Request) {
	rep, err := budget.MeasureProject(s.projectDir)
	if err != nil {
		s.renderPage(w, r, "budget.html", map[string]any{"Error": err.Error()})
		return
	}
	var kinds []region
	for _, k := range primitive.KnownKinds {
		if t := rep.ByKind[string(k)]; t > 0 {
			kinds = append(kinds, region{Name: string(k), Count: t})
		}
	}
	largest := rep.OnDemand
	if len(largest) > 10 {
		largest = largest[:10]
	}
	s.renderPage(w, r, "budget.html", map[string]any{
		"Report": rep, "Kinds": kinds, "Largest": largest, "AlwaysCap": budget.AlwaysCap(s.projectDir),
	})
}

// handleCoverage renders the charter-coverage page: how much of the
// project's source a guide governs, and which regions are uncharted.
func (s *server) handleCoverage(w http.ResponseWriter, r *http.Request) {
	res, err := charter.Coverage(s.projectDir, config.DefaultAgentsRoot)
	if err != nil {
		s.renderPage(w, r, "coverage.html", map[string]any{"Error": err.Error()})
		return
	}
	pct := 100
	if res.Total > 0 {
		pct = res.Governed * 100 / res.Total
	}
	counts := res.UnchartedByRegion()
	regions := make([]region, 0, len(counts))
	for name, c := range counts {
		regions = append(regions, region{Name: name, Count: c})
	}
	sort.Slice(regions, func(i, j int) bool {
		if regions[i].Count != regions[j].Count {
			return regions[i].Count > regions[j].Count
		}
		return regions[i].Name < regions[j].Name
	})
	s.renderPage(w, r, "coverage.html", map[string]any{
		"Total":     res.Total,
		"Governed":  res.Governed,
		"Uncharted": len(res.Uncharted),
		"Pct":       pct,
		"Regions":   regions,
	})
}

// signalRow is one signal with the primitives that subscribe to it.
type signalRow struct {
	Name        string
	Builtin     bool
	Subscribers []string
}

// handleSignals renders the signals page: every known signal (built-in +
// project-declared) with the primitives that subscribe via `on:`, plus
// the host phases (bridged, not signals).
func (s *server) handleSignals(w http.ResponseWriter, r *http.Request) {
	var custom []string
	if cfg, err := config.ReadProjectConfig(s.projectDir); err == nil && cfg != nil {
		custom = cfg.Signals
	}
	subs := signalSubscribers(s.projectDir)
	rows := buildSignalRows(charter.Signals(custom), subs)
	s.renderPage(w, r, "signals.html", map[string]any{
		"Signals":    rows,
		"HostPhases": primitive.HostPhases,
	})
}

// signalSubscribers maps each signal name to the primitive ids that
// subscribe to it via `on:` (host phases excluded — those bridge).
func signalSubscribers(projectDir string) map[string][]string {
	subs := map[string][]string{}
	prims, _, err := primitive.Walk(projectDir, config.DefaultAgentsRoot)
	if err != nil {
		return subs
	}
	for _, p := range prims {
		ev := strings.TrimSpace(p.Event)
		if primitive.IsSignal(ev) {
			subs[ev] = append(subs[ev], p.Kind+"/"+p.ID)
		}
	}
	return subs
}

func buildSignalRows(names []string, subs map[string][]string) []signalRow {
	rows := make([]signalRow, 0, len(names))
	for _, n := range names {
		sort.Strings(subs[n])
		rows = append(rows, signalRow{Name: n, Builtin: primitive.IsBuiltinSignal(n), Subscribers: subs[n]})
	}
	return rows
}
