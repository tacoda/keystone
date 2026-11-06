# Claude Code — Lifecycle binding

How Claude Code runs the charter's lifecycle actions.

## Invocation

Every action is invoked via natural language: "run task on TICKET-123," "run verify," "do a review pass." The agent reads `CLAUDE.md` at session start, finds the action in the bulleted list, follows the link to `charter/actions/<action>.md`, and executes the playbook. No slash commands, no skill files, no policy install — the playbooks are agent-agnostic.

The canonical kickoff phrase is **"run task on `<ticket-id>`"** (or "run the task workflow") — `charter/actions/task.md` orchestrates `spec → orient → implementation → check-drift → verify → review`.

## Sensor execution

Claude Code can run shell commands directly via the Bash tool. **All sensors run autonomously** — no human paste-and-report loop. The agent reads sensor commands from `charter/corpus/state/CODEBASE_STATE.md`, invokes them via Bash, and consumes the output. See `charter/adapters/claude-code/sensors.md` for the per-sensor binding.

## Sub-agent parallelism

The Agent tool spawns parallel sub-agents. The **review** action takes advantage of this — `review-functional`, `review-security`, `review-risk`, and `review-deployment` run concurrently on the same diff, and findings are combined.

## Autonomy and modes

All three pacing modes (paired / solo / autopilot) are supported. The agent reads `charter/guides/process/modes.md` to determine current behavior.

## Tracker integration

Atlassian, Linear, GitHub Issues, and Asana MCP servers exist. `charter/actions/spec.md` fetches the card via whichever server is configured when a ticket ID is provided.

## Capability matrix

| Capability | Supported? |
|---|---|
| Autonomous shell sensor execution | ✓ |
| Sub-agent parallelism | ✓ |
| Autonomy levels (paired/solo/autopilot) | ✓ |
| Lazy-by-region idiom loading | ✓ (via Read tool) |
| Context-reset primitive | ✓ (`/clear` and `/compact`) |
| MCP tracker integration | ✓ (Atlassian, Linear, GitHub Issues, Asana servers exist) |
