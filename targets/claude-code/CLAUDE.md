# CLAUDE.md

This project uses a **project harness**. The corpus at [`harness/`](harness/) defines the engineering knowledge and the workflow phases you'll be operating within.

## Read first

- [`harness/README.md`](harness/README.md) — the five layers, the lifecycle actions, the flywheels.
- [`harness/adapters/claude-code/`](harness/adapters/claude-code/) — how each lifecycle action binds to Claude Code (slash commands, sub-agents, MCP tracker integration).

## Lifecycle in this project

| Action | Slash command |
|---|---|
| spec | `/<prefix>:spec` |
| orient | `/<prefix>:orient` |
| check-drift | `/<prefix>:check-drift` |
| verify | `/<prefix>:verify` |
| review | `/<prefix>:review` |
| learn | `/<prefix>:learn` |
| bootstrap | `/<prefix>:bootstrap` (one-time, on first install) |
| audit | `/<prefix>:audit` |
| synthesize | `/<prefix>:synthesize` |
| mode | `/<prefix>:mode <paired\|solo\|autopilot>` |

Replace `<prefix>` with the slash-command namespace this project uses. If commands were installed via a Claude Code plugin, the prefix is the plugin name. If installed into `.claude/commands/`, the prefix is whatever the project picked at bootstrap.

## Iron laws

Non-negotiable across every phase:

- **No proceeding without explicit acceptance criteria** in the spec.
- **No completion claims without fresh verification evidence** — sensors must have run this turn.
- **No commits with failing sensors.** Never `--no-verify`.
- **No AI attribution** in commits, PRs, or tracker comments — no `Co-Authored-By: Claude`, no `🤖 Generated with Claude Code` footer.
- **No silent overwrites** of state files — propose a diff, ask before applying.

## Plugin and sub-agent surface

- `.claude/commands/*.md` — project-local slash commands (if not installed as a plugin).
- `.claude/agents/review-*.md` — review agents spawned by the `review` action.
- `.claude/settings.json` — permissions, hooks, MCP servers.

## Prerequisites

- A way to track work — a tracker card (Atlassian / Linear / GitHub Issues / Asana MCP servers preferred), a `TODO.md`, or a conversation.
- Lint / type-check / test / build commands in `harness/state/CODEBASE_STATE.md`.
- PR workflow.
- CI (and ideally CD).
