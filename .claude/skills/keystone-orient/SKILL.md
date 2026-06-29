---
name: keystone-orient
description: Enter the planning phase — load codebase state + matching idioms for the touched region, then sketch a plan.
tools:
  - Read
  - Grep
  - Glob
  - Bash
  - Task
model: opus
---

# keystone:orient — planning phase

Phase 2 of the six-phase lifecycle. Loads
`.keystone/harness/corpus/state/CODEBASE_STATE.md`, the matching idioms
for the touched region from `corpus/idioms/<stack>/`, and the relevant
process guides. Output is a written plan the agent and user agree on
before implementation starts.

Canonical playbook: `.keystone/harness/actions/orient.md`. Full
discipline at `.keystone/harness/guides/process/planning.md`.

## Run

Open `.keystone/harness/actions/orient.md` and execute every activity.

## When to trigger

- After `/keystone:spec` lands acceptance criteria.
- Before writing any code on a non-trivial change.
- When the agent loses the plot mid-implementation — re-orient and
  resync.
