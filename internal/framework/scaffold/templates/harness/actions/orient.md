# orient

**Enter the planning phase.** Load codebase state and matching idioms for the touched region, then sketch a plan. Read [`harness/guides/process/planning.md`](../guides/process/planning.md) for the full discipline.

## Activities

1. **Read `harness/corpus/state/CODEBASE_STATE.md`** to learn:
   - Tool commands (lint, type-check, test, build)
   - Stacks present in the codebase
   - Regions and their applicable idioms
2. **Identify the touched region(s)** for this task. Map them to stacks in `CODEBASE_STATE.md`.
3. **Load matching idioms** by reading `harness/corpus/idioms/<stack>/*.md` for each stack the task touches.
4. **Load active migrations** for the touched region (`harness/corpus/state/migrations/active/*.md`).
5. **Sketch a plan** — step list with a verification check per step.

## Gate

Do not begin implementation until the user has accepted the plan.

## What to *not* load

- Idioms for stacks the task does not touch.
- Process docs for phases this task is not currently in.

Orient exists to keep context small and relevant. Reading the entire corpus on every task is what orient prevents.
