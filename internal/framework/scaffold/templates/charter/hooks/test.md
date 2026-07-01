---
kind: hook
mode: computational
event: pre-verify
run: '# TODO: wire the test command (see corpus/state/CODEBASE_STATE.md)'
id: test
description: 'The project''s test suite.'
---
# Sensor: test

The project's test suite.

- **Trigger** — implementation phase (after each green during TDD), verification phase (gate).
- **Inputs** — the project's test command from `corpus/state/CODEBASE_STATE.md`.
- **Exit condition** — exit code 0; 0 failures.
- **Output** — pass/fail with failure summary. Stale evidence does not count — see `guides/process/verification.md`'s IRON LAW.
- **State writes** — none directly; the [coverage](coverage.md) sensor reads test artifacts.
