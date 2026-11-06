---
name: keystone-orient
description: 'Enter the planning phase — load codebase state + matching idioms for the touched region, then sketch a plan. Triggers: keystone orient; keystone:orient; /keystone:orient; enter planning phase; orient on this task; load idioms for the touched region.'
allowed-tools: Read Grep Glob Bash Task
metadata:
  model: opus
---

# keystone:orient — planning phase

Phase 2 of the six-phase lifecycle. Loads
`.agents/references/state/CODEBASE_STATE.md`, the matching idioms
for the touched region from `.agents/references/idioms/<stack>/`, and the relevant
process rules. Output is a written plan the agent and user agree on
before implementation starts.

Full
discipline at `.agents/rules/process/planning.md`.

## Run

**Enter the planning phase.** Load codebase state and matching idioms for the touched region, then sketch a plan. Read [`.agents/rules/process/planning.md`](../../rules/process/planning.md) for the full discipline.

## Activities

1. **Read `.agents/references/state/CODEBASE_STATE.md`** to learn:
   - Tool commands (lint, type-check, test, build)
   - Stacks present in the codebase
   - Regions and their applicable idioms
2. **Identify the touched region(s) and touched-files set** for this task. Map regions to stacks in `CODEBASE_STATE.md`; the touched-files set is the list of paths the task will read or edit.
3. **Read `.agents/references/state/GLOBS_INDEX.md`** — the reverse-index of rule `paths:`. For each pattern in the index, check whether any touched file matches; collect the union of matched rules as the **globs-loaded set**.
4. **Load matching idioms.** For each stack the task touches, walk `.agents/rules/idioms/<stack>/*.md`. Load each rule unless it declared `paths:` *and* it isn't in the globs-loaded set from step 3. (Rules without `paths:` keep stack-based loading; rules with `paths:` load only when their patterns matched.) Read the corresponding `.agents/references/idioms/<stack>/<name>.md` for each loaded rule when you need the reasoning.
5. **Load active migrations** for the touched region (`.agents/references/state/migrations/active/*.md`).
6. **Sketch a plan** — step list with a verification check per step.

## Gate

Do not begin implementation until the user has accepted the plan.

## What to *not* load

- Idioms for stacks the task does not touch.
- Idioms whose `paths:` did not match any touched file (per step 3).
- Process docs for phases this task is not currently in.

Orient exists to keep context small and relevant. Reading every reference on every task is what orient prevents. The globs-index lookup tightens that further: a rule that declares `paths:` enters context only when its claimed paths are actually being touched.

## When to trigger

- After `/keystone:spec` lands acceptance criteria.
- Before writing any code on a non-trivial change.
- When the agent loses the plot mid-implementation — re-orient and
  resync.
