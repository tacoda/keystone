---
description: Full dual-flywheel audit — Learning (capture from review) and Pruning (remove dead rules)
---

Run the **audit** action.

Read `harness/README.md` (flywheels section) and `harness/learning/README.md`.

## Learning flywheel

1. Walk `harness/learning/inbox/` and flag candidates for promotion.
2. Surface patterns from recent commits that should become guides; capture as new inbox entries.

## Pruning flywheel

1. **Stale rules** — flag guides untouched for N months (`git log -- harness/guides/`).
2. **Dead idioms** — flag stacks no longer present in `CODEBASE_STATE.md`.
3. **Risk fingerprint** — refresh `harness/corpus/state/risk-fingerprints.md` per `harness/sensors/risk-fingerprint.md`.
4. **Traffic topology** — refresh `harness/corpus/state/traffic-topology.md` per `harness/sensors/traffic-topology.md`.

Output one report with Learn / Prune sections. No silent overwrites.
