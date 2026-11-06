---
kind: skill
id: keystone:spec
description: Capture intent + acceptance criteria for a unit of work. First action on any task.
triggers:
  - keystone spec
  - keystone:spec
  - /keystone:spec
  - capture acceptance criteria
  - spec this task
  - what are we building
model: opus
tools:
  - Read
  - Write
  - Edit
  - Grep
  - Glob
---

# keystone:spec — capture intent + AC

Phase 1 of the six-phase lifecycle. Forces the agent to write down what
"done" looks like before any code is touched. Output lands in the task
notes / PR description / equivalent — acceptance criteria the agent and
user both sign off on.

Canonical playbook: `.charter/actions/spec.md`. Full discipline
at `.charter/guides/process/spec.md`.

## Run

Open `.charter/actions/spec.md` and execute every activity.

## When to trigger

- Start of every task. Non-negotiable on anything beyond a typo fix.
- When mid-task scope drift makes the original intent stale — re-spec.
