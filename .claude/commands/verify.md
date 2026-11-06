---
description: "Run the verification hooks: lint, type-check, test, build, drift, commit-message."
---
# verify

**Run the verification hooks**: lint, type-check, test, build, drift, commit-message. Pre-commit gate. Read [`.agents/rules/process/verification.md`](../rules/process/verification.md).

## Activities

Run every hook classified as runnable in `.agents/references/state/CODEBASE_STATE.md`, **in this turn** (no stale evidence). For each hook:

1. **lint** — [`.agents/hooks/lint.md`](../hooks/lint.md). Invoke the project's lint command.
2. **type-check** — [`.agents/hooks/type-check.md`](../hooks/type-check.md). Invoke the type-check command. Skip if no type checker.
3. **test** — [`.agents/hooks/test.md`](../hooks/test.md). Invoke the test command. Fresh run.
4. **build** — [`.agents/hooks/build.md`](../hooks/build.md). Invoke the build command.
5. **drift** — [`.agents/hooks/drift.md`](../hooks/drift.md). Compare diff to loaded rules.
6. **commit-message** — [`.agents/hooks/commit-message.md`](../hooks/commit-message.md). Inspect the proposed message before `git commit`.

Show the actual tool output for each. Do not claim a hook passed without producing its output this turn.

## Iron law

**No completion claims without fresh verification evidence.** Hooks must run in this turn.

## On failure

Return to **implementation**. Do not "fix forward" inside verify. After the fix, re-invoke verify in a fresh turn.
