---
name: keystone-audit
description: 'Full dual-flywheel audit — Learning (capture from review) + Pruning (remove dead rules). Periodic charter hygiene. Triggers: keystone audit; keystone:audit; /keystone:audit; run audit; audit the charter; dual flywheel audit.'
---

# keystone:audit — periodic dual-flywheel review

The Pruning flywheel's main entry point, paired with one Learning
sweep. Walks the references + rules looking for stale, dead, or unused
content; proposes archive moves with a one-line reason per item.

Read it for
the 12-category Pruning checklist (stale rules, dead idioms,
placeholders, failing hooks, empty shells, uncited policies,
unresolved gaps, drifted state, strict-cascade violations, required-item
gaps, risk fingerprint, traffic topology).

## Run

**Full dual-flywheel audit.** Learning (capture from review) and Pruning (remove dead rules). Periodic. Read [`.agents/README.md`](../../README.md) (flywheels section) and [`.keystone/learning/README.md`](../../../.keystone/learning/README.md).

## Learning flywheel

1. Walk `.keystone/learning/inbox/` and flag candidates that should be promoted.
2. Run the review hooks on recent commits. Surface patterns that should become rules.
3. Capture findings as new inbox entries (see [`learn.md`](../keystone-learn/SKILL.md)).

## Pruning flywheel

The Pruning flywheel reads the [charter-debt hook](../../hooks/charter-debt.md) and proposes the diff to `.agents/references/state/charter-debt.md`. Categories the hook surfaces:

1. **Stale rules**: rules in `.agents/rules/` that no diff has touched in N months.
2. **Dead idioms** — idiom dirs whose stack is no longer in `CODEBASE_STATE.md`.
3. **Placeholders** — bootstrap `<...>` markers left unfilled.
4. **Failing hooks**: hooks recorded as available that error on invocation.
5. **Empty shells** — scaffolded dirs with no real content.
6. **Uncited policies**: installed policies whose rules were never referenced.
7. **Drifted state** — [stack-drift](../../hooks/stack-drift.md) findings + `CODEBASE_STATE.md` stale-`last_reconciled`.
8. **Strict-cascade violations** — run `keystone policy verify`; any strict-violation output is a hard debt entry. A project file at `.agents/<kind>/<name>.md` is overriding a `strict` item from a team or org policy. Remove the shadowing file or escalate to the policy owner.
9. **Required-item gaps** — same `keystone policy verify` run. Required items with no definition anywhere in the cascade are advisory debt: the project is expected to define them at `.agents/<kind>/<name>.md`. Pruning surfaces them; the project decides which to fill in.

Then update the empirical state files:

10. **Risk fingerprint** — read [`.agents/hooks/risk-fingerprint.md`](../../hooks/risk-fingerprint.md) and update `.agents/references/state/risk-fingerprints.md`.
11. **Traffic topology** — read [`.agents/hooks/traffic-topology.md`](../../hooks/traffic-topology.md) and update `.agents/references/state/traffic-topology.md`.

## Output

One report with two sections (Learn / Prune), each listing concrete proposed charter edits. Pruning's diffs land in `.agents/references/state/charter-debt.md`. Propose every state-file diff before applying; do not silently overwrite.

## Paths in audit

Audit walks the full repo, but each rule's `paths:` still filter findings. A rule reports only against files it claims (see [`hooks/drift.md`](../../hooks/drift.md)). Audit is the only command where most rules match somewhere. Per-rule `paths:` keep the report focused on real violations.

## Index refresh

After any archive move or content change, run `keystone index` so `.keystone/INDEX.json` no longer references the archived primitive. The `keystone-index` skill wraps the CLI invocation. The dashboard's `/prune` view picks up the same heuristics this command surfaces. Use it to spot-check findings between audits.

## When to trigger

- Monthly cadence (recommended).
- After any major refactor or stack change.
- When the dashboard's debt count starts to feel wrong.

## Asymmetry

Rules churn often; references rarely. Audit reflects that: rule proposals
are routine, reference proposals are rare and load-bearing.
