package projection

import (
	"path"
	"strings"

	"github.com/tacoda/keystone/internal/framework/primitive"
)

// Host profiles, one per supported host (docs/hosts.md). Hooks and
// permissions are projected separately (Phase 5); MCP registration stays
// with `keystone mcp install`.

const (
	skillK    = primitive.KindSkill
	ruleK     = primitive.KindRule
	commandK  = primitive.KindCommand
	subagentK = primitive.KindSubagent
	hookK     = primitive.KindHook
)

var (
	index  = Rule{Strategy: Index}
	native = Rule{Strategy: Native}
)

// ClaudeCode: skills are linked per entry (Claude Code reads the Agent
// Skills format and follows symlinked skill dirs, but not .agents/skills/).
// Rules, commands, and subagents differ in format (rules read only
// `paths:`; subagents need `name:`; commands use `argument-hint:`), so they
// get the Claude Code shims.
func ClaudeCode() Profile {
	shim := Rule{Strategy: Shim, Dest: primitive.ProjectionRelPath}
	return Profile{
		Host:  "claude-code",
		Hooks: claudeHooks,
		Rules: map[primitive.Kind]Rule{
			skillK:    {Strategy: Symlink, Dest: skillDirDest(".claude/skills"), Target: skillDir},
			ruleK:     shim,
			commandK:  shim,
			subagentK: shim,
			hookK:     shim, // inferential hooks only (Dest is "" otherwise)
		},
		Reads: map[primitive.Kind][]string{skillK: {".claude/skills"}},
	}
}

type cursorRule struct {
	Description string   `yaml:"description"`
	Globs       []string `yaml:"globs"`
	AlwaysApply bool     `yaml:"alwaysApply"`
}

type namedDoc struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Model       string `yaml:"model,omitempty"`
}

type userSkill struct {
	Name                   string `yaml:"name"`
	Description            string `yaml:"description"`
	DisableModelInvocation bool   `yaml:"disable-model-invocation"`
}

// Cursor reads .agents/skills/ and .claude/agents/ natively. Plain .md in
// .cursor/rules/ is ignored, so rules are .mdc. Commands are being folded
// into skills, so a command projects as a user-invoked skill.
func Cursor() Profile {
	return Profile{
		Host:  "cursor",
		Hooks: cursorHooks,
		Rules: map[primitive.Kind]Rule{
			skillK: native,
			ruleK: {Strategy: Shim, Dest: pathRuleDest(".cursor/rules", ".mdc"), Render: ruleRenderer(func(p primitive.Primitive) any {
				return cursorRule{Description: p.Description, Globs: p.Paths}
			})},
			commandK: {Strategy: Shim, Dest: nameDest(".cursor/skills", "/SKILL.md"), Render: bodyRenderer(func(p primitive.Primitive) any {
				return userSkill{Name: primitive.ProjectedName(p.ID), Description: p.Description, DisableModelInvocation: true}
			})},
			subagentK: {Strategy: Shim, Dest: nameDest(".cursor/agents", ".md"), Render: bodyRenderer(func(p primitive.Primitive) any {
				return namedDoc{Name: primitive.ProjectedName(p.ID), Description: p.Description, Model: p.Model}
			})},
		},
		Reads: map[primitive.Kind][]string{
			skillK:    {".agents/skills", ".cursor/skills", ".claude/skills", ".codex/skills"},
			subagentK: {".cursor/agents", ".claude/agents", ".codex/agents"},
		},
	}
}

type copilotInstruction struct {
	ApplyTo     string `yaml:"applyTo"`
	Description string `yaml:"description,omitempty"`
}

type promptFile struct {
	Description  string `yaml:"description"`
	ArgumentHint string `yaml:"argument-hint,omitempty"`
}

