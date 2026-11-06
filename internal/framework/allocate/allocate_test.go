package allocate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tacoda/keystone/internal/framework/config"
	"github.com/tacoda/keystone/internal/framework/primitive"
)

// fixture is a mixed setup: one file per host source allocate reads, plus
// keystone's own projections, which must be skipped.
var fixture = map[string]string{
	".claude/skills/deploy/SKILL.md":          "---\nname: deploy\ndescription: Deploy the app.\n---\nSteps.\n",
	".claude/skills/deploy/scripts/run.sh":    "echo deploy\n",
	".claude/agents/reviewer.md":              "---\nname: reviewer\ndescription: Reviews diffs.\ntools: Read, Grep\n---\nReview.\n",
	".claude/commands/release.md":             "---\ndescription: Cut a release.\n---\nDo it.\n",
	".claude/rules/go/errors.md":              "---\npaths:\n  - \"**/*.go\"\n---\nWrap errors.\n",
	".claude/rules/keystone-owned.md":         "keystone shim\n",
	".cursor/rules/ts.mdc":                    "---\ndescription: TS style\nglobs: src/**/*.ts, test/**/*.ts\nalwaysApply: false\n---\nUse strict.\n",
	".cursor/rules/always.mdc":                "---\nalwaysApply: true\n---\nBe brief.\n",
	".cursor/rules/db-migrations.mdc":         "---\ndescription: How to write a DB migration\n---\nExpand, then contract.\n",
	".cursorrules":                            "Legacy cursor rules.\n",
	".github/copilot-instructions.md":         "Repo-wide Copilot instructions.\n",
	".github/instructions/py.instructions.md": "---\napplyTo: \"**/*.py\"\n---\nType hints.\n",
	".github/prompts/triage.prompt.md":        "---\ndescription: Triage an issue\n---\nTriage.\n",
	".github/agents/planner.agent.md":         "---\ndescription: Plans work\n---\nPlan.\n",
	".clinerules/testing.md":                  "Test first.\n",
	".windsurfrules":                          "Windsurf rules.\n",
	".continue/rules/sql.md":                  "---\nname: sql\nglobs: \"**/*.sql\"\n---\nNo SELECT *.\n",
	".opencode/agent/docs.md":                 "---\ndescription: Writes docs\n---\nDocs.\n",
	".claude/settings.json": `{"permissions":{"allow":["Bash(go test:*)","Bash(make)"],"deny":["Read(./.env*)"]},
		"hooks":{"Stop":[{"matcher":"","hooks":[{"type":"command","command":"make lint"},{"type":"command","command":"keystone hook fire Stop","statusMessage":"keystone:bridge:Stop"}]}],
		"PreToolUse":[{"matcher":"Bash","hooks":[{"type":"command","command":"./guard.sh"}]}]}}`,
}

var wantTargets = map[string]string{
	".claude/skills/deploy/SKILL.md":          "skill .agents/skills/deploy/SKILL.md",
	".claude/skills/deploy/scripts/run.sh":    "skill .agents/skills/deploy/scripts/run.sh",
	".claude/agents/reviewer.md":              "subagent .agents/agents/reviewer.md",
	".claude/commands/release.md":             "command .agents/commands/release.md",
	".claude/rules/go/errors.md":              "rule .agents/rules/go/errors.md",
	".cursor/rules/ts.mdc":                    "rule .agents/rules/ts.md",
	".cursor/rules/always.mdc":                "rule .agents/rules/always.md",
	".cursor/rules/db-migrations.mdc":         "skill .agents/skills/db-migrations/SKILL.md",
	".cursorrules":                            "rule .agents/rules/cursorrules.md",
	".github/copilot-instructions.md":         "rule .agents/rules/copilot-instructions.md",
	".github/instructions/py.instructions.md": "rule .agents/rules/py.md",
	".github/prompts/triage.prompt.md":        "command .agents/commands/triage.md",
	".github/agents/planner.agent.md":         "subagent .agents/agents/planner.md",
	".clinerules/testing.md":                  "rule .agents/rules/cline/testing.md",
	".windsurfrules":                          "rule .agents/rules/windsurfrules.md",
	".continue/rules/sql.md":                  "rule .agents/rules/sql.md",
	".opencode/agent/docs.md":                 "subagent .agents/agents/docs.md",
	".claude/settings.json permissions":       "permission .agents/permissions/claude-settings.md",
	".claude/settings.json hooks.Stop":        "hook .agents/hooks/claude-stop-1.md",
}

