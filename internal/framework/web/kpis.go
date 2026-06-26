package web

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// kpiNames is the ordered list of KPI widget names rendered into
// the Observability KPI strip. Add a new one here + a `case` in
// handleKPIWidget below; the template loops on this list so the
// addition shows up automatically.
func kpiNames() []string {
	return []string{"primitives", "inbox", "lint", "index"}
}

// kpi is one of the headline numbers on the Observability landing
// page. Each is loaded lazily by its own widget endpoint so a slow
// computation never delays first paint, and refreshed live via the
// narrowest SSE topic it cares about.
type kpi struct {
	Label string   // e.g. "primitives"
	Value string   // displayed number / status
	Tone  string   // "" | "ok" | "warn" | "bad" — CSS class
	Hint  string   // optional muted line under the value
	Topic sseTopic // SSE topic that should re-trigger this widget
}

// kpiPrimitives returns the total primitive count.
func (s *server) kpiPrimitives() kpi {
	primitives, err := s.loadPrimitives()
	if err != nil {
		return kpi{Label: "primitives", Value: "—", Tone: "bad", Hint: err.Error(), Topic: topicPrimitives}
	}
	return kpi{
		Label: "primitives",
		Value: fmt.Sprintf("%d", len(primitives)),
		Topic: topicPrimitives,
	}
}

// kpiInbox returns the count of un-triaged learning candidates.
// Tone escalates as the backlog grows — devs use this as a
// "do I owe someone synthesize?" signal.
func (s *server) kpiInbox() kpi {
	_, count := inboxStats(s.projectDir)
	tone := ""
	switch {
	case count == 0:
		tone = "ok"
	case count > 10:
		tone = "warn"
	}
	return kpi{
		Label: "inbox",
		Value: fmt.Sprintf("%d", count),
		Tone:  tone,
		Hint:  "learning candidates",
		Topic: topicInbox,
	}
}

// kpiLint returns the count of hard lint errors across the harness.
// Bad tone if non-zero; ok otherwise. Lint is the cheapest signal
// the agent has that the harness is in a broken state.
func (s *server) kpiLint() kpi {
	primitives, err := s.loadPrimitives()
	if err != nil {
		return kpi{Label: "lint", Value: "—", Tone: "bad", Hint: err.Error(), Topic: topicPrimitives}
	}
	findings := primitive.Lint(primitives)
	errs := 0
	for _, f := range findings {
		if f.Severity == primitive.FindingError {
			errs++
		}
	}
	tone := "ok"
	if errs > 0 {
		tone = "bad"
	}
	return kpi{
		Label: "lint errors",
		Value: fmt.Sprintf("%d", errs),
		Tone:  tone,
		Topic: topicPrimitives,
	}
}

// kpiIndex reports whether INDEX.json is fresher than every
// primitive on disk. Stale index = the agent sees a different
// surface than what's actually shipped — the most common cause of
// "I edited the rule and the agent didn't notice."
func (s *server) kpiIndex() kpi {
	indexPath := filepath.Join(s.projectDir, config.KeystoneDir(config.DefaultHarnessRoot), config.IndexName)
	if _, err := os.Stat(indexPath); err != nil {
		return kpi{Label: "index", Value: "missing", Tone: "bad", Hint: "run `keystone index`", Topic: topicPrimitives}
	}
	snap, err := s.collectMetrics(context.Background())
	if err != nil {
		return kpi{Label: "index", Value: "—", Tone: "warn", Hint: err.Error(), Topic: topicPrimitives}
	}
	tone, value := "ok", "fresh"
	if !snap.IndexFresh {
		tone, value = "warn", "stale"
	}
	return kpi{Label: "index", Value: value, Tone: tone, Hint: snap.IndexAge, Topic: topicPrimitives}
}

// handleKPIWidget dispatches `/web/widgets/kpi/<name>` to the right
// computation and renders the shared `_kpi.html` partial. Unknown
// name yields 404.
//
// Routes register `/web/widgets/kpi/` as a NoRoute prefix; this
// handler reads the trailing segment.
func (s *server) handleKPIWidget(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/web/widgets/kpi/")
	name = strings.Trim(name, "/")
	var k kpi
	switch name {
	case "primitives":
		k = s.kpiPrimitives()
	case "inbox":
		k = s.kpiInbox()
	case "lint":
		k = s.kpiLint()
	case "index":
		k = s.kpiIndex()
	default:
		http.NotFound(w, r)
		return
	}
	s.render(w, "_kpi.html", map[string]any{
		"KPI":  k,
		"Name": name,
	})
}

