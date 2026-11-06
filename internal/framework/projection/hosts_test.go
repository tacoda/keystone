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
	root := seedCharter(t)
	got := map[string][]string{}
	for _, o := range Plan(walk(t, root), Enabled(allHosts)).Outputs {
		got[o.Host] = append(got[o.Host], string(o.Strategy)+" "+o.Dest)
	}
	want := map[string][]string{
		"claude-code": {"shim .claude/agents/keystone-reviewer.md", "shim .claude/commands/keystone-verify.md",
			"shim .claude/rules/keystone-go-style.md", "symlink .claude/skills/deploy"},
		// Cursor reads .agents/skills and .claude/agents: no skill or subagent output.
		"cursor": {"shim .cursor/rules/keystone-go-style.mdc", "shim .cursor/skills/keystone-verify/SKILL.md"},
		"github-copilot": {"shim .github/agents/keystone-reviewer.agent.md", "shim .github/instructions/keystone-go-style.instructions.md",
			"shim .github/prompts/keystone-verify.prompt.md"},
		// Continue reads .claude/skills, so the Claude Code link covers it.
		"continue": {"index ", "shim .continue/prompts/keystone-verify.md", "shim .continue/rules/keystone-go-style.md"},
		"opencode": {"index ", "shim .opencode/agents/keystone-reviewer.md", "shim .opencode/commands/keystone-verify.md"},
		"codex":    {"index ", "index ", "shim .codex/agents/keystone-reviewer.toml"},
		"cline":    {"copy .cline/skills/deploy", "index ", "index ", "shim .clinerules/keystone-go-style.md"},
		// Goose reads .claude/agents: no subagent output.
		"goose": {"index ", "index "},
		"pi":    {"index ", "index ", "shim .pi/prompts/keystone-verify.md"},
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

func TestHosts_ShimFormats(t *testing.T) {
	root := seedCharter(t)
	outs := Plan(walk(t, root), Enabled(allHosts)).Outputs
	if _, err := Apply(root, outs, nil, Options{Mode: LinkSymlink}); err != nil {
		t.Fatal(err)
	}
	for dest, wants := range map[string][]string{
		".cursor/rules/keystone-go-style.mdc":                    {"globs:\n  - '**/*.go'", "alwaysApply: false", "# Style", "- gofmt"},
		".cursor/skills/keystone-verify/SKILL.md":                {"name: keystone-verify", "disable-model-invocation: true", "# Verify"},
		".github/instructions/keystone-go-style.instructions.md": {"applyTo: '**/*.go'"},
		".github/prompts/keystone-verify.prompt.md":              {"description: Verify."},
		".continue/rules/keystone-go-style.md":                   {"name: keystone-go-style", "globs:"},
		".continue/prompts/keystone-verify.md":                   {"invokable: true"},
		".opencode/agents/keystone-reviewer.md":                  {"mode: subagent", "# Reviewer"},
		".codex/agents/keystone-reviewer.toml":                   {`name = "keystone-reviewer"`, `developer_instructions = "# Reviewer"`},
		".clinerules/keystone-go-style.md":                       {"paths:", "- gofmt"},
		".pi/prompts/keystone-verify.md":                         {"description: Verify."},
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
