# opencode — Lifecycle binding

How each abstract lifecycle action is invoked in opencode. opencode reads
`AGENTS.md` at the repo root on every session, runs shell commands
autonomously, and supports slash commands, subagents, and MCP servers —
so it covers the harness lifecycle natively, closer to Claude Code than to
Codex.

## Invocation

Two surfaces, pick whichever fits:

- **Slash command** — `keystone project` projects every command into
  `.opencode/commands/keystone-<id>.md`. Type `/keystone-<id>`
  (e.g. `/keystone-verify`).
- **Natural language** — "run task on TICKET-123," "run verify," "do a
  review pass." The agent reads `AGENTS.md` at session start, finds the
  command, follows the link to its `.harness/commands/<id>.md` body, and
  executes.

The canonical kickoff phrase is **"run task on `<ticket-id>`"** (or the
`/keystone-task` command) — `.harness/playbooks/task.md` orchestrates
`spec → orient → implementation → check-drift → verify → review`.

## Capability matrix

| Capability | Supported? |
|---|---|
| Autonomous shell execution | ✓ |
| Sub-agent parallelism | ✓ (subagents in `.opencode/agents/`) |
| Slash commands | ✓ (`.opencode/commands/`) |
| Skills | ✓ (`.opencode/skills/`) |
| Rules | ✓ (`.opencode/rules/` via `instructions` — always-on, not per-file gated) |
| MCP servers | ✓ (`mcp` key in `opencode.json`) |
| Lazy-by-region | partial (rules load always; orient reads region guides on demand) |
| Context-reset primitive | new session |
| Project-scoped settings | ✓ (`opencode.json` + `.opencode/`) |

## MCP server

Register the Keystone MCP server so opencode can consult the harness at
runtime without reading every markdown file:

```
keystone mcp install --agent opencode
```

This writes the server into `opencode.json`'s `mcp` key as a `local`
(stdio) server running `keystone mcp serve`. Restart opencode to pick it
up.

## Tracker integration

Fetch cards via opencode's shell access or a configured MCP server:

- GitHub Issues: `gh issue view <id>`
- Linear / Jira: the respective CLI, an MCP server, or paste card content

The **spec** action's tracker-fetcher sensor falls back to "paste the
card" if no source is configured.
