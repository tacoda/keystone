---
kind: computational
---

# Sensor: commit-message

Validates conventional-commit format and absence of AI attribution.

- **Trigger** — release phase (final gate before `git commit`).
- **Inputs** — the staged commit message.
- **Exit condition** — message matches `<type>(<scope>): <subject>`, title under 70 chars, no mention of Claude / AI agents / co-authors / tool attribution.
- **Output** — pass/fail. On fail: the violated rule and a suggested fix.
- **State writes** — none.
