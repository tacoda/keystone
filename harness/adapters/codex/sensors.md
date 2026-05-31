# Codex CLI — Sensor binding

How sensors fire inside Codex CLI.

## Execution model

Codex runs shell commands directly. **All sensors run autonomously** — the agent reads tool commands from `harness/state/CODEBASE_STATE.md`, invokes them via shell, and consumes the output. No human paste-and-report loop.

How aggressively shell runs without approval depends on the Codex flag chosen at session start:

- `codex` (default) → asks before each shell command.
- `codex --auto-edit` → writes files autonomously, asks for shell.
- `codex --full-auto` → both file edits and shell run without per-action approval.

Pick the mode that matches the harness's pacing mode (see `lifecycle.md` for the mapping).

## Per-sensor binding

| Sensor | Fires inside | Implementation |
|---|---|---|
| lint | **verify** action | Shell with the project's lint command from `CODEBASE_STATE.md`. |
| type-check | **verify** action | Shell with type-check command. Skipped when no type checker. |
| test | **verify** action | Shell with test command. Fresh run per turn. |
| build | **verify** action | Shell with build command. |
| drift | **check-drift** + **verify** + **audit** | Agent reads loaded corpus rules + the diff; reports findings. |
| coverage | **verify** + **audit** | Shell runs the coverage command; agent proposes a diff to `state/CODEBASE_STATE.md`. |
| risk-fingerprint | **audit** | Shell for complexity metrics + `git log`; agent writes `state/risk-fingerprints.md` diff. |
| traffic-topology | **audit** | `git log` + `state/CODEBASE_STATE.md` criticality → `state/traffic-topology.md` diff. |
| state-region | **orient** action | Agent reads `state/CODEBASE_STATE.md` and active migrations for touched paths. |
| commit-message | **release** phase, pre-commit | Agent inspects proposed message before `git commit`. |
| tracker-card-fetcher | **spec** action | Shell-based — `gh issue view`, `jira` CLI, etc. Falls back to human paste. |
| spec-adherence | **review** action | Agent reads the spec + diff and walks AC line-by-line. |

## Approval-mode interaction

In `paired` mode (default Codex flag), every sensor invocation prompts the user. This is slow but safe. For active development sessions, switch to `solo` or `autopilot` (Codex `--auto-edit` or `--full-auto`).

The **mode** action updates `harness/process/modes.md`; the user is expected to choose the corresponding Codex flag at the *next* session start. Codex does not switch modes mid-session.

## Stale evidence guard

Same IRON LAW as Claude Code: claims need fresh evidence. The agent must show the shell output for any sensor it claims passed; without fresh output in the current session, the claim is invalid.

## Sensor failure handling

1. The **verify** action surfaces structured output from the failing sensor.
2. The agent returns to the **implementation** phase. Do not "fix forward" inside verification.
3. After the fix, re-invoke verify. Each retry is a fresh evidence cycle.
