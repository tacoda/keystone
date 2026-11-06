package primitive

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// fixtureTree writes a minimal charter + .claude layout under t.TempDir
// and returns the project root.
func fixtureTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	files := map[string]string{
		"charter/guides/idioms/rails/migrations.md": `---
kind: guide
id: idioms/rails/migrations
description: Migrations are reversible and small.
tier: golden-rule
globs:
  - "db/migrate/**"
corpus:
  - corpus/idioms/rails/migrations
---
# body
`,
		"charter/guides/README.md": "# README skipped\n",
		"charter/corpus/idioms/rails/migrations.md": `---
kind: corpus
id: corpus/idioms/rails/migrations
description: Why migrations stay small.
---
# body
`,
		"charter/sensors/drift.md": `---
kind: sensor
id: drift
description: Flag rules whose corpus is missing.
phase: verify
---
# body
`,
		"charter/sensors/build.md": `---
kind: sensor
id: build
description: Build at the verify gate.
mode: computational
on: pre-verify
run: go build ./...
---
# body
`,
		"charter/skills/review-code/SKILL.md": `---
kind: skill
id: review-code
description: Review a PR for logic, style, security.
triggers:
  - review this PR
  - code review
  - /review
---
# body
`,
		"charter/agents/cavecrew-reviewer.md": `---
kind: agent
id: cavecrew-reviewer
description: Compressed diff reviewer.
tools:
  - Read
  - Grep
  - Bash
---
# body
`,
		"charter/commands/verify.md": `---
kind: command
id: verify
description: Slash command for the verify lifecycle step.
args:
  - target
---
# body
`,
		// File w/ no frontmatter — must be silently skipped.
		"charter/guides/stub.md": "# no frontmatter\n",
	}
	for rel, body := range files {
		abs := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestWalk_FindsAllKindsSkipsReadmesAndNonFM(t *testing.T) {
	root := fixtureTree(t)
	primitives, warnings, err := Walk(root, "charter")
	if err != nil {
		t.Fatalf("Walk err: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("warnings = %v; want none", warnings)
	}
	got := map[string]string{}
	for _, p := range primitives {
		got[p.Kind+"/"+p.ID] = p.Path
	}
	wantIDs := []string{
		"rule/idioms/rails/migrations",
		"reference/corpus/idioms/rails/migrations",
		"hook/drift",
		"hook/build",
		"skill/review-code",
		"subagent/cavecrew-reviewer",
		"command/verify",
	}
	for _, id := range wantIDs {
		if _, ok := got[id]; !ok {
			t.Errorf("missing primitive: %s", id)
		}
	}
	if len(primitives) != len(wantIDs) {
		t.Errorf("primitive count = %d, want %d; got = %+v", len(primitives), len(wantIDs), got)
	}
}

func TestBuildAndWrite(t *testing.T) {
	root := fixtureTree(t)
	primitives, _, err := Walk(root, "charter")
	if err != nil {
		t.Fatal(err)
	}
	idx := Build(primitives, time.Date(2026, 6, 17, 0, 0, 0, 0, time.UTC))
	if idx.Version != IndexVersion {
		t.Errorf("Version = %q, want %q", idx.Version, IndexVersion)
	}
	if IndexVersion != "5.0" {
		t.Errorf("IndexVersion = %q, want 5.0", IndexVersion)
	}
	if idx.Generated != "2026-06-17T00:00:00Z" {
		t.Errorf("Generated = %q", idx.Generated)
	}
	if len(idx.ByKind["rule"]) != 1 {
		t.Errorf("ByKind[rule] = %v", idx.ByKind["rule"])
	}
	if len(idx.ByPath["db/migrate/**"]) != 1 ||
		idx.ByPath["db/migrate/**"][0] != "rule/idioms/rails/migrations" {
		t.Errorf("ByPath = %v", idx.ByPath)
	}

	out := filepath.Join(root, "charter", "INDEX.json")
	if err := Write(out, idx); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Index
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal INDEX.json: %v", err)
	}
	if len(decoded.Primitives) != len(primitives) {
		t.Errorf("decoded primitives = %d, want %d", len(decoded.Primitives), len(primitives))
	}
}

// TestBuild_ByCategory — every indexed primitive lands in its category.
func TestBuild_ByCategory(t *testing.T) {
	primitives, _, err := Walk(fixtureTree(t), "charter")
	if err != nil {
		t.Fatal(err)
	}
	idx := Build(primitives, time.Unix(0, 0))
	// rule + reference + 2 hooks; skill + subagent + command.
	if c, k := len(idx.ByCategory["constraint"]), len(idx.ByCategory["capability"]); c != 4 || k != 3 {
		t.Errorf("ByCategory = %v, want 4 constraints and 3 capabilities", idx.ByCategory)
	}
}

// TestBuildLite_CategoryOrder — the lite index lists constraints before
// capabilities and stamps each entry with its category.
func TestBuildLite_CategoryOrder(t *testing.T) {
	prims := []Primitive{
		{Frontmatter: Frontmatter{Kind: "skill", ID: "a"}},
		{Frontmatter: Frontmatter{Kind: "hook", ID: "b"}},
		{Frontmatter: Frontmatter{Kind: "command", ID: "c"}},
		{Frontmatter: Frontmatter{Kind: "rule", ID: "d"}},
	}
	lite := BuildLite(Build(prims, time.Unix(0, 0)))
	want := []string{"constraint/hook/b", "constraint/rule/d", "capability/command/c", "capability/skill/a"}
	for i, e := range lite.Primitives {
		if got := e.Category + "/" + e.Kind + "/" + e.ID; got != want[i] {
			t.Errorf("lite[%d] = %s, want %s", i, got, want[i])
		}
	}
}
