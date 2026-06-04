---
name: keystone:check-drift
description: Run the drift sensor — compare the current diff against loaded harness rules
---

Run the **drift sensor** on the current diff.

Read `harness/sensors/drift.md` and `harness/guides/process/implementation.md`.

## Activities

1. **Identify the diff** — `git diff`, or the staged changes, or the in-progress edits.
2. **List which guides apply** — load `harness/guides/idioms/<stack>/*.md` for each touched stack, plus `harness/guides/process/implementation.md`.
3. **Compare diff to guides** — line-by-line, surface violations or hot drift signals.
4. **Report findings** — one bullet per drift. Cite the guide that was violated.

## When to invoke

Between **implementation** and **verify**. Faster and cheaper than the full verify; meant to catch obvious drifts before sensors run.
