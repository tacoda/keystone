# Cursor — Sensor binding

How sensors actually fire inside Cursor.

## Execution model

Cursor has two interaction modes:

- **Agent mode** — the agent can request shell commands; the user accepts each one. Sensors run autonomously (with per-command approval). This is the recommended mode for any session that will invoke the **verify** action.
- **Chat / ask mode** — the agent surfaces commands as text; the user runs them in a terminal and pastes the output back. The slow path; the charter still works but every sensor round-trips through the human.

The per-command approval in agent mode is Cursor's only autonomy lever. Even in "auto-accept" toggles, the charter assumes the user is reviewing each shell command before it runs.

## Per-sensor binding

The sensor commands are read from `charter/corpus/state/CODEBASE_STATE.md` (populated by the **bootstrap** action). Cursor invokes them via the shell tool when in agent mode.

| Sensor | Fires inside | Implementation |
|---|---|---|
| lint | **verify** action | Shell tool with the project's lint command. |
| type-check | **verify** action | Shell tool with type-check command. Skipped when no type checker. |
| test | **verify** action | Shell tool with test command. Fresh run per turn. |
| build | **verify** action | Shell tool with build command. |
| drift | **check-drift** + **verify** + **audit** | Agent reads loaded rules + diff; compares; reports findings. No shell needed. |
| coverage | **verify** + **audit** | Shell tool runs coverage command; agent proposes a diff to `state/CODEBASE_STATE.md` (user confirms in the standard tool-edit flow). |
| risk-fingerprint | **audit** | Shell tool for complexity metrics; `git log` for churn; agent writes table to `state/risk-fingerprints.md`. |
| traffic-topology | **audit** | `git log` + `state/CODEBASE_STATE.md` criticality → diff to `state/traffic-topology.md`. |
| state-region | **orient** action | Agent reads `state/CODEBASE_STATE.md` and active migrations for the touched paths. No shell needed. |
| commit-message | **release** phase, pre-commit | Agent inspects the proposed message before invoking `git commit`. |
| tracker-card-fetcher | **spec** action | No native integration — user pastes card content, or the agent fetches via web in agent mode. |
| spec-adherence | **review** action | Agent reads spec + diff and walks AC line-by-line. No shell needed. |

## State writes

Sensors that update state files propose edits via Cursor's standard edit-confirmation flow. The user accepts, edits, or rejects each diff before it lands. No silent overwrites — the same as Claude Code's behavior, just via a different UI affordance.

## Stale evidence guard

The **verify** action's IRON LAW — "no completion claims without fresh verification evidence" — depends on the sensor commands actually running in the current chat turn. The `verify.mdc` rule body instructs the agent to:

1. Run each sensor via the shell tool.
2. Wait for output.
3. Report the result inline.

In chat-only mode (no shell), the agent surfaces the commands and waits for the user to paste output. The IRON LAW is harder to enforce there — see the `_generic` adapter's discussion of the paste-and-report model.

## Sensor failure handling

When a sensor fails:

1. The **verify** action surfaces the structured output (linter findings, test failures) from the shell tool result.
2. The agent returns to the **implementation** phase; do not "fix forward" inside verification.
3. After the fix, re-invoke `@keystone-verify`. Each fix-and-retry is a fresh shell tool call → fresh evidence.

## Differences from Claude Code

- Cursor has **no sub-agent parallelism** — the **review** action runs each review concern sequentially over the diff in the same chat. Slower; otherwise identical in shape.
- Cursor has **no native tracker MCP** — Atlassian/Linear/GitHub Issues integration is via web fetch or user paste.
- Cursor has **per-command approval** as its only autonomy lever — there is no "autopilot" equivalent. Treat all sessions as `paired`.
