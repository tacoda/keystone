# Aider — Sensor binding

How sensors actually fire inside Aider.

## Execution model

Aider can run shell commands directly via `/run <cmd>` and (specifically for the test sensor) `/test`. **Sensors run autonomously** within Aider's session — the user does not paste output back from a separate terminal.

The `/run` and `/test` commands feed their output back into Aider's conversation. This is the same shape as Claude Code's Bash tool or Cursor's shell tool, with one quirk: Aider considers `/run` output a candidate for inclusion in future LLM calls, so verbose sensor output can balloon the context. The mitigation is `/clear` after a verify cycle if the output was large.

## Per-sensor binding

Sensor commands come from `harness/state/CODEBASE_STATE.md`, populated by the **bootstrap** action.

| Sensor | Fires inside | Implementation |
|---|---|---|
| lint | **verify** action | `/run <lint command>` |
| type-check | **verify** action | `/run <type-check command>`. Skipped when no type checker. |
| test | **verify** action | `/test` (Aider's native test command) — or `/run <test command>` if the project does not configure `test-cmd` in `.aider.conf.yml`. Fresh run per turn. |
| build | **verify** action | `/run <build command>` |
| drift | **check-drift** + **verify** + **audit** | Agent reads loaded corpus rules + `git diff`; compares; reports findings. No shell needed. |
| coverage | **verify** + **audit** | `/run <coverage command>`; agent proposes a diff to `state/CODEBASE_STATE.md` via Aider's edit flow. |
| risk-fingerprint | **audit** | `/run` for complexity metrics + `git log --stat`; agent writes table to `state/risk-fingerprints.md`. |
| traffic-topology | **audit** | `git log` + `state/CODEBASE_STATE.md` criticality → `state/traffic-topology.md`. |
| state-region | **orient** action | Aider reads `state/CODEBASE_STATE.md` and active migrations directly. No shell needed. |
| commit-message | **release** phase, pre-commit | Aider inspects the proposed message before the commit. **Requires `auto-commits: false` in `.aider.conf.yml`** — see activation.md. |
| tracker-card-fetcher | **spec** action | `/run gh issue view <id>` for GitHub Issues; otherwise user pastes. |
| spec-adherence | **review** action | Aider reads spec + diff and walks AC. No shell needed. |

## State writes

Sensors that update state files propose edits via Aider's standard edit flow. Aider shows the diff; the user can accept, reject, or modify. `auto-commits: false` is required so the proposed edit can be inspected before it lands in a commit.

## Stale evidence guard

The **verify** action's IRON LAW — "no completion claims without fresh verification evidence" — requires sensors to run in the current Aider turn. The agent runs each sensor via `/run` (or `/test`) and reports the output inline. Aider's transcript serves as the evidence record.

When in doubt about freshness, the user can ask Aider to "re-run the sensors" — `/test` and `/run <cmd>` are explicit shell invocations, and their output is visible.

## Sensor failure handling

When a sensor fails:

1. The **verify** action surfaces the structured output (linter findings, test failures) from the `/run` or `/test` result.
2. Aider returns to the **implementation** phase. Do not "fix forward" inside verify.
3. After the fix, re-run `/test` (or the failing sensor). Fresh `/run` = fresh evidence.

## A useful pattern: scripts/verify.sh

If the project has 4–6 sensors per verify cycle, the `/run` round-trips can dominate. A consolidated `scripts/verify.sh` that runs all sensors and exits non-zero on any failure reduces the cycle to a single `/run`. The harness recommends this for any project where the verify cycle takes more than ~30 seconds wall-clock per sensor.

## Differences from Claude Code

- Aider has **no sub-agent parallelism** — review runs sequentially.
- Aider has **no native tracker MCP** — tracker integration via shell (`gh`, `curl`, paste).
- Aider has **no in-place context compression** equivalent of `/compact` — `/clear` is the only reset.
- Aider's **`auto-commits: true`** default conflicts with the commit-message sensor; the harness requires it off.
