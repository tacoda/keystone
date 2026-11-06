---
kind: hook
id: lint
description: 'Surface-level style and pattern checks.'
tags:
  - computational
mode: computational
event: Stop
run: go vet ./...
---
# Hook: lint

Surface-level style and pattern checks.

- **Trigger** — implementation phase (continuous, fast feedback) and verification phase (gate).
- **Inputs** — the project's lint command from `.agents/references/state/CODEBASE_STATE.md`.
- **Exit condition** — exit code 0; no errors.
- **Output** — pass/fail. On fail: the linter's structured output passed through.
- **State writes** — none.
