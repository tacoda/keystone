package budget

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
	"github.com/tacoda/keystone/internal/framework/projection"
)

var charterFixture = map[string]string{
	"AGENTS.md":                      "one two three four five\n",
	"CLAUDE.md":                      "six seven\n",
	".agents/skills/deploy/SKILL.md": "---\nname: deploy\ndescription: Deploy it now.\n---\n" + strings.Repeat("word ", 40),
	".agents/rules/go/style.md":      "---\nid: go/style\ndescription: Go style.\npaths:\n  - \"**/*.go\"\n---\n" + strings.Repeat("word ", 20),
	".agents/references/why.md":      "---\nid: why\ndescription: Why.\n---\n" + strings.Repeat("word ", 100),
	".agents/hooks/lint.md":          "---\nid: lint\ndescription: Lint.\nmode: computational\nevent: Stop\nrun: \"true\"\n---\n",
}

func measureFixture(t *testing.T, budgets map[string]config.BudgetSpec, adapters ...string) Report {
	t.Helper()
	root := t.TempDir()
	for rel, body := range charterFixture {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	prims, _, err := primitive.Walk(root, config.DefaultAgentsRoot)
	if err != nil {
		t.Fatal(err)
	}
	return Measure(root, prims, projection.Enabled(adapters), budgets)
}

// TestMeasure_TwoTiersPerHost — a host loads its root files and the
// listing of each skill it projects; bodies and path-scoped rules wait.
func TestMeasure_TwoTiersPerHost(t *testing.T) {
	r := measureFixture(t, nil, "codex")
	always := map[string]int{}
	for _, h := range r.Hosts {
		always[h.Host] = h.Always
	}
	listing := Estimate([]byte("deploy Deploy it now."))
	if want := 5 + 2 + listing; always["claude-code"] != want {
		t.Errorf("claude-code always = %d, want %d (CLAUDE.md + AGENTS.md + skill listing): %+v", always["claude-code"], want, r.Hosts)
	}
	if want := 5 + listing; always["codex"] != want {
		t.Errorf("codex always = %d, want %d (AGENTS.md + skill listing)", always["codex"], want)
	}
	for _, kind := range []string{"skill", "rule", "reference"} {
		if r.ByKind[kind] == 0 {
			t.Errorf("on-demand %s missing: %v", kind, r.ByKind)
		}
	}
	if r.ByKind["hook"] != 0 {
		t.Error("a hook is not context")
	}
	if r.OnDemand[0].Path != ".agents/references/why.md" {
		t.Errorf("largest on-demand = %s", r.OnDemand[0].Path)
	}
}

func TestMeasure_Caps(t *testing.T) {
	r := measureFixture(t, map[string]config.BudgetSpec{
		AlwaysLoaded: {MaxTokens: 8},
		"corpus":     {MaxTokensPerLoad: 50}, // 4.x port key → reference
		"skills":     {MaxTokens: 1000},
	})
	var scopes []string
	for _, v := range r.Violations {
		scopes = append(scopes, v.Scope)
	}
	if got := strings.Join(scopes, ","); got != "always_loaded (claude-code),reference per load" {
		t.Errorf("violations = %s", got)
	}
}
