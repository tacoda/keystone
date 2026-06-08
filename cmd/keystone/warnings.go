package main

import (
	"fmt"
	"os"
)

// agentGap describes a harness feature the selected agent does not natively
// support, with actionable remedies the user can apply after install.
type agentGap struct {
	feature   string // short, user-facing capability name
	impact    string // what breaks or degrades if the gap is not closed
	configFix string // how to configure the agent to close the gap (empty = not configurable)
	fallback  string // a harness file the user can add to document handling — empty if not applicable
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
			fallback:  "harness/adapters/aider/review-strategy.md",
		},
		{
			feature:   "commit-message sensor gate",
			impact:    "Aider's default auto-commits fire before the commit-message sensor can inspect the message",
			configFix: "set `auto-commits: false` in `.aider.conf.yml`",
			fallback:  "",
		},
		{
			feature:   "native tracker integration",
			impact:    "the tracker-card-fetcher sensor falls back to shell or user paste",
			configFix: "use `/run gh issue view <id>` for GitHub Issues, or paste card content for Jira/Linear/Asana",
			fallback:  "harness/adapters/aider/tracker-workflow.md",
		},
	},
	"continue": {
		{
			feature:   "lifecycle slash commands",
			impact:    "without slash commands, every lifecycle action must be invoked as natural-language chat",
			configFix: "add a `.continue/config.yaml` with the `prompts:` block from `harness/adapters/continue/lifecycle.md`",
			fallback:  "",
		},
		{
			feature:   "sub-agent parallelism (review)",
			impact:    "the review action runs each concern sequentially in one chat",
			configFix: "",
			fallback:  "harness/adapters/continue/review-strategy.md",
		},
		{
			feature:   "native tracker integration",
			impact:    "the tracker-card-fetcher sensor needs an MCP server or shell fallback",
			configFix: "configure an Atlassian/Linear/GitHub MCP server in `.continue/config.yaml`'s `mcpServers:` block",
			fallback:  "harness/adapters/continue/tracker-workflow.md",
		},
	},
	"cline": {
		{
			feature:   "menu auto-load",
			impact:    "Cline only auto-reads `.clinerules` on newer versions; older versions need the menu in the extension's custom-instructions field",
			configFix: "paste the contents of `cline-instructions.md` into Cline's `Custom Instructions` field in VS Code settings",
			fallback:  "",
		},
		{
			feature:   "lifecycle workflows",
			impact:    "without saved workflows, every lifecycle action must be invoked as natural-language chat",
			configFix: "add the `Keystone: <action>` workflows listed in `harness/adapters/cline/lifecycle.md` via the side panel's workflows menu",
			fallback:  "",
		},
		{
			feature:   "auto-approve for verify cycle",
			impact:    "without auto-approve, the verify action stalls on a click per sensor command",
			configFix: "in Cline's settings, auto-approve file reads + read-only git commands + project test/lint/build commands",
			fallback:  "",
		},
		{
			feature:   "sub-agent parallelism (review)",
			impact:    "review runs sequentially; subtasks are also sequential",
			configFix: "",
			fallback:  "harness/adapters/cline/review-strategy.md",
		},
	},
	"goose": {
		{
			feature:   "developer extension",
			impact:    "without the developer extension, sensors degrade to 'agent surfaces commands, user runs them'",
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
			impact:    "the tracker-card-fetcher sensor needs an MCP extension or shell fallback",
			configFix: "run `goose configure --add-extension` and configure an Atlassian/Linear/GitHub MCP server",
			fallback:  "harness/adapters/goose/tracker-workflow.md",
		},
		{
			feature:   "sub-agent parallelism (review)",
			impact:    "review runs sequentially",
			configFix: "",
			fallback:  "harness/adapters/goose/review-strategy.md",
		},
	},
	"github-copilot": {
		{
			feature:   "lifecycle slash commands",
			impact:    "Copilot has no slash-command primitive for project actions; every action is a natural-language ask",
			configFix: "",
			fallback:  "harness/adapters/github-copilot/invocation-cheatsheet.md",
		},
		{
			feature:   "sub-agent parallelism (review)",
			impact:    "review runs sequentially in one chat",
			configFix: "",
			fallback:  "harness/adapters/github-copilot/review-strategy.md",
		},
		{
			feature:   "tracker integration outside GitHub",
			impact:    "Copilot's native tracker fetch only covers GitHub Issues; Jira/Linear/Asana need paste or a separate MCP",
			configFix: "use `gh issue view` for GitHub Issues; for other trackers, paste card content or configure an MCP",
			fallback:  "harness/adapters/github-copilot/tracker-workflow.md",
		},
	},
	"generic": {
		{
			feature:   "lifecycle invocation",
			impact:    "the generic adapter assumes only that the agent can read markdown on demand; lifecycle actions are entirely user-driven",
			configFix: "if your agent has a rules file, slash-command, or saved-workflow primitive, document your bindings in a new `harness/adapters/<your-agent>/lifecycle.md`",
			fallback:  "harness/adapters/generic/lifecycle-overrides.md",
		},
		{
			feature:   "autonomous sensor execution",
			impact:    "unknown whether the agent can run shell during a turn; sensors may degrade to 'agent surfaces commands, user runs them'",
			configFix: "if the agent has shell access, document it in `harness/adapters/<your-agent>/sensors.md`; otherwise document the surfaced-command workflow",
			fallback:  "harness/adapters/generic/sensors-overrides.md",
		},
		{
			feature:   "sub-agent parallelism (review)",
			impact:    "assume review runs sequentially",
			configFix: "",
			fallback:  "harness/adapters/generic/review-strategy.md",
		},
		{
			feature:   "tracker integration",
			impact:    "unknown; assume paste or shell fallback",
			configFix: "if the agent supports MCP, configure a tracker server; otherwise document the manual workflow",
			fallback:  "harness/adapters/generic/tracker-workflow.md",
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
	fmt.Fprintf(os.Stdout, "\nThe %s adapter does not natively cover every harness feature.\n", agent)
	fmt.Fprintf(os.Stdout, "Each gap below has a remedy: either configure the agent, or add a\n")
	fmt.Fprintf(os.Stdout, "harness file documenting how your team handles the gap. Some agents\n")
	fmt.Fprintf(os.Stdout, "fundamentally cannot close certain gaps — fall back to the harness\n")
	fmt.Fprintf(os.Stdout, "file approach in those cases.\n\n")

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
			fmt.Fprintf(os.Stdout, "      Workaround: see harness/adapters/%s/.\n", agent)
		}
		fmt.Fprintln(os.Stdout)
	}
}
