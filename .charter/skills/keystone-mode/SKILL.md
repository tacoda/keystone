---
kind: skill
id: keystone:mode
description: Switch the charter pacing mode (paired / solo / autopilot).
triggers:
  - keystone mode
  - keystone:mode
  - /keystone:mode
  - switch pacing mode
  - go autopilot
  - go solo
  - go paired
model: sonnet
tools:
  - Read
  - Write
  - Edit
---

# keystone:mode — switch pacing mode

Toggles how often the agent stops to ask. Three modes:

- **paired** — agent confirms at every gate; default for new projects.
- **solo** — agent runs through phases, asks at unclear forks.
- **autopilot** — agent runs end-to-end; asks only on hard blockers.

Canonical playbook: `.charter/actions/mode.md`. Mode definitions
live in `.charter/guides/process/modes.md`.

## Run

Open `.charter/actions/mode.md`. Decide the target mode + write
the change to `.charter/corpus/state/MODE.md`.

## When to trigger

- Start of a task: pick paired for risky work, autopilot for routine sweeps.
- Mid-task: dial in or out as confidence shifts.
