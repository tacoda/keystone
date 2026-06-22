---
kind: sensor
id: build
description: 'The project''s build / compile / package step.'
host_triggers:
  - phase: Stop
    command: go build ./...
    timeout: 120
---
# Sensor: build

The project's build / compile / package step.

- **Trigger** — verification phase (gate).
- **Inputs** — the project's build command from `corpus/state/CODEBASE_STATE.md`.
- **Exit condition** — exit code 0; artifacts produced where expected.
- **Output** — pass/fail.
- **State writes** — none.
