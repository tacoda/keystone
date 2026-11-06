package primitive

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInferKind(t *testing.T) {
	cases := map[string]Kind{
		// 5.0 directories.
		".charter/rules/go/stdlib-first.md":   KindRule,
		".charter/references/process/spec.md": KindReference,
		".charter/hooks/build.md":             KindHook,
		".charter/permissions/default.md":     KindPermission,
		".charter/evals/regression.md":        KindEval,
		".charter/skills/task/SKILL.md":       KindSkill,
		".charter/commands/verify.md":         KindCommand,
		".charter/agents/code-reviewer.md":    KindSubagent,
		".charter/tools/grep-symbols.md":      KindTool,
		".charter/documents/adr.md":           KindDocument,
		// 4.x directories normalize to their 5.0 kinds.
		".charter/guides/idioms/go/stdlib-first.md": KindRule,
		".charter/sensors/build.md":                 KindHook,
		".charter/playbooks/task.md":                KindSkill,
		".charter/patterns/retry-with-backoff.md":   KindSkill,
		".charter/posture/default.md":               KindPermission,
		".charter/corpus/process/spec.md":           KindReference,
		".charter/concerns/reads-diff.md":           KindConcern,
		// Off-convention paths infer nothing.
		".charter/sources/jira.md": "",
		".charter/whatever/x.md":   "",
		"README.md":                "",
	}
	for path, want := range cases {
		if got := InferKind(path); got != want {
			t.Errorf("InferKind(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestNormalizeKind(t *testing.T) {
	cases := []struct {
		in     string
		want   Kind
		legacy bool
	}{
		{"guide", KindRule, true},
		{"corpus", KindReference, true},
		{"sensor", KindHook, true},
		{"posture", KindPermission, true},
		{"agent", KindSubagent, true},
		{"playbook", KindSkill, true},
		{"pattern", KindSkill, true},
		{"rule", KindRule, false},
		{"skill", KindSkill, false},
		{"bogus", Kind("bogus"), false},
	}
	for _, c := range cases {
		got, legacy := NormalizeKind(c.in)
		if got != c.want || legacy != c.legacy {
			t.Errorf("NormalizeKind(%q) = %q,%v want %q,%v", c.in, got, legacy, c.want, c.legacy)
		}
	}
}

func TestCategoryOf(t *testing.T) {
	want := map[Kind]Category{
		KindRule: CategoryConstraint, KindReference: CategoryConstraint,
		KindHook: CategoryConstraint, KindPermission: CategoryConstraint,
		KindEval:  CategoryConstraint,
		KindSkill: CategoryCapability, KindCommand: CategoryCapability,
		KindSubagent: CategoryCapability, KindTool: CategoryCapability,
		KindDocument: CategoryCapability,
		KindConcern:  "", // retired: no category
	}
	for k, c := range want {
		if got := CategoryOf(k); got != c {
			t.Errorf("CategoryOf(%q) = %q, want %q", k, got, c)
		}
	}
	if len(want)-1 != len(KnownKinds) {
		t.Errorf("KnownKinds has %d kinds, table covers %d", len(KnownKinds), len(want)-1)
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
	// No `kind:` declared — must be inferred as rule from the 4.x guides/ dir.
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
	if primitives[0].Kind != "rule" {
		t.Errorf("inferred Kind = %q, want rule (4.x guides/ dir)", primitives[0].Kind)
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
	// The explicit 4.x `kind: corpus` wins over the dir and normalizes.
	if primitives[0].Kind != "reference" {
		t.Errorf("explicit Kind = %q, want reference (override, normalized)", primitives[0].Kind)
	}
}
