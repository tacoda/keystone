---
name: keystone-verify
description: 'Run the verification gate before a commit: every runnable hook (lint, type-check, test, build, drift, commit-message), then keystone verify for policy drift and strict locks. Triggers: keystone verify; keystone:verify; /keystone:verify; verify the charter cascade; audit policy drift; check strict violations.'
allowed-tools: Read Bash Grep Glob
metadata:
  model: opus
---

# keystone:verify — the verification gate

Pairs every claim with fresh evidence before a commit. Run each runnable
hook in this turn, then run `keystone verify`: it checks vendored policies
for drift, project files that shadow a policy's strict items, budget caps,
and host projection drift. `keystone verify` exits non-zero on a strict
violation, an exceeded cap, or projection drift.

## Run

**Run the verification hooks**: lint, type-check, test, build, drift, commit-message. Pre-commit gate. Read [`.agents/rules/process/verification.md`](../../rules/process/verification.md).

## Activities

Run every hook classified as runnable in `.agents/references/state/CODEBASE_STATE.md`, **in this turn** (no stale evidence). For each hook:

1. **lint** — [`.agents/hooks/lint.md`](../../hooks/lint.md). Invoke the project's lint command.
2. **type-check** — [`.agents/hooks/type-check.md`](../../hooks/type-check.md). Invoke the type-check command. Skip if no type checker.
3. **test** — [`.agents/hooks/test.md`](../../hooks/test.md). Invoke the test command. Fresh run.
4. **build** — [`.agents/hooks/build.md`](../../hooks/build.md). Invoke the build command.
5. **drift** — [`.agents/hooks/drift.md`](../../hooks/drift.md). Compare diff to loaded rules.
6. **commit-message** — [`.agents/hooks/commit-message.md`](../../hooks/commit-message.md). Inspect the proposed message before `git commit`.
7. **charter** — run `keystone verify` and show its output.

Show the actual tool output for each. Do not claim a hook passed without producing its output this turn.

## Iron law

**No completion claims without fresh verification evidence.** Hooks must run in this turn.

## On failure

Return to **implementation**. Do not "fix forward" inside verify. After the fix, re-invoke verify in a fresh turn.

## When to trigger

- Before committing any charter change — catches accidental shadows of
  strict policy items.
- After `keystone install` or `keystone policy update` — confirms the
  install is clean.
- In CI to gate merges on cascade integrity.

## Interpreting output

- `✓ keystone verify clean` — pass.
- `✗ N strict violation(s)` — remove the offending project file(s) or
  surface the conflict to the policy author.
- `▸ drift detected — resetting N policies` — vendored files diverged
  from the lockfile; re-run `keystone install` to restore.
