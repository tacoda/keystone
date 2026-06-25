package primitive

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInferKind(t *testing.T) {
	cases := map[string]Kind{
		".keystone/harness/rules/idioms/go/stdlib-first.md": KindRule,
		".keystone/harness/hooks/build.md":                  KindHook,
		".keystone/harness/commands/verify.md":              KindCommand,
		".keystone/harness/skills/task/SKILL.md":            KindSkill,
		".keystone/harness/agents/code-reviewer.md":         KindAgent,
		".keystone/harness/documents/adr.md":                KindDocument,
		".keystone/harness/corpus/process/spec.md":          KindCorpus,
		".keystone/harness/concerns/reads-diff.md":          KindConcern,
		".keystone/policies/acme/harness/rules/x.md":        KindRule, // policy-nested
		".keystone/harness/whatever/x.md":                   "",       // off-convention
		"README.md":                                         "",       // no harness/
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
	src := filepath.Join(root, ".keystone/harness/rules/process/spec.md")
	if err := os.MkdirAll(filepath.Dir(src), 0o755); err != nil {
		t.Fatal(err)
	}
	// No `kind:` declared — must be inferred as rule from rules/.
	body := "---\nid: rules/process/spec\ndescription: Capture intent first.\n---\n# Spec\n"
	if err := os.WriteFile(src, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	primitives, _, err := Walk(root, ".keystone/harness")
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	if len(primitives) != 1 {
		t.Fatalf("expected 1 primitive, got %d", len(primitives))
	}
	if primitives[0].Kind != "rule" {
		t.Errorf("inferred Kind = %q, want rule", primitives[0].Kind)
	}
}

// TestWalk_ExplicitKindWins — an explicit `kind:` overrides the directory
// inference (escape hatch).
func TestWalk_ExplicitKindWins(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, ".keystone/harness/rules/special.md")
	if err := os.MkdirAll(filepath.Dir(src), 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\nkind: corpus\nid: corpus/special\ndescription: x\n---\nbody\n"
	if err := os.WriteFile(src, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	primitives, _, err := Walk(root, ".keystone/harness")
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	if primitives[0].Kind != "corpus" {
		t.Errorf("explicit Kind = %q, want corpus (override)", primitives[0].Kind)
	}
}
