# audit

**Full dual-flywheel audit.** Learning (capture from review) and Pruning (remove dead rules). Periodic. Read [`harness/README.md`](../README.md) (flywheels section) and [`harness/learning/README.md`](../learning/README.md).

## Learning flywheel

1. Walk `harness/learning/inbox/` and flag candidates that should be promoted.
2. Run review sensors on recent commits — surface patterns that should become guides.
3. Capture findings as new inbox entries (see [`learn.md`](learn.md)).

## Pruning flywheel

1. **Stale rules** — list guides in `harness/guides/` that no diff has touched in N months (`git log -- harness/guides/`). Flag for review.
2. **Dead idioms** — list idioms in `harness/corpus/idioms/<stack>/` whose stack is no longer present in `CODEBASE_STATE.md`.
3. **Risk fingerprint** — read [`harness/sensors/risk-fingerprint.md`](../sensors/risk-fingerprint.md) and update `harness/corpus/state/risk-fingerprints.md`.
4. **Traffic topology** — read [`harness/sensors/traffic-topology.md`](../sensors/traffic-topology.md) and update `harness/corpus/state/traffic-topology.md`.

## Output

One report with two sections (Learn / Prune), each listing concrete proposed harness edits. Propose every state-file diff before applying; do not silently overwrite.
