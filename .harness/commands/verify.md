---
kind: command
id: verify
description: 'Run the verification sensors — lint, type-check, test, build, drift, commit-message.'
---
# verify

**Run the verification sensors** — lint, type-check, test, build, drift, commit-message. Pre-commit gate. Read [`harness/guides/process/verification.md`](guides/process/verification.md).

## Activities

Run every sensor classified as runnable in `harness/corpus/state/CODEBASE_STATE.md`, **in this turn** (no stale evidence). For each sensor:

1. **lint** — [`harness/sensors/lint.md`](sensors/lint.md). Invoke the project's lint command.
2. **type-check** — [`harness/sensors/type-check.md`](sensors/type-check.md). Invoke the type-check command. Skip if no type checker.
3. **test** — [`harness/sensors/test.md`](sensors/test.md). Invoke the test command. Fresh run.
4. **build** — [`harness/sensors/build.md`](sensors/build.md). Invoke the build command.
5. **drift** — [`harness/sensors/drift.md`](sensors/drift.md). Compare diff to loaded guides.
6. **commit-message** — [`harness/sensors/commit-message.md`](sensors/commit-message.md). Inspect the proposed message before `git commit`.

Show the actual tool output for each. Do not claim a sensor passed without producing its output this turn.

## Iron law

**No completion claims without fresh verification evidence.** Sensors must run in this turn.

## On failure

Return to **implementation**. Do not "fix forward" inside verify. After the fix, re-invoke verify in a fresh turn.
