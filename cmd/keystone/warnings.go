package main

import (
	"fmt"
	"os"
)

// agentGap describes a charter feature the selected agent does not natively
// support, with actionable remedies the user can apply after install.
type agentGap struct {
	feature   string // short, user-facing capability name
	impact    string // what breaks or degrades if the gap is not closed
	configFix string // how to configure the agent to close the gap (empty = not configurable)
	fallback  string // a charter file the user can add to document handling — empty if not applicable
}

// agentWarnings lists capability gaps per agent. Keyed by the same `--agent`
// values the CLI accepts. Agents not in the map are considered "full support"
// and get no warning. Order in each slice is the order printed.
var agentWarnings = map[string][]agentGap{
	"aider": {
		{
			feature:   "sub-agent parallelism (review)",
			impact:    "the review action runs each concern (functional + security) sequentially in one chat instead of in parallel",
			configFix: "",
			fallback:  ".agents/references/hosts/aider/review-strategy.md",
		},
		{
			feature:   "commit-message hook gate",
			impact:    "Aider's default auto-commits fire before the commit-message hook can inspect the message",
			configFix: "set `auto-commits: false` in `.aider.conf.yml`",
			fallback:  "",
		},
		{
			feature:   "native tracker integration",
			impact:    "the tracker-card-fetcher subagent falls back to shell or user paste",
			configFix: "use `/run gh issue view <id>` for GitHub Issues, or paste card content for Jira/Linear/Asana",
			fallback:  ".agents/references/hosts/aider/tracker-workflow.md",
		},
	},
	"continue": {
		{
			feature:   "lifecycle skills",
			impact:    "the keystone-* lifecycle skills reach Continue as copies in `.continue/skills/`; without them, every lifecycle step must be asked for in chat",
			configFix: "run `keystone project` after each charter change so the copies stay current",
			fallback:  "",
		},
		{
			feature:   "sub-agent parallelism (review)",
			impact:    "the review action runs each concern sequentially in one chat",
			configFix: "",
			fallback:  ".agents/references/hosts/continue/review-strategy.md",
		},
		{
			feature:   "native tracker integration",
			impact:    "the tracker-card-fetcher subagent needs an MCP server or shell fallback",
			configFix: "configure an Atlassian/Linear/GitHub MCP server in `.continue/config.yaml`'s `mcpServers:` block",
			fallback:  ".agents/references/hosts/continue/tracker-workflow.md",
		},
	},
	"cline": {
		{
			feature:   "menu auto-load",
			impact:    "Cline only auto-reads `.clinerules` on newer versions; older versions need the charter's orientation in the extension's custom-instructions field",
			configFix: "paste the keystone block of `AGENTS.md` into Cline's `Custom Instructions` field in VS Code settings",
			fallback:  "",
		},
		{
			feature:   "lifecycle workflows",
			impact:    "without saved workflows, every lifecycle step must be asked for in chat",
			configFix: "save a `Keystone: <step>` workflow per keystone-* lifecycle skill (spec, verify, review, …) via the side panel's workflows menu",
			fallback:  "",
		},
		{
			feature:   "auto-approve for verify cycle",
			impact:    "without auto-approve, the verify skill stalls on a click per hook command",
			configFix: "in Cline's settings, auto-approve file reads + read-only git commands + project test/lint/build commands",
			fallback:  "",
		},
		{
			feature:   "sub-agent parallelism (review)",
			impact:    "review runs sequentially; subtasks are also sequential",
			configFix: "",
			fallback:  ".agents/references/hosts/cline/review-strategy.md",
		},
	},
	"goose": {
		{
			feature:   "developer extension",
			impact:    "without the developer extension, hooks degrade to 'agent surfaces commands, user runs them'",
			configFix: "run `goose configure --enable-extension developer` (or enable it in the desktop app's Extensions settings)",
			fallback:  "",
		},
		{
			feature:   "lifecycle recipes",
			impact:    "non-interactive lifecycle invocation needs recipes; without them, every action is a natural-language ask in `goose session`",
			configFix: "the installer drops recipes at `.goose/recipes/keystone-<action>.yaml` — review and adjust per project; invoke with `goose run --recipe <path>`",
			fallback:  "",
		},
		{
			feature:   "native tracker integration",
			impact:    "the tracker-card-fetcher subagent needs an MCP extension or shell fallback",
			configFix: "run `goose configure --add-extension` and configure an Atlassian/Linear/GitHub MCP server",
			fallback:  ".agents/references/hosts/goose/tracker-workflow.md",
		},
		{
			feature:   "sub-agent parallelism (review)",
			impact:    "review runs sequentially",
			configFix: "",
			fallback:  ".agents/references/hosts/goose/review-strategy.md",
		},
	},
	"github-copilot": {
		{
			feature:   "lifecycle slash commands",
			impact:    "Copilot has no slash-command primitive for project actions; every action is a natural-language ask",
			configFix: "",
			fallback:  ".agents/references/hosts/github-copilot/invocation-cheatsheet.md",
		},
		{
			feature:   "sub-agent parallelism (review)",
			impact:    "review runs sequentially in one chat",
			configFix: "",
			fallback:  ".agents/references/hosts/github-copilot/review-strategy.md",
		},
		{
			feature:   "tracker integration outside GitHub",
			impact:    "Copilot's native tracker fetch only covers GitHub Issues; Jira/Linear/Asana need paste or a separate MCP",
			configFix: "use `gh issue view` for GitHub Issues; for other trackers, paste card content or configure an MCP",
			fallback:  ".agents/references/hosts/github-copilot/tracker-workflow.md",
		},
	},
	"generic": {
		{
			feature:   "lifecycle invocation",
			impact:    "the generic host assumes only that the agent can read markdown on demand; lifecycle steps are entirely user-driven",
			configFix: "if your agent has a rules file, slash-command, or saved-workflow primitive, define it as a custom host (`keystone new adapter <name>`) or record your bindings in `.agents/references/hosts/<your-agent>/lifecycle.md`",
			fallback:  ".agents/references/hosts/generic/lifecycle-overrides.md",
		},
		{
			feature:   "autonomous hook execution",
			impact:    "unknown whether the agent can run shell during a turn; hooks may degrade to 'agent surfaces commands, user runs them'",
			configFix: "if the agent has shell access, document it in `.agents/references/hosts/<your-agent>/hooks.md`; otherwise document the surfaced-command workflow",
			fallback:  ".agents/references/hosts/generic/hooks-overrides.md",
		},
		{
			feature:   "sub-agent parallelism (review)",
			impact:    "assume review runs sequentially",
			configFix: "",
			fallback:  ".agents/references/hosts/generic/review-strategy.md",
		},
		{
			feature:   "tracker integration",
			impact:    "unknown; assume paste or shell fallback",
			configFix: "if the agent supports MCP, configure a tracker server; otherwise document the manual workflow",
			fallback:  ".agents/references/hosts/generic/tracker-workflow.md",
		},
	},
}

