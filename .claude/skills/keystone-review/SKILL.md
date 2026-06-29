---
name: keystone-review
description: Run the four review sensors on the current diff (functional, security, risk, deployment).
tools:
  - Read
  - Grep
  - Glob
  - Bash
  - Task
model: opus
---

# keystone:review — semantic review of the current diff

Phase 5 of the six-phase lifecycle. Fires the four inferential review
sensors against the in-progress diff:

- **review-functional** — does the change implement the spec?
- **review-security** — auth, input validation, secrets, injection.
- **review-risk** — blast radius, data integrity, observability.
- **review-deployment** — migrations, rollback, config, infra deps.

Each sensor returns a structured finding list. Semantic check — runs
after computational sensors (lint / type / test / build) pass.

Canonical playbook: `.keystone/harness/actions/review.md`. Full
discipline at `.keystone/harness/guides/process/review.md`.

## Run

Open `.keystone/harness/actions/review.md` and execute every activity.

## When to trigger

- After `/keystone:verify` is clean.
- Before opening a PR.
- Anytime a change feels load-bearing and a second pass is cheap.
