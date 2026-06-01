# GitHub Copilot — Sensor binding

How sensors actually fire inside GitHub Copilot.

## Execution model

Copilot in VS Code agent mode and Copilot CLI can both run shell commands directly. **Sensors run autonomously** in either surface, with **per-command approval** as the only autonomy lever — the user accepts or denies each shell command before it executes.

In VS Code's chat-only (non-agent) mode, the agent surfaces commands as text and the user runs them in a separate terminal. The slow path; the harness still works but every sensor round-trips through the human.

## Per-sensor binding

Sensor commands come from `harness/state/CODEBASE_STATE.md`, populated by the **bootstrap** action.

| Sensor | Fires inside | Implementation |
|---|---|---|
| lint | **verify** action | Shell tool with the project's lint command. |
| type-check | **verify** action | Shell tool with type-check command. Skipped when no type checker. |
| test | **verify** action | Shell tool with test command. Fresh run per turn. |
| build | **verify** action | Shell tool with build command. |
| drift | **check-drift** + **verify** + **audit** | Agent reads loaded corpus rules + `git diff`; compares; reports findings. No shell needed. |
| coverage | **verify** + **audit** | Shell tool runs coverage command; agent proposes a diff to `state/CODEBASE_STATE.md` via the editor's standard diff-confirmation flow. |
| risk-fingerprint | **audit** | Shell for complexity metrics + `git log --stat`; agent writes table to `state/risk-fingerprints.md`. |
| traffic-topology | **audit** | `git log` + `state/CODEBASE_STATE.md` criticality → `state/traffic-topology.md`. |
| state-region | **orient** action | Agent reads `state/CODEBASE_STATE.md` and active migrations. No shell needed. |
| commit-message | **release** phase, pre-commit | Agent inspects the proposed message before invoking `git commit`. |
| tracker-card-fetcher | **spec** action | `gh issue view <id>` (or native issue context if Copilot is fetching directly). Paste for non-GitHub trackers. |
| spec-adherence | **review** action | Agent reads spec + diff and walks AC. No shell needed. |
| ci-status | **release** action | `gh run list --branch <current>` / `gh run view <id>` — checks CI before opening / merging PR. |

## State writes

Sensors that update state files propose edits via the editor's standard diff-confirmation flow (in VS Code) or as printed suggestions (in CLI). The user accepts or rejects each edit. No silent overwrites.

## Stale evidence guard

The **verify** action's IRON LAW — "no completion claims without fresh verification evidence" — depends on each sensor command running in the current chat turn. The agent shows:

1. The shell command it invoked.
2. The tool output.
3. The PASS / FAIL determination.

Claims without the tool-output evidence are flagged by the **review** action's spec-adherence pass.

## Sensor failure handling

When a sensor fails:

1. Surface the structured output (linter findings, test failures) from the shell result.
2. Return to the **implementation** phase. Do not "fix forward" inside verify.
3. After the fix, re-invoke the sensor. Each fix-and-retry is a fresh shell call → fresh evidence.

## GitHub-native sensor: ci-status

Distinct from the other adapters, the GitHub Copilot adapter exposes **ci-status** as a first-class sensor in the **release** phase. Before opening a PR, before merging:

- `gh run list --branch $(git branch --show-current) --limit 5` — latest CI runs for this branch.
- `gh run view <id> --log-failed` — details on any failing run.

The harness's `release.md` describes when to gate on this; the binding above is how Copilot implements it.

## Differences from Claude Code

- Copilot has **no sub-agent parallelism** — review is sequential.
- Copilot has **per-command approval** as the only autonomy lever — no `autopilot`.
- Copilot has **native GitHub integration** that surpasses every other adapter for GitHub-Issues-tracked projects.
- Copilot has **no MCP equivalent** for Jira / Linear / Asana — those trackers are paste-driven.
