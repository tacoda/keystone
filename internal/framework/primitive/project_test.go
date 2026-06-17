package primitive

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProjectionRelPath(t *testing.T) {
	cases := []struct {
		kind, id string
		want     string
	}{
		{"skill", "keystone:index", filepath.Join(".claude", "skills", "keystone-index", "SKILL.md")},
		{"subagent", "code-reviewer", filepath.Join(".claude", "agents", "code-reviewer.md")},
		{"command", "review", filepath.Join(".claude", "commands", "review.md")},
		// Kinds without a projection target return "".
		{"guide", "process/spec", ""},
		{"corpus", "corpus/process/spec", ""},
		{"sensor", "build", ""},
		{"action", "verify", ""},
		{"playbook", "task", ""},
		{"rule", "no-secrets", ""},
	}
	for _, c := range cases {
		p := Primitive{Frontmatter: Frontmatter{Kind: c.kind, ID: c.id}}
		got := ProjectionRelPath(p)
		if got != c.want {
			t.Errorf("ProjectionRelPath(%s/%s) = %q, want %q", c.kind, c.id, got, c.want)
		}
	}
}

func TestProject_CopiesSkillSubagentCommand(t *testing.T) {
	root := t.TempDir()

	// Seed canonical sources.
	files := map[string]string{
		".keystone/harness/skills/keystone-demo/SKILL.md": "---\nkind: skill\nid: keystone:demo\ndescription: x\ntriggers: [demo]\n---\nbody\n",
		".keystone/harness/agents/reviewer.md":            "---\nkind: subagent\nid: reviewer\ndescription: x\ntools: [Read]\n---\nbody\n",
		".keystone/harness/commands/check.md":             "---\nkind: command\nid: check\ndescription: x\n---\nbody\n",
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
