package primitive

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInferKind(t *testing.T) {
	cases := map[string]Kind{
		".charter/guides/idioms/go/stdlib-first.md": KindGuide,
		".charter/sensors/build.md":                 KindSensor,
		".charter/hooks/pre-verify.md":              KindHook,
		".charter/commands/verify.md":               KindCommand,
		".charter/skills/task/SKILL.md":             KindSkill,
		".charter/agents/code-reviewer.md":          KindAgent,
		".charter/playbooks/task.md":                KindPlaybook,
		".charter/patterns/retry-with-backoff.md":   KindPattern,
		".charter/posture/default.md":               KindPosture,
		".charter/tools/grep-symbols.md":            KindTool,
		".charter/documents/adr.md":                 KindDocument,
		".charter/corpus/process/spec.md":           KindCorpus,
		".charter/concerns/reads-diff.md":           KindConcern,
		".charter/evals/regression.md":              KindEval,
		// `rule` is not a kind in 3.0 — author a guide. `source` is no longer
		// a kind either (external access is a tool). Both dirs are
		// off-convention and infer nothing.
		".charter/rules/x.md":      "",
		".charter/sources/jira.md": "",
		".charter/whatever/x.md":   "", // off-convention
		"README.md":                "", // no charter/
	}
	for path, want := range cases {
		if got := InferKind(path); got != want {
			t.Errorf("InferKind(%q) = %q, want %q", path, got, want)
		}
	}
}

// TestWalk_InfersKindFromDir — a file with no `kind:` in frontmatter takes
// its kind from the canonical directory (convention over configuration).
func TestWalk_InfersKindFromDir(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, ".charter/guides/process/spec.md")
	if err := os.MkdirAll(filepath.Dir(src), 0o755); err != nil {
		t.Fatal(err)
	}
	// No `kind:` declared — must be inferred as guide from guides/.
	body := "---\nid: guides/process/spec\ndescription: Capture intent first.\n---\n# Spec\n"
	if err := os.WriteFile(src, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	primitives, _, err := Walk(root, ".charter")
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	if len(primitives) != 1 {
		t.Fatalf("expected 1 primitive, got %d", len(primitives))
	}
	if primitives[0].Kind != "guide" {
		t.Errorf("inferred Kind = %q, want guide", primitives[0].Kind)
	}
}

// TestWalk_ExplicitKindWins — an explicit `kind:` overrides the directory
// inference (escape hatch).
func TestWalk_ExplicitKindWins(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, ".charter/guides/special.md")
	if err := os.MkdirAll(filepath.Dir(src), 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\nkind: corpus\nid: corpus/special\ndescription: x\n---\nbody\n"
	if err := os.WriteFile(src, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	primitives, _, err := Walk(root, ".charter")
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	if primitives[0].Kind != "corpus" {
		t.Errorf("explicit Kind = %q, want corpus (override)", primitives[0].Kind)
	}
}
