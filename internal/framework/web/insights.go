package web

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// Insight is one suggested action the dashboard surfaces to improve
// the harness. Severities are "high" (act soon), "medium" (worth
// addressing next pass), "low" (informational).
type Insight struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Detail   string `json:"detail"`
	Severity string `json:"severity"`
	// Action is the canonical follow-up — either a URL into the
	// dashboard, or a CLI invocation the user copies.
	Action     string `json:"action,omitempty"`
	ActionLink string `json:"action_link,omitempty"`
}

// collectInsights derives suggestions from the current harness state.
// Pure read; no I/O outside the harness tree. Stable order: severity
// then id.
func (s *server) collectInsights(primitives []primitive.Primitive) []Insight {
	var out []Insight

	// 1. Lint errors → high.
	findings := primitive.Lint(primitives)
	errs := 0
	for _, f := range findings {
		if f.Severity == primitive.FindingError {
			errs++
		}
	}
	if errs > 0 {
		out = append(out, Insight{
			ID: "lint.errors", Severity: "high",
			Title:      "lint errors present",
			Detail:     "Hard schema errors (missing fields, duplicate ids, malformed frontmatter) make primitives invisible to the agent. Fix or prune.",
			Action:     "review /flywheels/prune",
			ActionLink: "/flywheels/prune",
		})
	}

	// 2. Inbox candidates ageing → medium / high.
	inboxAge, inboxCount := inboxStats(s.projectDir)
	if inboxCount > 0 {
		sev := "medium"
		if inboxAge > 14*24*time.Hour {
			sev = "high"
		}
		out = append(out, Insight{
			ID: "inbox.backlog", Severity: sev,
			Title:      "learning inbox is backed up",
			Detail:     "Captured candidates haven't been triaged. Long backlogs make the agent stop capturing — promote or reject.",
			Action:     "walk /flywheels/inbox",
			ActionLink: "/flywheels/inbox",
		})
	}

	// 3. Stale INDEX → medium.
	indexPath := filepath.Join(s.projectDir, config.KeystoneDir(config.DefaultHarnessRoot), config.IndexName)
	indexInfo, indexErr := os.Stat(indexPath)
	if indexErr != nil {
		out = append(out, Insight{
			ID: "index.missing", Severity: "high",
			Title:      "no .keystone/INDEX.json",
			Detail:     "Agents read the index first. Without it, every primitive must be opened blindly — defeats the staged resolution flow.",
			Action:     "run `keystone index`",
			ActionLink: "",
		})
	} else {
		stale := false
		for _, p := range primitives {
			if pi, err := os.Stat(filepath.Join(s.projectDir, p.Path)); err == nil {
				if pi.ModTime().After(indexInfo.ModTime()) {
					stale = true
					break
				}
			}
		}
		if stale {
			out = append(out, Insight{
				ID: "index.stale", Severity: "medium",
				Title:      "INDEX.json is older than at least one primitive",
				Detail:     "Re-run `keystone index` so the agent's descriptor surface matches the harness on disk.",
				Action:     "run `keystone index`",
			})
		}
	}

	// 4. No sources configured → low (informational).
	if cfg, _ := readContextDoc(s.projectDir); cfg != nil {
		if arr, _ := cfg["sources"].([]any); len(arr) == 0 {
			out = append(out, Insight{
				ID: "sources.none", Severity: "low",
				Title:      "no external sources configured",
				Detail:     "Stage 3 of the resolution flow is unavailable. Configure folder / url adapters in .keystone/context.json when in-harness rules + corpus aren't enough.",
				Action:     "/sources/new",
				ActionLink: "/sources/new",
			})
		}
	}

	// 5. Corpus without a referencing guide → low.
	orphans := orphanCorpus(primitives)
	if len(orphans) > 0 {
		out = append(out, Insight{
			ID: "corpus.orphan", Severity: "low",
			Title:      "corpus entries without a referrer",
			Detail:     "These corpus files aren't reached by any rule's `traces:`. Either link them or prune.",
			Action:     "review /flywheels/prune",
			ActionLink: "/flywheels/prune",
		})
	}

	// 6. No skills authored → low.
	if countByKind(primitives, "skill") == 0 {
		out = append(out, Insight{
			ID: "skills.zero", Severity: "low",
			Title:      "no project-authored skills",
			Detail:     "Skills are host-native agent abstractions Claude Code auto-loads by trigger phrase. Adding even one (`keystone new skill`) makes recurring tasks one-phrase.",
			Action:     "/harness/primitives/new",
			ActionLink: "/harness/primitives/new",
		})
	}

	// 7. Iron-law concentration — too few high-severity rules can
	//    indicate undertyped guides; too many can indicate cargo-
	//    culting. Surface both as low.
	if mustCount, totalGuides := severityCount(primitives, "must"); totalGuides > 0 {
		ratio := float64(mustCount) / float64(totalGuides)
		switch {
		case ratio > 0.6:
			out = append(out, Insight{
				ID: "severity.over-iron", Severity: "low",
				Title:      "more than 60% of guides are severity: must",
				Detail:     "When everything is iron-law, nothing is. Demote routine rules to `should` so the agent can tell signal from boilerplate.",
				Action:     "/harness/primitives?kind=guide",
				ActionLink: "/harness/primitives?kind=guide",
			})
		case ratio < 0.05 && totalGuides > 10:
			out = append(out, Insight{
				ID: "severity.under-iron", Severity: "low",
				Title:      "<5% of guides are severity: must",
				Detail:     "If no rules are iron-law, the agent has no clear non-negotiables. Promote a handful (security, data-loss prevention, etc.).",
				Action:     "/harness/primitives?kind=guide",
				ActionLink: "/harness/primitives?kind=guide",
			})
		}
	}

	sort.SliceStable(out, func(i, j int) bool {
		sev := func(s string) int {
			switch s {
			case "high":
				return 0
			case "medium":
				return 1
			default:
				return 2
			}
		}
		if sev(out[i].Severity) != sev(out[j].Severity) {
			return sev(out[i].Severity) < sev(out[j].Severity)
		}
		return out[i].ID < out[j].ID
	})
	return out
}

