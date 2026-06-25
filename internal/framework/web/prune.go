package web

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

// PruneCandidate is one primitive flagged for review by the pruner.
// The Reason explains why; the dashboard renders it next to a
// "prune" button that DELETEs the file.
type PruneCandidate struct {
	Kind     string `json:"kind"`
	ID       string `json:"id"`
	Path     string `json:"path"`
	Reason   string `json:"reason"`
	Severity string `json:"severity"` // "error" | "warning" | "info"
}

// findPruneCandidates runs the heuristics described in the pruning
// flywheel against the current harness. Aggregates with lint
// findings so the dashboard surfaces both in one place.
//
// Heuristics applied:
//   - lint errors  → severity=error
//   - lint warnings → severity=warning
//   - guides / rules with no `traces:` AND no other primitive's
//     `deps:` referencing them → severity=info ("not referenced")
//   - primitives with empty body (≤ 200 bytes incl. frontmatter)
//     → severity=info ("empty body")
//   - primitives sharing an identical description with another
//     → severity=warning ("duplicate description")
//
// Effectiveness ("rule was applied but had no impact") needs
// telemetry we don't have yet — surfaced in docs as future work,
// not flagged here.
func findPruneCandidates(projectDir string, primitives []primitive.Primitive) []PruneCandidate {
	var out []PruneCandidate

	// Lint findings.
	findings := primitive.Lint(primitives)
	for _, f := range findings {
		out = append(out, PruneCandidate{
			Kind:     f.Kind,
			ID:       f.ID,
			Path:     f.Path,
			Reason:   "lint: " + f.Message,
			Severity: string(f.Severity),
		})
	}

	// Reference graph — which ids are pointed at by another
	// primitive's deps or traces.
	referenced := map[string]bool{}
	for _, p := range primitives {
		for _, dep := range p.Deps {
			referenced[dep] = true
		}
		for _, tr := range p.Corpus {
			if !strings.Contains(tr, "/") {
				tr = "corpus/" + tr
			}
			referenced["corpus/"+tr] = true
			referenced[tr] = true
		}
	}

	// Not-referenced: guides/rules nobody points at + corpus
	// entries nobody traces to.
	for _, p := range primitives {
		key := p.Kind + "/" + p.ID
		if p.Kind == "corpus" && !referenced[key] && !referenced[p.ID] {
			out = append(out, PruneCandidate{
				Kind:     p.Kind,
				ID:       p.ID,
				Path:     p.Path,
				Reason:   "not referenced by any rule's `traces:`",
				Severity: "info",
			})
		}
	}

	// Empty-body candidates.
	for _, p := range primitives {
		info, err := os.Stat(filepath.Join(projectDir, p.Path))
		if err != nil {
			continue
		}
		if info.Size() <= 200 {
			out = append(out, PruneCandidate{
				Kind:     p.Kind,
				ID:       p.ID,
				Path:     p.Path,
				Reason:   "body ≤ 200 bytes — likely empty stub",
				Severity: "info",
			})
		}
	}

	// Duplicate descriptions.
	byDesc := map[string][]primitive.Primitive{}
	for _, p := range primitives {
		d := strings.TrimSpace(p.Description)
		if d == "" {
			continue
		}
		byDesc[d] = append(byDesc[d], p)
	}
	for desc, group := range byDesc {
		if len(group) < 2 {
			continue
		}
		others := []string{}
		for _, g := range group {
			others = append(others, g.Kind+"/"+g.ID)
		}
		for _, g := range group {
			out = append(out, PruneCandidate{
				Kind:     g.Kind,
				ID:       g.ID,
				Path:     g.Path,
				Reason:   "duplicate description (also: " + strings.Join(stripSelf(others, g.Kind+"/"+g.ID), ", ") + ") — \"" + desc + "\"",
				Severity: "warning",
			})
		}
	}

	// Stable order: severity (error > warning > info), then path.
	sort.SliceStable(out, func(i, j int) bool {
		sev := func(s string) int {
			switch s {
			case "error":
				return 0
			case "warning":
				return 1
			default:
				return 2
			}
		}
		if sev(out[i].Severity) != sev(out[j].Severity) {
			return sev(out[i].Severity) < sev(out[j].Severity)
		}
		return out[i].Path < out[j].Path
	})
	return out
}

func stripSelf(in []string, self string) []string {
	out := in[:0:0]
	for _, s := range in {
		if s == self {
			continue
		}
		out = append(out, s)
	}
	return out
}

// handlePrune renders the prune dashboard.
func (s *server) handlePrune(w http.ResponseWriter, r *http.Request) {
	primitives, err := s.loadPrimitives()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	candidates := findPruneCandidates(s.projectDir, primitives)
	counts := map[string]int{}
	for _, c := range candidates {
		counts[c.Severity]++
	}
	s.render(w, "prune.html", map[string]any{
		"ProjectDir": s.projectDir,
		"Candidates": candidates,
		"Counts":     counts,
	})
}
