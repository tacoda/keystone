---
name: keystone:audit
description: Full dual-flywheel audit — Learning (capture from review) and Pruning (remove dead rules)
---

You are running the **audit** action. Read `harness/README.md` (flywheels section) and `harness/learning/README.md`.

## Activities

### Learning flywheel

1. Walk `harness/learning/inbox/` and flag candidates that should be promoted.
2. Spawn relevant review sub-agents on recent commits — surface patterns that should become guides.
3. Capture findings as new inbox entries (see the **learn** action).

### Pruning flywheel

1. **Stale rules** — list guides in `harness/guides/` that no diff has touched in N months (read `git log -- harness/guides/`). Flag for review.
2. **Dead idioms** — list idioms in `harness/corpus/idioms/<stack>/` whose stack is no longer present in `CODEBASE_STATE.md`.
3. **Risk fingerprint** — read `harness/sensors/risk-fingerprint.md` and update `harness/corpus/state/risk-fingerprints.md`.
4. **Traffic topology** — read `harness/sensors/traffic-topology.md` and update `harness/corpus/state/traffic-topology.md`.

## Output

A single report with two sections (Learn / Prune), each listing concrete proposed harness edits. Propose every state file diff via Edit; do not silently overwrite.
