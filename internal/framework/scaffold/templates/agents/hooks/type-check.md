---
kind: hook
mode: computational
event: pre-verify
run: '# TODO: wire the type-check command (see corpus/state/CODEBASE_STATE.md)'
id: type-check
description: 'Signature and contract consistency.'
---
# Sensor: type-check

Signature and contract consistency.

- **Trigger** — implementation phase (continuous), verification phase (gate).
- **Inputs** — the project's type-check command from `corpus/state/CODEBASE_STATE.md`. Skipped if the project has no type checker.
- **Exit condition** — exit code 0; no type errors.
- **Output** — pass/fail. On fail: errors as the type checker emits them.
- **State writes** — none.
