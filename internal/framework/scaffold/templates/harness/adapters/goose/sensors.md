# Goose — Sensor binding

How sensors actually fire inside Goose.

## Execution model

Goose runs shell commands through the **developer extension's `bash` tool**. In interactive sessions, the user sees the command before it runs and can approve or block it (Goose's default risk-classifier asks for confirmation on commands it deems risky). In recipe mode, the recipe runs to completion without prompts unless explicit confirmation steps are configured.

**Sensors run autonomously** when the developer extension is enabled — the user is not pasting output from a separate terminal. Without the developer extension, sensors degrade to "agent surfaces, user runs."

## Per-sensor binding

Sensor commands come from `harness/corpus/state/CODEBASE_STATE.md`, populated by the **bootstrap** action.

| Sensor | Fires inside | Implementation |
|---|---|---|
| [lint](../../sensors/lint.md) | **verify** action | `bash` tool running `<lint command>` |
| [type-check](../../sensors/type-check.md) | **verify** action | `bash` tool running `<type-check command>`. Skipped when no type checker. |
| [test](../../sensors/test.md) | **verify** action | `bash` tool running `<test command>`. Fresh run per turn. |
| [build](../../sensors/build.md) | **verify** action | `bash` tool running `<build command>` |
| [drift](../../sensors/drift.md) | **check-drift** + **verify** + **audit** | `bash` for `git diff`; agent reads loaded guides and compares. No state writes. |
| [coverage](../../sensors/coverage.md) | **verify** + **audit** | `bash` running `<coverage command>`; agent proposes a diff to `corpus/state/CODEBASE_STATE.md` via the `text_editor` tool. |
| [risk-fingerprint](../../sensors/risk-fingerprint.md) | **audit** | `bash` for complexity metrics + `git log --stat`; agent writes table to `corpus/state/risk-fingerprints.md`. |
| [traffic-topology](../../sensors/traffic-topology.md) | **audit** | `git log` + criticality from `corpus/state/CODEBASE_STATE.md` → `corpus/state/traffic-topology.md`. |
| [state-region](../../sensors/state-region.md) | **orient** action | Agent reads `corpus/state/CODEBASE_STATE.md` and active migrations via `text_editor`. No shell needed. |
| [commit-message](../../sensors/commit-message.md) | **release** phase, pre-commit | Agent inspects the proposed message; runs `git commit` via `bash` only after the message passes. |
| [tracker-card-fetcher](../../sensors/tracker-card-fetcher.md) | **spec** action | MCP extension (GitHub / Atlassian / Linear) if installed; otherwise `bash` running `gh issue view`. Fallback: user pastes. |
| [spec-adherence](../../sensors/spec-adherence.md) | **review** action | Agent reads spec + diff and walks AC. No shell needed. |

## State writes

Sensors that update state files propose edits via the developer extension's `text_editor` tool. In interactive mode, the user sees the proposed change inline; in recipe mode, edits land directly — recipes should be reserved for read-only or human-approved-elsewhere actions. The **audit** recipe specifically should be run interactively or with a review step.

## Stale evidence guard

The **verify** action's IRON LAW requires sensors to have run *this turn*. Goose's session transcript records every `bash` invocation; the agent reports completion only after each sensor passes in the current session. When in doubt, ask Goose to "re-run the sensors" — fresh invocations.

## Sensor failure handling

When a sensor fails:

1. The **verify** action surfaces the structured output (linter findings, test failures) from the `bash` result.
2. The agent returns to the **implementation** phase. Do not "fix forward" inside verify.
3. After the fix, re-invoke the failing sensor. Fresh `bash` output = fresh evidence.

## A useful pattern: keystone-verify recipe

A consolidated recipe that runs every sensor and reports collectively:

```yaml
# .goose/recipes/keystone-verify.yaml
version: 1.0.0
title: Keystone — verify
description: Run every sensor against the current change
instructions: |
  You are running the verify action. Read harness/guides/process/verification.md
  for the contract. Sensor commands live in harness/corpus/state/CODEBASE_STATE.md.

  Run each sensor in order. Stop and report immediately on the first failure;
  do not continue to subsequent sensors.

  Report format:
  - lint: pass | fail (<error count>)
  - type-check: pass | fail | skipped
  - test: pass | fail (<failure count>)
  - build: pass | fail
  - drift: pass | fail (<violation count>)
  - commit-message: deferred to release

  On any failure, return to implementation phase.
extensions:
  - name: developer
prompt: |
  Run the verify action for the current diff.
```

Invoke: `goose run --recipe .goose/recipes/keystone-verify.yaml`.

## Recipe vs. session for sensors

| Sensor cycle | Recommended invocation |
|---|---|
| Fast iteration during implementation | Interactive session (`goose session`); ask "run the verify action" |
| Pre-commit check | `keystone-verify` recipe |
| Full audit | `keystone-audit` recipe — but review the output in interactive mode before any state writes land |

## Differences from Claude Code

- Goose has **no sub-agent parallelism** — review runs sequentially.
- Goose's **risk-classifier** asks for confirmation on commands it deems risky; whitelist trusted commands in the global config or run via recipe to suppress prompts.
- Goose **must have the developer extension enabled** for the harness to be useful; without it, sensors degrade.
- Goose recipes are the closest analog to Claude Code's slash commands; the harness uses them as such.
