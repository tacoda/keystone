package projection

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMergeBlock(t *testing.T) {
	block := BlockStart + "\nNEW\n" + BlockEnd
	cases := map[string]struct{ in, want string }{
		"absent":         {"", block + "\n"},
		"replace":        {"top\n\n" + BlockStart + "\nOLD\n" + BlockEnd + "\n\nbottom\n", "top\n\n" + block + "\n\nbottom\n"},
		"after H1":       {"# Title\nuser text\n", "# Title\n\n" + block + "\n\nuser text\n"},
		"no H1, prepend": {"user text\n", block + "\n\nuser text\n"},
	}
	for name, c := range cases {
		if got := MergeBlock(c.in, "NEW"); got != c.want {
			t.Errorf("%s:\n got %q\nwant %q", name, got, c.want)
		}
	}
}

func rootDests(outs []Output) string {
	var ds []string
	for _, o := range outs {
		ds = append(ds, string(o.Strategy)+" "+o.Dest)
	}
	return strings.Join(ds, "|")
}

func TestRootOutputs(t *testing.T) {
	prims := walk(t, seedCharter(t))
	if got := rootDests(RootOutputs(prims, Enabled(nil))); got != "block AGENTS.md|block CLAUDE.md" {
		t.Errorf("default hosts: %s", got)
	}
	if got := rootDests(RootOutputs(prims, Enabled([]string{"aider"}))); got != "block AGENTS.md|block CLAUDE.md|shim .aider.conf.yml" {
		t.Errorf("with aider: %s", got)
	}
}

func TestAgentsMD_Content(t *testing.T) {
	root := seedCharter(t)
	write(t, root, map[string]string{".agents/rules/process/spec.md": "---\nid: process/spec\ndescription: Capture intent first.\n---\n# Spec\n"})
	md := AgentsMD(walk(t, root), Enabled([]string{"codex", "aider"}))
	for _, want := range []string{
		"## Constraints", "**rule** (`.agents/rules/`)", "## Capabilities", "**subagent** (`.agents/agents/`)",
		"## Rules that apply everywhere", "`.agents/rules/process/spec.md` — Capture intent first.",
		"## Iron laws", "## On each host",
		"- **Codex** — loads skills, subagents by itself; open rules, commands from `.agents/` yourself",
		"- **Aider** — open rules, skills, commands, subagents from `.agents/` yourself; does **not** run hooks",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("AGENTS.md missing %q", want)
		}
	}
	if strings.Contains(md, "**Claude Code** — loads rules, hooks (inferential), skills, commands, subagents by itself; does") {
		t.Error("Claude Code bridges hooks; it must not be told to run them by hand")
	}
	if strings.Contains(md, "go/style.md") {
		t.Error("path-scoped rules belong to host rule files, not the always-loaded list")
	}
}

// TestApply_BlockKeepsUserContent — the managed block never touches text
// outside the markers, is idempotent, and Check notices an edited block.
func TestApply_BlockKeepsUserContent(t *testing.T) {
	root := seedCharter(t)
	write(t, root, map[string]string{"AGENTS.md": "# My project\n\nMy own notes.\n"})
	outs := Plan(walk(t, root), Enabled(nil)).Outputs
	first := apply(t, root, outs, nil)
	got, _ := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if !strings.HasPrefix(string(got), "# My project\n\n"+BlockStart) || !strings.HasSuffix(string(got), "My own notes.\n") {
		t.Errorf("user content not preserved around the block:\n%s", got)
	}
	if again := apply(t, root, outs, first.Owned); len(again.Wrote) != 0 {
		t.Errorf("second apply wrote %v", again.Wrote)
	}
	write(t, root, map[string]string{"AGENTS.md": strings.Replace(string(got), "## Iron laws", "## Edited", 1)})
	if d := Check(root, outs, LinkSymlink); len(d) != 1 || d[0].Dest != "AGENTS.md" {
		t.Errorf("want AGENTS.md block drift, got %v", d)
	}
}

func TestApply_AiderConfRefusedUnlessOwned(t *testing.T) {
	root := seedCharter(t)
	write(t, root, map[string]string{".aider.conf.yml": "read:\n  - MINE.md\n"})
	outs := Plan(walk(t, root), Enabled([]string{"aider"})).Outputs
	if _, err := Apply(root, outs, nil, Options{Mode: LinkSymlink}); err == nil || !strings.Contains(err.Error(), ".aider.conf.yml") {
		t.Fatalf("expected refusal naming .aider.conf.yml, got %v", err)
	}
}