func seed(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, body := range files {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

var (
	ownedFixture   = []string{".claude/rules/keystone-owned.md"}
	managedFixture = map[string][]string{".claude/settings.json": {"allow:Bash(make)"}}
)

func TestPlan_MapsEverySource(t *testing.T) {
	items, err := Plan(seed(t, fixture), ownedFixture, managedFixture)
	if err != nil {
		t.Fatal(err)
	}
	got := placements(items)
	for src, want := range wantTargets {
		if got[src] != want {
			t.Errorf("%s → %q, want %q", src, got[src], want)
		}
	}
	if got[".claude/settings.json hooks.PreToolUse[Bash]"] != "stays" {
		t.Error("a hook with a matcher must stay in settings.json")
	}
	if _, ok := got[".claude/rules/keystone-owned.md"]; ok {
		t.Error("keystone's own projection must be skipped")
	}
	if n := len(got); n != len(wantTargets)+1 {
		t.Errorf("got %d items, want %d: %v", n, len(wantTargets)+1, got)
	}
}

// placements maps each item's source to "kind target", or "stays".
func placements(items []Item) map[string]string {
	got := map[string]string{}
	for _, it := range items {
		got[it.Source] = "stays"
		if it.Moves() {
			got[it.Source] = string(it.Kind) + " " + it.Target
		}
	}
	return got
}

// TestApply_RoundTrip — the converted charter lints clean, every original
// is unchanged, and a second apply overwrites nothing.
func TestApply_RoundTrip(t *testing.T) {
	root := seed(t, fixture)
	items, _ := Plan(root, ownedFixture, managedFixture)
	if _, err := Apply(root, items); err != nil {
		t.Fatal(err)
	}
	assertOriginalsUnchanged(t, root)
	assertLintClean(t, root)
	assertFile(t, root, ".agents/rules/ts.md", "paths:\n  - src/**/*.ts\n  - test/**/*.ts", "id: ts")
	assertFile(t, root, ".agents/agents/reviewer.md", "tools:\n  - Read\n  - Grep")
	assertFile(t, root, ".agents/permissions/claude-settings.md", `"Bash(go test:*)"`, `"Read(./.env*)"`)
	assertAbsent(t, root, ".agents/permissions/claude-settings.md", "Bash(make)") // keystone's managed allow
	assertAbsent(t, root, ".agents/rules/always.md", "paths")                     // alwaysApply: true
	again, _ := Apply(root, items)
	if len(again.Wrote) != 0 || len(again.Exists) == 0 {
		t.Errorf("second apply wrote %v", again.Wrote)
	}
}

func assertOriginalsUnchanged(t *testing.T, root string) {
	t.Helper()
	for rel, body := range fixture {
		if data, _ := os.ReadFile(filepath.Join(root, rel)); string(data) != body {
			t.Errorf("original %s changed", rel)
		}
	}
}

// assertLintClean: no lint error except a subagent's missing `tools:` (it
// had every tool on its host; allocate says so in the item's note rather
// than guess a list).
func assertLintClean(t *testing.T, root string) {
	t.Helper()
	prims, warns, err := primitive.Walk(root, config.DefaultAgentsRoot)
	if err != nil || len(warns) > 0 {
		t.Fatalf("walk: %v %v", err, warns)
	}
	for _, f := range primitive.Lint(prims) {
		if f.Severity == primitive.FindingError && !strings.Contains(f.Message, "missing `tools:`") {
			t.Errorf("lint: %s", f)
		}
	}
}

func assertAbsent(t *testing.T, root, rel, unwanted string) {
	t.Helper()
	if data, _ := os.ReadFile(filepath.Join(root, rel)); strings.Contains(string(data), unwanted) {
		t.Errorf("%s must not contain %q", rel, unwanted)
	}
}

func assertFile(t *testing.T, root, rel string, wants ...string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Fatalf("%s not written: %v", rel, err)
	}
	for _, w := range wants {
		if !strings.Contains(string(data), w) {
			t.Errorf("%s missing %q:\n%s", rel, w, data)
		}
	}
}