// GitHubCopilot reads .agents/skills/ natively; rules become path-scoped
// instruction files, commands prompt files, subagents custom agents.
func GitHubCopilot() Profile {
	return Profile{
		Host:  "github-copilot",
		Hooks: copilotHooks,
		Rules: map[primitive.Kind]Rule{
			skillK: native,
			ruleK: {Strategy: Shim, Dest: pathRuleDest(".github/instructions", ".instructions.md"), Render: ruleRenderer(func(p primitive.Primitive) any {
				return copilotInstruction{ApplyTo: strings.Join(p.Paths, ","), Description: p.Description}
			})},
			commandK: {Strategy: Shim, Dest: nameDest(".github/prompts", ".prompt.md"), Render: bodyRenderer(func(p primitive.Primitive) any {
				return promptFile{Description: p.Description, ArgumentHint: argumentHint(p.Args)}
			})},
			subagentK: {Strategy: Shim, Dest: nameDest(".github/agents", ".agent.md"), Render: bodyRenderer(func(p primitive.Primitive) any {
				return namedDoc{Name: primitive.ProjectedName(p.ID), Description: p.Description}
			})},
		},
		Reads: map[primitive.Kind][]string{skillK: {".agents/skills", ".github/skills", ".claude/skills"}},
	}
}

type continueRule struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Globs       []string `yaml:"globs"`
	AlwaysApply bool     `yaml:"alwaysApply"`
}

type continuePrompt struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Invokable   bool   `yaml:"invokable"`
}

// Continue does not scan .agents/skills/; the IDE reads .claude/skills/
// and the CLI also .continue/skills/ (and skips symlinked dirs), so skills
// are copied. No subagent feature was found, so subagents are indexed.
func Continue() Profile {
	return Profile{
		Host: "continue",
		Rules: map[primitive.Kind]Rule{
			skillK: {Strategy: Copy, Dest: skillDirDest(".continue/skills"), Target: skillDir},
			ruleK: {Strategy: Shim, Dest: pathRuleDest(".continue/rules", ".md"), Render: ruleRenderer(func(p primitive.Primitive) any {
				return continueRule{Name: primitive.RuleSlug(p.ID), Description: p.Description, Globs: p.Paths}
			})},
			commandK: {Strategy: Shim, Dest: nameDest(".continue/prompts", ".md"), Render: bodyRenderer(func(p primitive.Primitive) any {
				return continuePrompt{Name: primitive.ProjectedName(p.ID), Description: p.Description, Invokable: true}
			})},
			subagentK: index,
		},
		Reads: map[primitive.Kind][]string{skillK: {".claude/skills", ".continue/skills"}},
	}
}

type opencodeCommand struct {
	Description string `yaml:"description"`
	Model       string `yaml:"model,omitempty"`
}

type opencodeAgent struct {
	Description string `yaml:"description"`
	Mode        string `yaml:"mode"`
	Model       string `yaml:"model,omitempty"`
}

// OpenCode reads .agents/skills/ natively. Its `instructions` load every
// file on every turn (no path scoping), so rules are indexed from AGENTS.md
// instead of costing always-loaded context.
func OpenCode() Profile {
	return Profile{
		Host:  "opencode",
		Hooks: opencodeHooks,
		Rules: map[primitive.Kind]Rule{
			skillK: native,
			ruleK:  index,
			commandK: {Strategy: Shim, Dest: nameDest(".opencode/commands", ".md"), Render: bodyRenderer(func(p primitive.Primitive) any {
				return opencodeCommand{Description: p.Description, Model: p.Model}
			})},
			subagentK: {Strategy: Shim, Dest: nameDest(".opencode/agents", ".md"), Render: bodyRenderer(func(p primitive.Primitive) any {
				return opencodeAgent{Description: p.Description, Mode: "subagent", Model: p.Model}
			})},
		},
		Reads: map[primitive.Kind][]string{skillK: {".agents/skills", ".opencode/skills", ".claude/skills"}},
	}
}

