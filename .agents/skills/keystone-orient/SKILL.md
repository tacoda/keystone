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

Canonical command: `.agents/commands/orient.md`. Full
discipline at `.agents/rules/process/planning.md`.

## Run

Open `.agents/commands/orient.md` and execute every activity.

## When to trigger

- After `/keystone:spec` lands acceptance criteria.
- Before writing any code on a non-trivial change.
- When the agent loses the plot mid-implementation — re-orient and
  resync.
