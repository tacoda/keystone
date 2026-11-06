# Continuous Delivery — rules

The rules from [`corpus/continuous-delivery.md`](../corpus/continuous-delivery.md).
Loaded ambient; enforced by the [drift sensor](../../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**Main is always shippable.** Every commit on the trunk must produce an artifact that could be deployed to production right now. A commit that breaks the pipeline, breaks the trunk, or leaves the system in a half-built state is reverted, not "fixed later." The asymmetry is intentional: the cost of reverting is small; the cost of keeping main broken is paid by every subsequent commit until the fix lands.

## GOLDEN RULE

- **Aim to release small.** Many small releases beat few large ones — smaller blast radius, faster diagnosis, easier rollback.
- **Aim to remove the deploy ceremony.** Deploying should be a non-event. If the team holds a meeting before a deploy, the deploy is rarer than it should be.
- **Aim to deploy on Friday afternoon.** The classic shibboleth: a team that won't deploy late on a Friday has insufficient confidence in its pipeline. Build the confidence; remove the taboo.
- **Aim to roll forward, not back.** Most bugs in production are faster to fix forward than to roll back, *when* the pipeline is fast and the deploys are small. Reserve rollbacks for incidents.
- **Aim for the four DORA metrics as health indicators**: deployment frequency, lead time for changes, change failure rate, mean time to recover. See [[observability]] — what gets measured stays alive.

---

Traces to: [`corpus/continuous-delivery.md`](../corpus/continuous-delivery.md).
