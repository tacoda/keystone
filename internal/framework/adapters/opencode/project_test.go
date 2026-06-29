package opencode

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

// writeSource writes a minimal primitive source file under projectDir so
// RenderForHost has bytes to read, and returns the primitive pointing at it.
func writeSource(t *testing.T, root, kind, id, rel, body string) primitive.Primitive {
	t.Helper()
	abs := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatal(err)
	}
	src := "---\nkind: " + kind + "\nid: " + id + "\ndescription: d\n---\n" + body
	if err := os.WriteFile(abs, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	return primitive.Primitive{
		Frontmatter: primitive.Frontmatter{Kind: kind, ID: id, Description: "d"},
		Path:        rel,
	}
}

func TestProjectAgents_MirrorsSkillsAgentsCommands(t *testing.T) {
	root := t.TempDir()
	prims := []primitive.Primitive{
		writeSource(t, root, "skill", "foo", ".harness/skills/foo/SKILL.md", "skill body\n"),
		writeSource(t, root, "command", "verify", ".harness/commands/verify.md", "command body\n"),
		writeSource(t, root, "agent", "reviewer", ".harness/agents/reviewer.md", "agent body\n"),
	}
	// A globbed guide projects to .opencode/rules/ as a rule-shim.
	guide := writeSource(t, root, "guide", "guides/idioms/go/stdlib-first",
		".harness/guides/idioms/go/stdlib-first.md", "# Stdlib first\n\n## IRON LAW\n- x\n")
	guide.Globs = []string{"**/*.go"}
	prims = append(prims, guide)

	res, err := ProjectAgents(root, prims)
	if err != nil {
		t.Fatalf("ProjectAgents: %v", err)
	}
	if res.Wrote != 4 {
		t.Errorf("expected 4 written, got %d", res.Wrote)
	}
	if res.Rules != 1 {
		t.Errorf("expected 1 rule, got %d", res.Rules)
	}

	for _, want := range []string{
		"skills/keystone-foo/SKILL.md",
		"commands/keystone-verify.md",
		"agents/keystone-reviewer.md",
		"rules/keystone-go-stdlib-first.md",
	} {
		if _, err := os.Stat(filepath.Join(root, Root, want)); err != nil {
			t.Errorf("missing projection %s: %v", want, err)
		}
	}

	// The rules glob must be registered in opencode.json instructions.
	raw, err := os.ReadFile(filepath.Join(root, "opencode.json"))
	if err != nil {
		t.Fatalf("opencode.json not written: %v", err)
	}
	if !strings.Contains(string(raw), RulesGlob) {
		t.Errorf("opencode.json missing instructions glob %q\n%s", RulesGlob, raw)
	}
}

func TestProjectAgents_IsIdempotent(t *testing.T) {
	root := t.TempDir()
	prims := []primitive.Primitive{
		writeSource(t, root, "skill", "foo", ".harness/skills/foo/SKILL.md", "body\n"),
	}
	if _, err := ProjectAgents(root, prims); err != nil {
		t.Fatal(err)
	}
	res2, err := ProjectAgents(root, prims)
	if err != nil {
		t.Fatal(err)
	}
	if res2.Wrote != 0 || res2.Unchanged != 1 {
		t.Errorf("expected idempotent unchanged, got wrote=%d unchanged=%d", res2.Wrote, res2.Unchanged)
	}
}

func TestEnsureInstructions_PreservesAndDeduplicates(t *testing.T) {
	root := t.TempDir()
	existing := `{"$schema":"x","mcp":{"keystone":{}},"instructions":["CONTRIBUTING.md"]}`
	if err := os.WriteFile(filepath.Join(root, "opencode.json"), []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ensureInstructions(root, RulesGlob); err != nil {
		t.Fatal(err)
	}
	// Second call is a no-op (no duplicate entry).
	if err := ensureInstructions(root, RulesGlob); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(root, "opencode.json"))
	s := string(raw)
	if !strings.Contains(s, "CONTRIBUTING.md") {
		t.Errorf("existing instruction dropped:\n%s", s)
	}
	if !strings.Contains(s, `"mcp"`) {
		t.Errorf("existing mcp key dropped:\n%s", s)
	}
	if strings.Count(s, RulesGlob) != 1 {
		t.Errorf("glob should appear exactly once, got %d:\n%s", strings.Count(s, RulesGlob), s)
	}
}
