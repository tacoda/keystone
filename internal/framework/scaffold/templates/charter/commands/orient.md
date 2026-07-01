---
kind: command
id: orient
description: 'Enter the planning phase.'
---
# orient

**Enter the planning phase.** Load codebase state and matching idioms for the touched region, then sketch a plan. Read [`charter/guides/process/planning.md`](guides/process/planning.md) for the full discipline.

## Activities

1. **Read `charter/corpus/state/CODEBASE_STATE.md`** to learn:
   - Tool commands (lint, type-check, test, build)
   - Stacks present in the codebase
   - Regions and their applicable idioms
2. **Identify the touched region(s) and touched-files set** for this task. Map regions to stacks in `CODEBASE_STATE.md`; the touched-files set is the list of paths the task will read or edit.
3. **Read `charter/corpus/state/GLOBS_INDEX.md`** — the reverse-index of guide `globs:`. For each pattern in the index, check whether any touched file matches; collect the union of matched guides as the **globs-loaded set**.
4. **Load matching idioms.** For each stack the task touches, walk `charter/guides/idioms/<stack>/*.md`. Load each guide unless it declared `globs:` *and* it isn't in the globs-loaded set from step 3. (Guides without `globs:` keep today's stack-based loading; guides with `globs:` load only when their patterns matched.) Read the corresponding `charter/corpus/idioms/<stack>/<name>.md` for each loaded guide when you need the reasoning.
5. **Load active migrations** for the touched region (`charter/corpus/state/migrations/active/*.md`).
6. **Sketch a plan** — step list with a verification check per step.

## Gate

Do not begin implementation until the user has accepted the plan.

## What to *not* load

- Idioms for stacks the task does not touch.
- Idioms whose `globs:` did not match any touched file (per step 3).
- Process docs for phases this task is not currently in.

Orient exists to keep context small and relevant. Reading the entire corpus on every task is what orient prevents. The globs-index lookup tightens that further — a guide that declares `globs:` enters context only when its claimed paths are actually being touched.
