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
	r := measureFixture(t, nil, "claude-code", "codex")
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
	}, "claude-code")
	var scopes []string
	for _, v := range r.Violations {
		scopes = append(scopes, v.Scope)
	}
	if got := strings.Join(scopes, ","); got != "always_loaded (claude-code),reference per load" {
		t.Errorf("violations = %s", got)
	}
}

// TestMeasure_DefaultAlwaysCap — with no budgets configured, the 6.0
// default always_loaded cap applies and names the over-cap files; an
// explicit cap overrides it; max_tokens: 0 disables it.
func TestMeasure_DefaultAlwaysCap(t *testing.T) {
	// No budgets: a fat AGENTS.md blows the default cap.
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
	fat := strings.Repeat("word ", DefaultAlwaysLoadedCap+100) + "\n"
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte(fat), 0o644); err != nil {
		t.Fatal(err)
	}
	prims, _, err := primitive.Walk(root, config.DefaultAgentsRoot)
	if err != nil {
		t.Fatal(err)
	}
	r := Measure(root, prims, projection.Enabled([]string{"codex"}), nil)
	var vs []Violation
	for _, v := range r.Violations {
		if v.Scope == AlwaysLoaded+" (codex)" {
			vs = append(vs, v)
		}
	}
	if len(vs) != 1 {
		t.Fatalf("default cap: want 1 always_loaded violation, got %+v", r.Violations)
	}
	if vs[0].Cap != DefaultAlwaysLoadedCap || vs[0].OverBy() <= 0 {
		t.Errorf("violation = %+v", vs[0])
	}
	if !strings.Contains(vs[0].Path, "AGENTS.md") {
		t.Errorf("violation must name the over-cap file, got path %q", vs[0].Path)
	}

	// Explicit smaller cap wins over the default.
	r = Measure(root, prims, projection.Enabled([]string{"codex"}), map[string]config.BudgetSpec{
		AlwaysLoaded: {MaxTokens: 10},
	})
	found := false
	for _, v := range r.Violations {
		if v.Scope == AlwaysLoaded+" (codex)" && v.Cap == 10 {
			found = true
		}
	}
	if !found {
		t.Errorf("explicit cap not applied: %+v", r.Violations)
	}

	// Explicit zero disables the cap.
	r = Measure(root, prims, projection.Enabled([]string{"codex"}), map[string]config.BudgetSpec{
		AlwaysLoaded: {MaxTokens: 0},
	})
	for _, v := range r.Violations {
		if v.Scope == AlwaysLoaded+" (codex)" {
			t.Errorf("max_tokens: 0 must disable the cap, got %+v", v)
		}
	}
}
