---
description: Run the verification sensors — lint, type-check, test, build, drift, commit-message.
---
# verify

**Run the verification sensors** — lint, type-check, test, build, drift, commit-message. Pre-commit gate. Read [`.charter/guides/process/verification.md`](guides/process/verification.md).

## Activities

Run every sensor classified as runnable in `.charter/corpus/state/CODEBASE_STATE.md`, **in this turn** (no stale evidence). For each sensor:

1. **lint** — [`.charter/sensors/lint.md`](sensors/lint.md). Invoke the project's lint command.
2. **type-check** — [`.charter/sensors/type-check.md`](sensors/type-check.md). Invoke the type-check command. Skip if no type checker.
3. **test** — [`.charter/sensors/test.md`](sensors/test.md). Invoke the test command. Fresh run.
4. **build** — [`.charter/sensors/build.md`](sensors/build.md). Invoke the build command.
5. **drift** — [`.charter/sensors/drift.md`](sensors/drift.md). Compare diff to loaded guides.
6. **commit-message** — [`.charter/sensors/commit-message.md`](sensors/commit-message.md). Inspect the proposed message before `git commit`.

Show the actual tool output for each. Do not claim a sensor passed without producing its output this turn.

## Iron law

**No completion claims without fresh verification evidence.** Sensors must run in this turn.

## On failure

Return to **implementation**. Do not "fix forward" inside verify. After the fix, re-invoke verify in a fresh turn.