// printAgentWarnings prints capability gaps for the selected agent. Called
// after install lands and before the success / next-steps block, so the gaps
// are visible before the user moves on.
//
// Agents not in agentWarnings are silently skipped — these are the
// fully-supported adapters (claude-code, codex, pi, cursor at this writing).
func printAgentWarnings(agent string) {
	gaps, ok := agentWarnings[agent]
	if !ok || len(gaps) == 0 {
		return
	}

	fmt.Fprintf(os.Stdout, "\n⚠ %s adapter — capability gaps to address\n", agent)
	fmt.Fprintf(os.Stdout, "\nThe %s adapter does not natively cover every charter feature.\n", agent)
	fmt.Fprintf(os.Stdout, "Each gap below has a remedy: either configure the agent, or add a\n")
	fmt.Fprintf(os.Stdout, "charter file documenting how your team handles the gap. Some agents\n")
	fmt.Fprintf(os.Stdout, "fundamentally cannot close certain gaps — fall back to the charter\n")
	fmt.Fprintf(os.Stdout, "file approach in those cases.\n\n")

	// Fallback paths point at a reference the team writes under
	// .agents/references/hosts/ (opened on demand).
	for _, g := range gaps {
		fmt.Fprintf(os.Stdout, "  • %s\n", g.feature)
		fmt.Fprintf(os.Stdout, "      Impact: %s.\n", g.impact)
		if g.configFix != "" {
			fmt.Fprintf(os.Stdout, "      Configure: %s.\n", g.configFix)
		}
		if g.fallback != "" {
			fmt.Fprintf(os.Stdout, "      Or document: add %s describing how your team handles this.\n", g.fallback)
		}
		if g.configFix == "" && g.fallback == "" {
			fmt.Fprintln(os.Stdout, "      Workaround: see the host's line in AGENTS.md (keystone block).")
		}
		fmt.Fprintln(os.Stdout)
	}
}
