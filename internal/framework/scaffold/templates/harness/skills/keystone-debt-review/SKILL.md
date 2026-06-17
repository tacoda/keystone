---
kind: skill
id: keystone:debt-review
description: Triage the code-debt ledger — review the debt sensor's findings and update corpus/state/code-debt.md.
triggers:
  - keystone debt-review
  - keystone:debt-review
  - /keystone:debt-review
  - review code debt
  - triage debt ledger
  - debt review
---

# keystone:debt-review — triage the code-debt ledger

Reads the code-debt sensor's latest findings and reconciles them against
`.keystone/harness/corpus/state/code-debt.md`. Each finding gets one of:
keep (still real), close (resolved), defer (not now), or escalate
(promote to harness-debt or a guide).

Canonical playbook: `.keystone/harness/actions/debt-review.md`. Sensor
contract at `.keystone/harness/sensors/code-debt.md`.

## Run

Open `.keystone/harness/actions/debt-review.md` and execute every
activity.

## When to trigger

- Weekly or biweekly cadence.
- Before a major refactor — clears the slate.
- After incident response — fresh findings need triage before they rot.
