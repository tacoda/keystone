---
kind: hook
id: coverage
description: 'Reads test coverage and updates the State layer.'
tags:
  - computational
mode: computational
event: Stop
run: "true"
---
# Hook: coverage

Reads test coverage and updates the State layer.

- **Trigger** — verification phase (proposes state update), **audit**.
- **Inputs** — the project's coverage command from `.agents/references/state/CODEBASE_STATE.md`. Skipped if no coverage tool is configured.
- **Exit condition** — coverage report produced. No minimum threshold by default — projects may set one in `.agents/references/state/CODEBASE_STATE.md`.
- **Output** — coverage stats per region.
- **State writes** — proposes a diff to `.agents/references/state/CODEBASE_STATE.md` updating coverage per region. User accepts or edits.
