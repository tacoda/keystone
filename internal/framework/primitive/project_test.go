package primitive

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// seedFiles writes a set of rel-path → body fixtures under root, creating
// parent directories. Shared by the projection integration tests.
func seedFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, body := range files {
		abs := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestProjectionRelPath(t *testing.T) {
	cases := []struct {
		kind, id, mode string
		globs          []string
		want           string
	}{
		// Projected names are kebab-case + keystone-prefixed; an id already
		// in the keystone namespace is not double-prefixed.
		{"skill", "keystone:index", "", nil, filepath.Join(".claude", "skills", "keystone-index", "SKILL.md")},
		{"agent", "code-reviewer", "", nil, filepath.Join(".claude", "agents", "keystone-code-reviewer.md")},
		{"command", "review", "", nil, filepath.Join(".claude", "commands", "keystone-review.md")},
		{"command", "verify", "", nil, filepath.Join(".claude", "commands", "keystone-verify.md")},
		{"skill", "task", "", nil, filepath.Join(".claude", "skills", "keystone-task", "SKILL.md")},
		// playbook projects like skill — a composed SKILL.md orchestrator.
		{"playbook", "task", "", nil, filepath.Join(".claude", "skills", "keystone-task", "SKILL.md")},
		// Glob-scoped inferential guide → rule shim (empty mode defaults inferential).
		{"guide", "guides/idioms/go/stdlib-first", "", []string{"cmd/**/*.go"}, filepath.Join(".claude", "rules", "keystone-go-stdlib-first.md")},
		{"guide", "guides/idioms/harness-content/state-files", "inferential", []string{"x/**"}, filepath.Join(".claude", "rules", "keystone-harness-content-state-files.md")},
		// Computational guide → no shim (author a hook instead).
		{"guide", "guides/idioms/go/fmt", "computational", []string{"**/*.go"}, ""},
		// Guide without globs → no projection.
		{"guide", "guides/process/spec", "", nil, ""},
		// Inferential sensor (review) → subagent; computational sensor fires
		// via the hook layer (no adapter file).
		{"sensor", "review-functional", "inferential", nil, filepath.Join(".claude", "agents", "keystone-review-functional.md")},
		{"sensor", "review-security", "", nil, filepath.Join(".claude", "agents", "keystone-review-security.md")},
		{"sensor", "build", "computational", nil, ""},
		// Other no-projection kinds.
		{"corpus", "corpus/process/spec", "", nil, ""},
		{"hook", "on-gate", "", nil, ""},
		{"pattern", "repository", "", nil, ""},
		{"posture", "default", "", nil, ""},
		{"tool", "grep-symbols", "", nil, ""},
		{"document", "implementation-plan", "", nil, ""},
		{"eval", "demo", "", nil, ""},
	}
	for _, c := range cases {
		p := Primitive{Frontmatter: Frontmatter{Kind: c.kind, ID: c.id, Mode: c.mode, Globs: c.globs}}
		got := ProjectionRelPath(p)
		if got != c.want {
			t.Errorf("ProjectionRelPath(%s/%s, mode=%q, globs=%v) = %q, want %q", c.kind, c.id, c.mode, c.globs, got, c.want)
		}
	}
}

func TestProject_RuleShim(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, ".keystone/harness/guides/idioms/go/stdlib-first.md")
	if err := os.MkdirAll(filepath.Dir(src), 0o755); err != nil {
		t.Fatal(err)
	}
	body := `---
kind: guide
id: guides/idioms/go/stdlib-first
description: Prefer Go stdlib over new deps.
globs:
  - "cmd/**/*.go"
  - "go.mod"
---
# Stdlib first

Intro prose that should not appear in the shim.

## IRON LAW

- No new direct dep without naming what stdlib it replaces.

## GOLDEN RULE

- Filesystem → io/fs, os.

## Why this is agent-specific

This prose section must be excluded from the shim.

## Anti-patterns

- Two libraries doing the same job.

## See also

[[deps]]
`
	if err := os.WriteFile(src, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	primitives, _, err := Walk(root, ".keystone/harness")
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	results, err := Project(root, primitives)
	if err != nil {
		t.Fatalf("Project: %v", err)
	}
	wrote := false
	for _, r := range results {
		if r.Action == "wrote" && r.Dest == ".claude/rules/keystone-go-stdlib-first.md" {
			wrote = true
		}
	}
	if !wrote {
		t.Fatal("expected shim projection")
	}

	out, err := os.ReadFile(filepath.Join(root, ".claude/rules/keystone-go-stdlib-first.md"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	assertContainsAll(t, s, []string{
		`kind: rule`,
		`id: rules/keystone-go-stdlib-first`,
		`description: Prefer Go stdlib over new deps.`,
		`globs:`,
		`  - "cmd/**/*.go"`,
		`  - "go.mod"`,
		`source: .keystone/harness/guides/idioms/go/stdlib-first.md`,
		`generated_by: keystone-project`,
		`# Stdlib first`,
		`## IRON LAW`,
		`## GOLDEN RULE`,
		`## Anti-patterns`,
		`No new direct dep`,
	})
	assertContainsNone(t, s, []string{
		"Intro prose",
		"Why this is agent-specific",
		"See also",
		"[[deps]]",
	})
}

func TestProject_RuleWithoutGlobsSkipped(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, ".keystone/harness/guides/process/spec.md")
	if err := os.MkdirAll(filepath.Dir(src), 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\nkind: guide\nid: guides/process/spec\ndescription: x\n---\n# Spec\n\n## IRON LAW\n\n- One.\n"
	if err := os.WriteFile(src, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	primitives, _, _ := Walk(root, ".keystone/harness")
	results, err := Project(root, primitives)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range results {
		if r.Kind == "guide" && r.Action == "wrote" {
			t.Errorf("expected globless guide to be skipped, got dest=%q", r.Dest)
		}
	}
	if _, err := os.Stat(filepath.Join(root, ".claude/rules")); !os.IsNotExist(err) {
		t.Errorf(".claude/rules created for globless guide: err=%v", err)
	}
}

func TestProject_CopiesSkillAgentCommand(t *testing.T) {
	root := t.TempDir()

	// Seed canonical sources across the projecting kinds.
	files := map[string]string{
		".keystone/harness/skills/keystone-demo/SKILL.md": "---\nkind: skill\nid: keystone:demo\ndescription: x\ntriggers: [demo]\n---\nbody\n",
		".keystone/harness/agents/reviewer.md":            "---\nkind: agent\nid: reviewer\ndescription: x\ntools: [Read]\n---\nbody\n",
		".keystone/harness/commands/check.md":             "---\nkind: command\nid: check\ndescription: x\n---\nbody\n",
		".keystone/harness/agents/security.md":            "---\nkind: agent\nid: security\ndescription: x\ntools: [Read]\n---\nbody\n",
		".keystone/harness/commands/ship.md":              "---\nkind: command\nid: ship\ndescription: x\n---\nbody\n",
		".keystone/harness/skills/onboard/SKILL.md":       "---\nkind: skill\nid: onboard\ndescription: x\ntriggers: [onboard]\n---\nbody\n",
		// Source that has no projection target — should be ignored without error.
		".keystone/harness/guides/process/spec.md": "---\nkind: guide\nid: guides/process/spec\ndescription: x\n---\nbody\n",
	}
	seedFiles(t, root, files)

	primitives, _, err := Walk(root, ".keystone/harness")
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	results, err := Project(root, primitives)
	if err != nil {
		t.Fatalf("Project: %v", err)
	}

	wrote, skipped := tallyProjections(results)
	assertAllWrote(t, wrote, []string{
		".claude/skills/keystone-demo/SKILL.md",
		".claude/agents/keystone-reviewer.md",
		".claude/commands/keystone-check.md",
		".claude/agents/keystone-security.md",
		".claude/commands/keystone-ship.md",
		".claude/skills/keystone-onboard/SKILL.md",
	})
	if skipped == 0 {
		t.Error("expected at least one skipped (the globless guide)")
	}
	// Frontmatter is lowered to host-native keys; the body is preserved.
	// Skills/agents carry name+description; keystone-only fields are stripped.
	skill, _ := os.ReadFile(filepath.Join(root, ".claude/skills/keystone-demo/SKILL.md"))
	assertContainsAll(t, string(skill), []string{"name: keystone-demo", "description: x", "body"})
	assertContainsNone(t, string(skill), []string{"kind:", "id:", "triggers:"})
	cmd, _ := os.ReadFile(filepath.Join(root, ".claude/commands/keystone-check.md"))
	assertContainsAll(t, string(cmd), []string{"description: x", "body"})
	assertContainsNone(t, string(cmd), []string{"kind:", "id:"})
}

// tallyProjections splits results into a wrote-set and a skipped count.
func tallyProjections(results []ProjectionResult) (map[string]bool, int) {
	wrote := map[string]bool{}
	skipped := 0
	for _, r := range results {
		switch r.Action {
		case "wrote":
			wrote[r.Dest] = true
		case "skipped-no-projection":
			skipped++
		}
	}
	return wrote, skipped
}

func assertAllWrote(t *testing.T, wrote map[string]bool, dests []string) {
	t.Helper()
	for _, dest := range dests {
		if !wrote[dest] {
			t.Errorf("expected projection %s", dest)
		}
	}
}

func assertContainsAll(t *testing.T, s string, wants []string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(s, want) {
			t.Errorf("missing %q\nin:\n%s", want, s)
		}
	}
}

func assertContainsNone(t *testing.T, s string, banned []string) {
	t.Helper()
	for _, b := range banned {
		if strings.Contains(s, b) {
			t.Errorf("leaked excluded content %q", b)
		}
	}
}

// TestProject_TypeAware — slice 2: an inferential sensor projects to a
// subagent file, a playbook projects to a SKILL.md, and a computational guide
// is NOT shimmed (a hook would carry it instead).
func TestProject_TypeAware(t *testing.T) {
	root := t.TempDir()
	seedFiles(t, root, map[string]string{
		".keystone/harness/sensors/review-functional.md": "---\nkind: sensor\nid: review-functional\ndescription: x\nmode: inferential\nagent: reviewer\nreturns: findings\n---\n# functional review\n",
		".keystone/harness/playbooks/task.md":            "---\nkind: playbook\nid: task\ndescription: x\n---\n# task orchestrator\n",
		".keystone/harness/guides/idioms/go/fmt.md":      "---\nkind: guide\nid: guides/idioms/go/fmt\ndescription: x\nmode: computational\nglobs:\n  - \"**/*.go\"\n---\n# fmt\n",
	})
	primitives, _, err := Walk(root, ".keystone/harness")
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	if _, err := Project(root, primitives); err != nil {
		t.Fatalf("Project: %v", err)
	}
	// Inferential sensor → subagent file (keystone-prefixed).
	if _, err := os.Stat(filepath.Join(root, ".claude/agents/keystone-review-functional.md")); err != nil {
		t.Errorf("inferential sensor not projected to .claude/agents/: %v", err)
	}
	// Playbook → SKILL.md (keystone-prefixed).
	if _, err := os.Stat(filepath.Join(root, ".claude/skills/keystone-task/SKILL.md")); err != nil {
		t.Errorf("playbook not projected to SKILL.md: %v", err)
	}
	// Computational guide → no rule shim.
	if _, err := os.Stat(filepath.Join(root, ".claude/rules")); !os.IsNotExist(err) {
		t.Errorf("computational guide must not produce a rule shim: err=%v", err)
	}
}

func TestProject_OverwritesHandEdits(t *testing.T) {
	root := t.TempDir()
	seedFiles(t, root, map[string]string{
		".keystone/harness/skills/x/SKILL.md": "---\nkind: skill\nid: x\ndescription: x\ntriggers: [x]\n---\nfresh body\n",
		".claude/skills/keystone-x/SKILL.md":  "STALE HAND EDIT", // pre-existing hand-edit
	})
	primitives, _, _ := Walk(root, ".keystone/harness")
	if _, err := Project(root, primitives); err != nil {
		t.Fatalf("Project: %v", err)
	}
	got, _ := os.ReadFile(filepath.Join(root, ".claude/skills/keystone-x/SKILL.md"))
	assertContainsNone(t, string(got), []string{"STALE HAND EDIT"})
	assertContainsAll(t, string(got), []string{"name: keystone-x", "fresh body"})
}
