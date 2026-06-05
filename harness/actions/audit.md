# audit

**Full dual-flywheel audit.** Learning (capture from review) and Pruning (remove dead rules). Periodic. Read [`harness/README.md`](../README.md) (flywheels section) and [`harness/learning/README.md`](../learning/README.md).

## Learning flywheel

1. Walk `harness/learning/inbox/` and flag candidates that should be promoted.
2. Run review sensors on recent commits — surface patterns that should become guides.
3. Capture findings as new inbox entries (see [`learn.md`](learn.md)).

## Pruning flywheel

The Pruning flywheel reads the [harness-debt sensor](../sensors/harness-debt.md) and proposes the diff to `harness/corpus/state/harness-debt.md`. Categories the sensor surfaces:

1. **Stale rules** — guides in `harness/guides/` that no diff has touched in N months.
2. **Dead idioms** — idiom dirs whose stack is no longer in `CODEBASE_STATE.md`.
3. **Placeholders** — bootstrap `<...>` markers left unfilled.
4. **Failing sensors** — sensors recorded as available that error on invocation.
5. **Empty shells** — scaffolded dirs with no real content.
6. **Uncited policies** — installed policies whose guides were never referenced.
7. **Unresolved gaps** — `harness/adapters/<agent>/<topic>.md` TODO placeholders.
8. **Drifted state** — [stack-drift](../sensors/stack-drift.md) findings + `CODEBASE_STATE.md` stale-`last_reconciled`.
9. **Strict-cascade violations** — run `keystone policy verify`; any strict-violation output is a hard debt entry. A project file at `harness/<kind>/<name>.md` is overriding a `strict` item from a team or org policy. Remove the shadowing file or escalate to the policy owner.
10. **Required-item gaps** — same `keystone policy verify` run. Required items with no definition anywhere in the cascade are advisory debt: the project is expected to define them at `harness/<kind>/<name>.md`. Pruning surfaces them; the project decides which to fill in.

Then update the empirical state files:

11. **Risk fingerprint** — read [`harness/sensors/risk-fingerprint.md`](../sensors/risk-fingerprint.md) and update `harness/corpus/state/risk-fingerprints.md`.
12. **Traffic topology** — read [`harness/sensors/traffic-topology.md`](../sensors/traffic-topology.md) and update `harness/corpus/state/traffic-topology.md`.

## Output

One report with two sections (Learn / Prune), each listing concrete proposed harness edits. Pruning's diffs land in `corpus/state/harness-debt.md`. Propose every state-file diff before applying; do not silently overwrite.
