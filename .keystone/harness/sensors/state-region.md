---
kind: sensor
id: state-region
description: 'Read-only.'
host_triggers:
  - phase: PostToolUse
    matcher: "Edit|Write"
    command: true
    timeout: 5
---
# Sensor: state-region

Read-only. Surfaces what is already in State for a touched region.

- **Trigger** — **orient** (planning).
- **Inputs** — current task's touched paths, `corpus/state/CODEBASE_STATE.md`, active migrations in `corpus/state/migrations/active/`.
- **Exit condition** — always succeeds; output is informational.
- **Output** — for the touched region: which idioms are loaded, coverage, last reconcile date, active migrations affecting this region.
- **State writes** — none.
