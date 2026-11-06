---
name: keystone-audit
description: 'Full dual-flywheel audit — Learning (capture from review) + Pruning (remove dead rules). Periodic charter hygiene. Triggers: keystone audit; keystone:audit; /keystone:audit; run audit; audit the charter; dual flywheel audit.'
allowed-tools: Read Grep Glob Bash Task
metadata:
  model: opus
---

# keystone:audit — periodic dual-flywheel review

The Pruning flywheel's main entry point, paired with one Learning
sweep. Walks the references + rules looking for stale, dead, or unused
content; proposes archive moves with a one-line reason per item.

Canonical command: `.agents/commands/audit.md`. Read it for
the 12-category Pruning checklist (stale rules, dead idioms,
placeholders, failing hooks, empty shells, uncited policies,
unresolved gaps, drifted state, strict-cascade violations, required-item
gaps, risk fingerprint, traffic topology).

## Run

Open `.agents/commands/audit.md` and execute every section.
Output is one report with Learn / Prune sections, each listing concrete
proposed charter edits. Pruning diffs land in
`.agents/references/state/charter-debt.md`.

## When to trigger

- Monthly cadence (recommended).
- After any major refactor or stack change.
- When the dashboard's debt count starts to feel wrong.

## Asymmetry

Rules churn often; references rarely. Audit reflects that: rule proposals
are routine, reference proposals are rare and load-bearing.
