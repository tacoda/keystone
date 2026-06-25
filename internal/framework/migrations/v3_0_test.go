package migrations

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

func seed24Install(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	files := map[string]string{
		".keystone/harness/guides/idioms/go/stdlib-first.md": "---\nkind: guide\nid: guides/idioms/go/stdlib-first\ndescription: x\nglobs:\n  - \"**/*.go\"\ntraces:\n  - corpus/idioms/go/stdlib-first\n---\n# Stdlib first\n",
		".keystone/harness/sensors/build.md":                 "---\nkind: sensor\nid: build\ndescription: x\nhost_triggers:\n  - phase: Stop\n    command: go build ./...\n---\n# build\n",
		".keystone/harness/sensors/review-functional.md":     "---\nkind: sensor\nid: review-functional\ndescription: x\n---\n# functional review\n",
		".keystone/harness/actions/verify.md":                "---\nkind: action\nid: verify\ndescription: x\n---\n# verify\n",
		".keystone/harness/playbooks/task.md":                "---\nkind: playbook\nid: task\ndescription: x\n---\n# task\n",
		".keystone/harness/personas/code-reviewer.md":        "---\nkind: persona\nid: code-reviewer\ndescription: x\ntools:\n  - Read\n---\n# reviewer\n",
		".keystone/harness/corpus/idioms/go/stdlib-first.md": "---\nkind: corpus\nid: corpus/idioms/go/stdlib-first\ndescription: x\n---\n# why\n",
	}
	for rel, body := range files {
		abs := filepath.Join(tmp, rel)
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return tmp
}

func runV3Up(t *testing.T, tmp string) {
	t.Helper()
	plan, err := planUp_3_0(tmp)
	if err != nil {
		t.Fatalf("planUp_3_0: %v", err)
	}
	if err := plan.Execute(tmp); err != nil {
		t.Fatalf("execute: %v", err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

func TestV3_Up_RenamesAndMoves(t *testing.T) {
	tmp := seed24Install(t)
	runV3Up(t, tmp)

	h := filepath.Join(tmp, ".harness")
	// guide → rule (rules/), kind rewritten, traces → corpus.
	rule := readFile(t, filepath.Join(h, "rules/idioms/go/stdlib-first.md"))
	if !contains(rule, "kind: rule") || contains(rule, "kind: guide") {
		t.Errorf("guide not converted to rule:\n%s", rule)
	}
	if !contains(rule, "corpus:") || contains(rule, "traces:") {
		t.Errorf("traces not renamed to corpus:\n%s", rule)
	}
	// Each retired primitive landed at its 3.0 canonical path:
	// computational sensor → hook; inferential sensor → agent;
	// action → command; persona → agent; playbook → skills/<id>/SKILL.md.
	for _, rel := range []string{
		"hooks/build.md",
		"agents/review-functional.md",
		"commands/verify.md",
		"agents/code-reviewer.md",
		"skills/task/SKILL.md",
		"work/roadmaps/.gitkeep",
	} {
		if _, err := os.Stat(filepath.Join(h, rel)); err != nil {
			t.Errorf("expected migrated path %s: %v", rel, err)
		}
	}
	// old dirs drained.
	for _, old := range []string{"guides", "sensors", "actions", "playbooks", "personas"} {
		if len(leafFiles(h, old)) != 0 {
			t.Errorf("old dir %s still has primitive files", old)
		}
	}
}

func TestV3_Up_PostMigrationLintClean(t *testing.T) {
	tmp := seed24Install(t)
	runV3Up(t, tmp)
	prims, _, err := primitive.Walk(tmp, ".harness")
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	for _, f := range primitive.Lint(prims) {
		if f.Severity == primitive.FindingError {
			t.Errorf("post-migration lint error: %s", f)
		}
	}
}

func TestV3_Up_Idempotent(t *testing.T) {
	tmp := seed24Install(t)
	runV3Up(t, tmp)
	// Second run must be a no-op (no old dirs left to convert).
	runV3Up(t, tmp)
	if _, err := os.Stat(filepath.Join(tmp, ".harness/rules/idioms/go/stdlib-first.md")); err != nil {
		t.Errorf("rule missing after second Up: %v", err)
	}
}

func contains(s, sub string) bool { return len(s) >= len(sub) && indexOf(s, sub) >= 0 }

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// leafFiles returns non-dir entries under harness/<dir>, ignoring nested
// dirs (so an emptied dir with leftover subdirs still reads as drained).
func leafFiles(harnessAbs, dir string) []string {
	var out []string
	entries, _ := os.ReadDir(filepath.Join(harnessAbs, dir))
	for _, e := range entries {
		if !e.IsDir() {
			out = append(out, e.Name())
		}
	}
	return out
}