// Codex reads AGENTS.md and .agents/skills/ natively. It has no repo-level
// commands (custom prompts are user-only and deprecated) and no path-scoped
// rules, so both are indexed. Subagents are TOML.
func Codex() Profile {
	return Profile{
		Host: "codex",
		Rules: map[primitive.Kind]Rule{
			skillK:    native,
			ruleK:     index,
			commandK:  index,
			subagentK: {Strategy: Shim, Dest: nameDest(".codex/agents", ".toml"), Render: codexAgent},
		},
		Reads: map[primitive.Kind][]string{skillK: {".agents/skills"}},
	}
}

// Cline reads `paths:` rules (the Claude Code shim format) from
// .clinerules/. Its symlink support is unverified, so skills are copied;
// its workflow and subagent formats are unverified, so both are indexed.
func Cline() Profile {
	return Profile{
		Host: "cline",
		Rules: map[primitive.Kind]Rule{
			skillK:    {Strategy: Copy, Dest: skillDirDest(".cline/skills"), Target: skillDir},
			ruleK:     {Strategy: Shim, Dest: pathRuleDest(".clinerules", ".md")},
			commandK:  index,
			subagentK: index,
		},
		Reads: map[primitive.Kind][]string{skillK: {".cline/skills"}},
	}
}

// Goose reads .agents/skills/ natively and subagents from .goose/agents/
// and .claude/agents/ (canonical .agents/agents/ files carry `id:`, not
// the `name:` Goose expects, so they are not counted as readable).
func Goose() Profile {
	return Profile{
		Host: "goose",
		Rules: map[primitive.Kind]Rule{
			skillK:   native,
			ruleK:    index,
			commandK: index,
			subagentK: {Strategy: Shim, Dest: nameDest(".goose/agents", ".md"), Render: bodyRenderer(func(p primitive.Primitive) any {
				return namedDoc{Name: primitive.ProjectedName(p.ID), Description: p.Description, Model: p.Model}
			})},
		},
		Reads: map[primitive.Kind][]string{
			skillK:    {".agents/skills"},
			subagentK: {".claude/agents", ".goose/agents"},
		},
	}
}

// Pi reads .agents/skills/ natively and prompt templates from .pi/prompts/.
// It has no subagents and no path-scoped rules.
func Pi() Profile {
	return Profile{
		Host: "pi",
		Rules: map[primitive.Kind]Rule{
			skillK: native,
			ruleK:  index,
			commandK: {Strategy: Shim, Dest: nameDest(".pi/prompts", ".md"), Render: bodyRenderer(func(p primitive.Primitive) any {
				return promptFile{Description: p.Description, ArgumentHint: argumentHint(p.Args)}
			})},
			subagentK: index,
		},
		Reads: map[primitive.Kind][]string{skillK: {".agents/skills"}},
	}
}

// Aider has none of these mechanisms; it reads AGENTS.md through
// .aider.conf.yml, so every kind is indexed there.
func Aider() Profile {
	return Profile{
		Host:  "aider",
		Rules: map[primitive.Kind]Rule{skillK: index, ruleK: index, commandK: index, subagentK: index},
	}
}

// profiles lists every host in plan order. Claude Code comes first because
// Cursor, Copilot, Continue, and Goose read .claude/ directories.
var profiles = []func() Profile{ClaudeCode, Cursor, GitHubCopilot, Continue, OpenCode, Codex, Cline, Goose, Pi, Aider}

// Enabled returns the profiles to project for keystone.json `adapters`, in
// plan order. Claude Code is always included (the 4.x default).
func Enabled(adapters []string) []Profile {
	want := map[string]bool{"claude-code": true}
	for _, a := range adapters {
		want[a] = true
	}
	var out []Profile
	for _, mk := range profiles {
		if p := mk(); want[p.Host] {
			out = append(out, p)
		}
	}
	return out
}

// skillDir returns the canonical directory of a skill (the parent of
// SKILL.md), repo-relative.
func skillDir(p primitive.Primitive) string { return path.Dir(p.Path) }
