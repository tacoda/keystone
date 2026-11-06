package projection

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

var allHosts = []string{"cursor", "github-copilot", "continue", "opencode", "codex", "cline", "goose", "pi", "aider"}

// TestHosts_PlanPerHost pins every host's outputs for the fixture charter
// (one skill, one path-scoped rule, one command, one subagent).
func TestHosts_PlanPerHost(t *testing.T) {
	got := outputsByHost(plan(t, seedCharter(t), Enabled(allHosts)...))
	want := map[string][]string{
		"claude-code": {"shim .claude/agents/reviewer.md", "shim .claude/commands/verify.md",
			"shim .claude/rules/go-style.md", "symlink .claude/skills/deploy"},
		// Cursor reads .agents/skills and .claude/agents: no skill or subagent output.
		"cursor": {"shim .cursor/rules/go-style.mdc", "shim .cursor/skills/verify/SKILL.md"},
		"github-copilot": {"shim .github/agents/reviewer.agent.md", "shim .github/instructions/go-style.instructions.md",
			"shim .github/prompts/verify.prompt.md"},
		// Continue reads .claude/skills, so the Claude Code link covers it.
		"continue": {"index ", "shim .continue/prompts/verify.md", "shim .continue/rules/go-style.md"},
		"opencode": {"index ", "shim .opencode/agents/reviewer.md", "shim .opencode/commands/verify.md"},
		"codex":    {"index ", "index ", "shim .codex/agents/reviewer.toml"},
		"cline":    {"copy .cline/skills/deploy", "index ", "index ", "shim .clinerules/go-style.md"},
		// Goose reads .claude/agents: no subagent output.
		"goose": {"index ", "index "},
		"pi":    {"index ", "index ", "shim .pi/prompts/verify.md"},
		"aider": {"index ", "index ", "index ", "index "},
	}
	for host, w := range want {
		g := got[host]
		sort.Strings(g)
		sort.Strings(w)
		if strings.Join(g, "|") != strings.Join(w, "|") {
			t.Errorf("%s:\n  got  %v\n  want %v", host, g, w)
		}
	}
}

func outputsByHost(outs []Output) map[string][]string {
	got := map[string][]string{}
	for _, o := range outs {
		got[o.Host] = append(got[o.Host], string(o.Strategy)+" "+o.Dest)
	}
	return got
}

func TestHosts_ShimFormats(t *testing.T) {
	root := seedCharter(t)
	outs := Plan(walk(t, root), Enabled(allHosts)).Outputs
	if _, err := Apply(root, outs, nil, Options{Mode: LinkSymlink}); err != nil {
		t.Fatal(err)
	}
	for dest, wants := range map[string][]string{
		".cursor/rules/go-style.mdc":                    {"globs:\n  - '**/*.go'", "alwaysApply: false", "# Style", "- gofmt"},
		".cursor/skills/verify/SKILL.md":                {"name: verify", "disable-model-invocation: true", "# Verify"},
		".github/instructions/go-style.instructions.md": {"applyTo: '**/*.go'"},
		".github/prompts/verify.prompt.md":              {"description: Verify."},
		".continue/rules/go-style.md":                   {"name: go-style", "globs:"},
		".continue/prompts/verify.md":                   {"invokable: true"},
		".opencode/agents/reviewer.md":                  {"mode: subagent", "# Reviewer"},
		".codex/agents/reviewer.toml":                   {`name = "reviewer"`, `developer_instructions = "# Reviewer"`},
		".clinerules/go-style.md":                       {"paths:", "- gofmt"},
		".pi/prompts/verify.md":                         {"description: Verify."},
		".cline/skills/deploy/scripts/run.sh":                    {"echo run"},
	} {
		data, err := os.ReadFile(filepath.Join(root, dest))
		if err != nil {
			t.Errorf("%s: %v", dest, err)
			continue
		}
		for _, w := range wants {
			if !strings.Contains(string(data), w) {
				t.Errorf("%s missing %q:\n%s", dest, w, data)
			}
		}
	}
	if d := Check(root, outs, LinkSymlink); len(d) != 0 {
		t.Errorf("fresh all-host projection drifted: %v", d)
	}
}
