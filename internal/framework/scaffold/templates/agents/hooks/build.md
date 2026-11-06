---
kind: hook
mode: computational
event: pre-verify
run: '# TODO: wire the build command (see .agents/references/state/CODEBASE_STATE.md)'
id: build
description: 'The project''s build / compile / package step.'
---
# Hook: build

The project's build / compile / package step.

- **Trigger** — verification phase (gate).
- **Inputs** — the project's build command from `.agents/references/state/CODEBASE_STATE.md`.
- **Exit condition** — exit code 0; artifacts produced where expected.
- **Output** — pass/fail.
- **State writes** — none.
