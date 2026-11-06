---
kind: skill
id: keystone:audit
description: Full dual-flywheel audit — Learning (capture from review) + Pruning (remove dead rules). Periodic charter hygiene.
triggers:
  - keystone audit
  - keystone:audit
  - /keystone:audit
  - run audit
  - audit the charter
  - dual flywheel audit
---

# keystone:audit — periodic dual-flywheel review

The Pruning flywheel's main entry point, paired with one Learning
sweep. Walks the corpus + guides looking for stale, dead, or unused
content; proposes archive moves with a one-line reason per item.

Canonical playbook: `.charter/actions/audit.md`. Read it for
the 12-category Pruning checklist (stale rules, dead idioms,
placeholders, failing sensors, empty shells, uncited policies,
unresolved gaps, drifted state, strict-cascade violations, required-item
gaps, risk fingerprint, traffic topology).

## Run

Open `.charter/actions/audit.md` and execute every section.
Output is one report with Learn / Prune sections, each listing concrete
proposed charter edits. Pruning diffs land in
`.charter/corpus/state/charter-debt.md`.

## When to trigger

- Monthly cadence (recommended).
- After any major refactor or stack change.
- When the dashboard's debt count starts to feel wrong.

## Asymmetry

Guides churn often; corpus rarely. Audit reflects that — guide proposals
are routine, corpus proposals are rare and load-bearing.
