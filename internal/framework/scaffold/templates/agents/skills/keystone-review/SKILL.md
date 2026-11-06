---
name: keystone-review
description: 'Run the four review hooks on the current diff (functional, security, risk, deployment). Triggers: keystone review; keystone:review; /keystone:review; review the diff; run review hooks; semantic review.'
---

# keystone:review — semantic review of the current diff

Phase 5 of the six-phase lifecycle. Fires the four inferential review
hooks against the in-progress diff:

- **review-functional** — does the change implement the spec?
- **review-security** — auth, input validation, secrets, injection.
- **review-risk** — blast radius, data integrity, observability.
- **review-deployment** — migrations, rollback, config, infra deps.

Each hook returns a structured finding list. Semantic check — runs
after computational hooks (lint / type / test / build) pass.

Canonical command: `.agents/commands/review.md`. Full
discipline at `.agents/rules/process/review.md`.

## Run

Open `.agents/commands/review.md` and execute every activity.

## When to trigger

- After `/keystone:verify` is clean.
- Before opening a PR.
- Anytime a change feels load-bearing and a second pass is cheap.
