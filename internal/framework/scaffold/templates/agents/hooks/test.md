---
kind: hook
mode: computational
event: pre-verify
run: '# TODO: wire the test command (see .agents/references/state/CODEBASE_STATE.md)'
id: test
description: 'The project''s test suite.'
---
# Hook: test

The project's test suite.

- **Trigger** — implementation phase (after each green during TDD), verification phase (gate).
- **Inputs** — the project's test command from `.agents/references/state/CODEBASE_STATE.md`.
- **Exit condition** — exit code 0; 0 failures.
- **Output** — pass/fail with failure summary. Stale evidence does not count — see `.agents/rules/process/verification.md`'s IRON LAW.
- **State writes** — none directly; the [coverage](coverage.md) hook reads test artifacts.