func inboxStats(projectDir string) (oldest time.Duration, count int) {
	dir := filepath.Join(projectDir, config.DefaultHarnessRoot, "learning", "inbox")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, 0
	}
	now := time.Now()
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") || e.Name() == "README.md" {
			continue
		}
		count++
		if info, err := e.Info(); err == nil {
			age := now.Sub(info.ModTime())
			if age > oldest {
				oldest = age
			}
		}
	}
	return oldest, count
}

func orphanCorpus(primitives []primitive.Primitive) []string {
	referenced := map[string]bool{}
	for _, p := range primitives {
		for _, t := range p.Traces {
			referenced[t] = true
			if !strings.HasPrefix(t, "corpus/") {
				referenced["corpus/"+t] = true
			}
		}
	}
	var out []string
	for _, p := range primitives {
		if p.Kind != "corpus" {
			continue
		}
		if !referenced[p.ID] && !referenced["corpus/"+p.ID] {
			out = append(out, p.ID)
		}
	}
	return out
}

func countByKind(primitives []primitive.Primitive, kind string) int {
	n := 0
	for _, p := range primitives {
		if p.Kind == kind {
			n++
		}
	}
	return n
}

func severityCount(primitives []primitive.Primitive, severity string) (matched, totalGuides int) {
	for _, p := range primitives {
		if p.Kind != "guide" {
			continue
		}
		totalGuides++
		if p.Severity == severity {
			matched++
		}
	}
	return matched, totalGuides
}

// handleInsights renders the insights page.
func (s *server) handleInsights(w http.ResponseWriter, r *http.Request) {
	primitives, err := s.loadPrimitives()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	insights := s.collectInsights(primitives)
	counts := map[string]int{}
	for _, in := range insights {
		counts[in.Severity]++
	}
	s.renderPage(w, r, "insights.html", map[string]any{
		"ProjectDir": s.projectDir,
		"Insights":   insights,
		"Counts":     counts,
	})
}

// /api/insights JSON.
func (s *server) apiInsights(w http.ResponseWriter, r *http.Request) {
	primitives, err := s.loadPrimitives()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"insights": s.collectInsights(primitives),
	})
}
