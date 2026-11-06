# Claude Code — Sensor binding

How sensors actually fire inside Claude Code.

## Execution model

Claude Code can run shell commands directly via the Bash tool during a turn. **All sensors run autonomously** — no human paste-and-report loop required. The agent reads the tool commands from `.charter/corpus/state/CODEBASE_STATE.md`, invokes them via Bash, and consumes the output.

## Per-sensor binding

| Sensor | Fires inside | Implementation |
|---|---|---|
| lint | **verify** action | `Bash` tool with the project's lint command from `CODEBASE_STATE.md`. |
| type-check | **verify** action | `Bash` tool with type-check command. Skipped when no type checker. |
| test | **verify** action | `Bash` tool with test command. Fresh run per turn (stale evidence does not count). |
| build | **verify** action | `Bash` tool with build command. |
| drift | **check-drift** + **verify** + **audit** | Read tool reads loaded corpus rules and the diff; agent compares; reports findings. |
| coverage | **verify** + **audit** | `Bash` runs the coverage command; agent proposes a diff to `state/CODEBASE_STATE.md`. |
| risk-fingerprint | **audit** | Agent reads code metrics (via `Bash` for cyclomatic complexity tools, `git log` for churn) and writes the fingerprint table to `state/risk-fingerprints.md`. |
| traffic-topology | **audit** | `git log` + `state/CODEBASE_STATE.md` criticality flags → `state/traffic-topology.md` diff. |
| state-region | **orient** action | Read tool reads `state/CODEBASE_STATE.md` and active migrations for the touched paths. |
| commit-message | **release** phase, pre-commit | Agent inspects the proposed message before invoking `git commit`. |
| tracker-card-fetcher | **spec** action | MCP server call (Atlassian / Linear / GitHub / Asana). |
| spec-adherence | **review** action | Agent reads the spec file + diff and walks AC line-by-line. |

## State writes

Sensors that update state files (`coverage`, `risk-fingerprint`, `traffic-topology`) propose a diff via the Edit tool. The user accepts, edits, or rejects via Claude Code's standard tool-confirmation prompt. No silent overwrites.

## Stale evidence guard

The **verify** action's IRON LAW — "no completion claims without fresh verification evidence" — is enforced by re-running each sensor in the current turn. The agent must show the Bash tool output for sensors it claims passed; the charter review agent (`review-functional`) flags claims without matching tool calls in the turn's transcript.

## Sensor failure handling

When a sensor fails:

1. The **verify** action surfaces the structured output (linter findings, test failures, etc.) from the Bash tool result.
2. The agent returns to the **implementation** phase. Do not "fix forward" inside verification.
3. After the fix, re-invoke **verify**. Each fix-and-retry is a fresh turn — fresh evidence.
