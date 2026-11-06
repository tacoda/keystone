---
kind: hook
mode: inferential
returns: review-findings
id: state-region
description: 'Read-only.'
---
# Hook: state-region

Read-only. Surfaces what is already in State for a touched region.

- **Trigger** — **orient** (planning).
- **Inputs** — current task's touched paths, `.agents/references/state/CODEBASE_STATE.md`, active migrations in `.agents/references/state/migrations/active/`.
- **Exit condition** — always succeeds; output is informational.
- **Output** — for the touched region: which idioms are loaded, coverage, last reconcile date, active migrations affecting this region.
- **State writes** — none.
