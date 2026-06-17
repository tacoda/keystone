package primitive

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// fixtureTree writes a minimal harness + .claude layout under t.TempDir
// and returns the project root.
func fixtureTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	files := map[string]string{
		"harness/guides/idioms/rails/migrations.md": `---
kind: guide
id: idioms/rails/migrations
description: Migrations are reversible and small.
severity: must
globs:
  - "db/migrate/**"
traces:
  - corpus/idioms/rails/migrations
---
# body
`,
		"harness/guides/README.md": "# README skipped\n",
		"harness/actions/verify.md": `---
kind: action
id: verify
description: Run pre-commit and review.
phase: verify
deps:
  - sensor/lint
---
# body
`,
		"harness/corpus/idioms/rails/migrations.md": `---
kind: corpus
id: corpus/idioms/rails/migrations
description: Why migrations stay small.
---
# body
`,
		"harness/sensors/drift.md": `---
kind: sensor
id: drift
description: Flag rules whose corpus is missing.
phase: verify
---
# body
`,
		"harness/skills/review-code/SKILL.md": `---
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
		"harness/agents/cavecrew-reviewer.md": `---
kind: subagent
id: cavecrew-reviewer
description: Compressed diff reviewer.
tools:
  - Read
  - Grep
  - Bash
---
# body
`,
		"harness/commands/verify.md": `---
kind: command
id: verify
description: Slash command for verify action.
args:
  - target
---
# body
`,
		// File w/ no frontmatter — must be silently skipped.
		"harness/actions/stub.md": "# no frontmatter\n",
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
	primitives, warnings, err := Walk(root, "harness")
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
		"guide/idioms/rails/migrations",
		"action/verify",
		"corpus/corpus/idioms/rails/migrations",
		"sensor/drift",
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
	primitives, _, err := Walk(root, "harness")
	if err != nil {
		t.Fatal(err)
	}
	idx := Build(primitives, time.Date(2026, 6, 17, 0, 0, 0, 0, time.UTC))
	if idx.Version != IndexVersion {
		t.Errorf("Version = %q, want %q", idx.Version, IndexVersion)
	}
	if IndexVersion != "2.0" {
		t.Errorf("IndexVersion = %q, want 2.0", IndexVersion)
	}
	if idx.Generated != "2026-06-17T00:00:00Z" {
		t.Errorf("Generated = %q", idx.Generated)
	}
	if len(idx.ByKind["guide"]) != 1 {
		t.Errorf("ByKind[guide] = %v", idx.ByKind["guide"])
	}
	if len(idx.ByGlob["db/migrate/**"]) != 1 ||
		idx.ByGlob["db/migrate/**"][0] != "guide/idioms/rails/migrations" {
		t.Errorf("ByGlob = %v", idx.ByGlob)
	}

	out := filepath.Join(root, "harness", "INDEX.json")
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
