package primitive

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProjectionRelPath(t *testing.T) {
	cases := []struct {
		kind, id string
		globs    []string
		want     string
	}{
		// Agent escape hatches.
		{"skill", "keystone:index", nil, filepath.Join(".claude", "skills", "keystone-index", "SKILL.md")},
		{"subagent", "code-reviewer", nil, filepath.Join(".claude", "agents", "code-reviewer.md")},
		{"command", "review", nil, filepath.Join(".claude", "commands", "review.md")},
		// Framework wrappers project to the same host paths as their
		// agent counterparts.
		{"persona", "security-reviewer", nil, filepath.Join(".claude", "agents", "security-reviewer.md")},
		{"action", "verify", nil, filepath.Join(".claude", "commands", "verify.md")},
		{"playbook", "task", nil, filepath.Join(".claude", "skills", "task", "SKILL.md")},
		// Idiom guide with globs → rule shim.
		{"guide", "guides/idioms/go/stdlib-first", []string{"cmd/**/*.go"}, filepath.Join(".claude", "rules", "go-stdlib-first.md")},
		{"guide", "guides/idioms/harness-content/state-files", []string{"x/**"}, filepath.Join(".claude", "rules", "harness-content-state-files.md")},
		// Guide without globs → no projection.
		{"guide", "process/spec", nil, ""},
		// Other no-projection kinds.
		{"corpus", "corpus/process/spec", nil, ""},
		{"sensor", "build", nil, ""},
		{"rule", "no-secrets", nil, ""},
		{"eval", "demo", nil, ""},
		{"source", "docs", nil, ""},
	}
	for _, c := range cases {
		p := Primitive{Frontmatter: Frontmatter{Kind: c.kind, ID: c.id, Globs: c.globs}}
		got := ProjectionRelPath(p)
		if got != c.want {
			t.Errorf("ProjectionRelPath(%s/%s, globs=%v) = %q, want %q", c.kind, c.id, c.globs, got, c.want)
		}
	}
}

func TestProject_GuideShim(t *testing.T) {
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
		if r.Action == "wrote" && r.Dest == ".claude/rules/go-stdlib-first.md" {
			wrote = true
		}
	}
	if !wrote {
		t.Fatal("expected shim projection")
	}

	out, err := os.ReadFile(filepath.Join(root, ".claude/rules/go-stdlib-first.md"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	for _, want := range []string{
		`kind: rule`,
		`id: rules/go-stdlib-first`,
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
	} {
		if !strings.Contains(s, want) {
			t.Errorf("shim missing %q\nshim:\n%s", want, s)
		}
	}
	for _, banned := range []string{
		"Intro prose",
		"Why this is agent-specific",
		"See also",
		"[[deps]]",
	} {
		if strings.Contains(s, banned) {
			t.Errorf("shim leaked excluded content %q", banned)
		}
	}
}

func TestProject_GuideWithoutGlobsSkipped(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, ".keystone/harness/guides/process/spec.md")
	if err := os.MkdirAll(filepath.Dir(src), 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\nkind: guide\nid: process/spec\ndescription: x\n---\n# Spec\n\n## IRON LAW\n\n- One.\n"
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

func TestProject_CopiesSkillSubagentCommand(t *testing.T) {
	root := t.TempDir()

	// Seed canonical sources.
	files := map[string]string{
		// Agent escape hatches.
		".keystone/harness/skills/keystone-demo/SKILL.md": "---\nkind: skill\nid: keystone:demo\ndescription: x\ntriggers: [demo]\n---\nbody\n",
		".keystone/harness/agents/reviewer.md":            "---\nkind: subagent\nid: reviewer\ndescription: x\ntools: [Read]\n---\nbody\n",
		".keystone/harness/commands/check.md":             "---\nkind: command\nid: check\ndescription: x\n---\nbody\n",
		// Framework wrappers.
		".keystone/harness/personas/security.md":  "---\nkind: persona\nid: security\ndescription: x\ntools: [Read]\n---\nbody\n",
		".keystone/harness/actions/ship.md":       "---\nkind: action\nid: ship\ndescription: x\n---\nbody\n",
		".keystone/harness/playbooks/onboard.md":  "---\nkind: playbook\nid: onboard\ndescription: x\ntriggers: [onboard]\n---\nbody\n",
		// Source that has no projection target — should be ignored without error.
		".keystone/harness/guides/process/spec.md": "---\nkind: guide\nid: process/spec\ndescription: x\n---\nbody\n",
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

	primitives, _, err := Walk(root, ".keystone/harness")
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	results, err := Project(root, primitives)
	if err != nil {
		t.Fatalf("Project: %v", err)
	}

	wrote := map[string]bool{}
	skipped := 0
	for _, r := range results {
		if r.Action == "wrote" {
			wrote[r.Dest] = true
		} else if r.Action == "skipped-no-projection" {
			skipped++
		}
	}
	if !wrote[".claude/skills/keystone-demo/SKILL.md"] {
		t.Error("expected skill projection")
	}
	if !wrote[".claude/agents/reviewer.md"] {
		t.Error("expected subagent projection")
	}
	if !wrote[".claude/commands/check.md"] {
		t.Error("expected command projection")
	}
	if !wrote[".claude/agents/security.md"] {
		t.Error("expected persona projection")
	}
	if !wrote[".claude/commands/ship.md"] {
		t.Error("expected action projection")
	}
	if !wrote[".claude/skills/onboard/SKILL.md"] {
		t.Error("expected playbook projection")
	}
	if skipped == 0 {
		t.Error("expected at least one skipped (the guide)")
	}

	// Confirm body integrity — projection should be byte-identical.
	for src, dest := range map[string]string{
		".keystone/harness/skills/keystone-demo/SKILL.md": ".claude/skills/keystone-demo/SKILL.md",
		".keystone/harness/agents/reviewer.md":            ".claude/agents/reviewer.md",
		".keystone/harness/commands/check.md":             ".claude/commands/check.md",
	} {
		a, _ := os.ReadFile(filepath.Join(root, src))
		b, err := os.ReadFile(filepath.Join(root, dest))
		if err != nil {
			t.Errorf("read projection %s: %v", dest, err)
			continue
		}
		if string(a) != string(b) {
			t.Errorf("projection %s differs from source %s", dest, src)
		}
	}
}

func TestProject_OverwritesHandEdits(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, ".keystone/harness/skills/x/SKILL.md")
	if err := os.MkdirAll(filepath.Dir(src), 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\nkind: skill\nid: x\ndescription: x\ntriggers: [x]\n---\nfresh body\n"
	if err := os.WriteFile(src, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	// Hand-edit the projection target before running Project.
	dest := filepath.Join(root, ".claude/skills/x/SKILL.md")
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dest, []byte("STALE HAND EDIT"), 0o644); err != nil {
		t.Fatal(err)
	}

	primitives, _, _ := Walk(root, ".keystone/harness")
	if _, err := Project(root, primitives); err != nil {
		t.Fatalf("Project: %v", err)
	}
	got, _ := os.ReadFile(dest)
	if string(got) != body {
		t.Errorf("projection not overwritten; got %q", string(got))
	}
}
