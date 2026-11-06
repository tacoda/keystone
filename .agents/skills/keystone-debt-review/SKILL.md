---
name: keystone-debt-review
description: 'Triage the code-debt ledger — review the code-debt findings and update .agents/references/state/code-debt.md. Triggers: keystone debt-review; keystone:debt-review; /keystone:debt-review; review code debt; triage debt ledger; debt review.'
allowed-tools: Read Grep Glob Bash Write Edit
metadata:
  model: opus
---

# keystone:debt-review — triage the code-debt ledger

Reads the code-debt subagent's latest findings and reconciles them against
`.agents/references/state/code-debt.md`. Each finding gets one of:
keep (still real), close (resolved), defer (not now), or escalate
(promote to charter-debt or a rule).

Canonical command: `.agents/commands/debt-review.md`. Subagent
contract at `.agents/agents/code-debt.md`.

## Run

Open `.agents/commands/debt-review.md` and execute every
activity.

## When to trigger

- Weekly or biweekly cadence.
- Before a major refactor — clears the slate.
- After incident response — fresh findings need triage before they rot.
