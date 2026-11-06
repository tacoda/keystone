---
kind: hook
mode: computational
event: pre-verify
run: '# TODO: wire the lint command (see corpus/state/CODEBASE_STATE.md)'
id: lint
description: 'Surface-level style and pattern checks.'
---
# Sensor: lint

Surface-level style and pattern checks.

- **Trigger** — implementation phase (continuous, fast feedback) and verification phase (gate).
- **Inputs** — the project's lint command from `corpus/state/CODEBASE_STATE.md`.
- **Exit condition** — exit code 0; no errors.
- **Output** — pass/fail. On fail: the linter's structured output passed through.
- **State writes** — none.
