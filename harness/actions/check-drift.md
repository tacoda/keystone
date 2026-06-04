# check-drift

**Run the drift sensor on the current diff.** Compare the in-progress changes against loaded harness rules. Fast pre-verify check. Read [`harness/sensors/drift.md`](../sensors/drift.md) and [`harness/guides/process/implementation.md`](../guides/process/implementation.md).

## Activities

1. **Identify the diff** — `git diff`, the staged changes, or the in-progress edits.
2. **List which guides apply** — load `harness/guides/idioms/<stack>/*.md` for each touched stack, plus `harness/guides/process/implementation.md` and any active policy guides (`harness/policies/*/guides/`).
3. **Compare diff to guides** — line by line, surface violations or hot drift signals.
4. **Report findings** — one bullet per drift. Cite the guide that was violated.

## When to invoke

Between **implementation** and **verify**. Faster and cheaper than the full verify pass; meant to catch obvious drifts before sensors run.
