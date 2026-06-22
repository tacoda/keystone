---
kind: sensor
id: commit-message
description: 'Validates conventional-commit format and absence of AI attribution.'
host_triggers:
  - phase: PreToolUse
    matcher: "Bash"
    command: keystone verify --sensor commit-message
    timeout: 5
tags:
  - computational
---
# Sensor: commit-message

Validates conventional-commit format and absence of AI attribution.

- **Trigger** — release phase (final gate before `git commit`).
- **Inputs** — the staged commit message.
- **Exit condition** — message matches `<type>(<scope>): <subject>`, title under 70 chars, no mention of Claude / AI agents / co-authors / tool attribution.
- **Output** — pass/fail. On fail: the violated rule and a suggested fix.
- **State writes** — none.
