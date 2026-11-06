# Continue — Sensor binding

How sensors actually fire inside Continue.

## Execution model

Continue runs shell commands two ways:

- **Agent mode's built-in terminal tool** — when the model decides to run a command, Continue prompts the user to approve, then executes. Output flows back into the conversation.
- **Custom slash commands with `cmd` steps in `config.yaml`** — declared commands the user can invoke explicitly.

**Sensors run autonomously** in agent mode (with per-command approval) — the user is not pasting output back from a separate terminal. The approval prompt is the gate; sensor commands declared in the charter are deterministic and safe to allow.

## Per-sensor binding

Sensor commands come from `charter/corpus/state/CODEBASE_STATE.md`, populated by the **bootstrap** action.

| Sensor | Fires inside | Implementation |
|---|---|---|
| [lint](sensors/lint.md) | **verify** action | Terminal tool / `cmd` step running `<lint command>` |
| [type-check](sensors/type-check.md) | **verify** action | Terminal tool running `<type-check command>`. Skipped when no type checker. |
| [test](sensors/test.md) | **verify** action | Terminal tool running `<test command>`. Fresh run per turn. |
| [build](sensors/build.md) | **verify** action | Terminal tool running `<build command>` |
| [drift](sensors/drift.md) | **check-drift** + **verify** + **audit** | Agent reads loaded guides + `git diff` (terminal tool or `diff` context provider); compares; reports findings. No state writes. |
| [coverage](sensors/coverage.md) | **verify** + **audit** | Terminal tool running `<coverage command>`; agent proposes a diff to `corpus/state/CODEBASE_STATE.md` via Continue's edit flow. |
| [risk-fingerprint](sensors/risk-fingerprint.md) | **audit** | Terminal tool for complexity metrics + `git log --stat`; agent writes table to `corpus/state/risk-fingerprints.md`. |
| [traffic-topology](sensors/traffic-topology.md) | **audit** | `git log` + `corpus/state/CODEBASE_STATE.md` criticality → `corpus/state/traffic-topology.md`. |
| [state-region](sensors/state-region.md) | **orient** action | Agent reads `corpus/state/CODEBASE_STATE.md` and active migrations directly. No shell needed. |
| [commit-message](sensors/commit-message.md) | **release** phase, pre-commit | Agent inspects the proposed message before the commit. Continue does not auto-commit; the agent runs `git commit` via terminal tool. |
| [tracker-card-fetcher](sensors/tracker-card-fetcher.md) | **spec** action | MCP tracker server (Atlassian / Linear) if configured; otherwise `cmd` step (`gh issue view`) or user pastes. |
| [spec-adherence](sensors/spec-adherence.md) | **review** action | Agent reads spec + diff (via `diff` context provider) and walks AC. No shell needed. |

## State writes

Sensors that update state files propose edits through Continue's standard apply-edit flow. Continue shows the diff in the side panel; the user accepts or declines. No silent writes.

## Stale evidence guard

The **verify** action's IRON LAW requires sensors to run in the current turn. The agent uses the terminal tool (with per-command approval) for each sensor; the conversation transcript records the invocation and output. When in doubt, ask Continue to "re-run the sensors" — fresh `cmd` invocations.

## Sensor failure handling

When a sensor fails:

1. The **verify** action surfaces the structured output (linter findings, test failures) from the terminal tool.
2. The agent returns to the **implementation** phase. Do not "fix forward" inside verify.
3. After the fix, re-invoke the failing sensor. Fresh terminal output = fresh evidence.

## A useful pattern: consolidated verify slash command

If the verify cycle runs 4–6 sensors, the per-command approval prompts add friction. A consolidated `verify` slash command that wraps a `scripts/verify.sh` (which runs every sensor and exits non-zero on any failure) reduces approvals to one:

```yaml
prompts:
  - name: verify-all
    description: "Run every sensor in one shot"
    prompt: |
      Run scripts/verify.sh via the terminal tool and report results.
```

Pair with `scripts/verify.sh` in the project; the charter recommends this when the verify cycle exceeds ~30 seconds wall-clock per sensor.

## Differences from Claude Code

- Continue has **no sub-agent parallelism** — review runs sequentially.
- Tracker integration is **MCP-server-only** (Atlassian, Linear) — no built-in tracker primitive.
- **Per-command approval** in agent mode means each terminal-tool call requires a click unless the user has set a project-wide auto-approve.
- Continue's `diff` context provider gives the model up-to-date diff state without invoking shell — useful for drift and review.
