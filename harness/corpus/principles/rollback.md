# Rollback

A deploy is one half of a change; the other half is the path back. Changes that ship without a rehearsed way home assume nothing will go wrong — an assumption the historical record does not support. The discipline of rollback is the discipline of building the return path *with* the change, not *after* the change has failed.

> **Rules extracted:** [`guides/principles/rollback.md`](../../guides/principles/rollback.md).

## What it asks of you

- Decouple *deployment* from *release*. Deploying code is a technical act; releasing capability is a product act. A feature flag separates them. A change can be deployed without being released, and released without being re-deployed.
- Design the rollback as part of the design. The PR that adds a capability includes the path to remove it. If the path is harder to think about than the addition, that is signal: the change is too coupled to existing state.
- Treat the rollback path as *real code*. Rehearse it. Stale rollback procedures are an incident waiting for an excuse.

## Why it holds

Humble & Farley's *Continuous Delivery* (2010) places rollback at the center of release engineering: a system where every change can be undone in minutes is a system where *teams take more risks*, not fewer, because the cost of being wrong is bounded. The opposite — where rollback is "restore from backup over the weekend" — is a system where teams ship rarely, conservatively, and in large batches, which is the exact pattern that *increases* the rate of incidents (Forsgren, Humble & Kim, *Accelerate*, 2018).

Feature flags (Hodgson, Fowler 2017) generalize the rollback path to *runtime*: the change is in production, off; turning it on is the release; turning it off is the rollback. This works as long as the flagged code does not entangle itself with un-flagged code — the "long-lived flag" pattern is a documented anti-pattern precisely because entanglement grows.

The blue-green pattern (Fowler 2010) is the same idea at the infrastructure level: deploy the new version next to the old, switch traffic, keep the old running until confidence is established. Rollback is a traffic switch, not a redeploy.

## Anti-patterns

- A migration with no rollback strategy beyond "restore from a backup that we hope works."
- A feature flag that has been on for 100% of traffic for six months. Either remove the flag (the code is now part of the system) or remove the feature (the flag is hiding dead code).
- A rollback procedure that has never been exercised in staging.
- "Forward-only" as a policy. Forward-only is a description of what happened when rollback failed; it is not a strategy.
- Coupling a deploy to a one-way data migration. The data change makes the rollback irreversible; the code change should not depend on the data change being permanent yet.

## References

- Humble, J. & Farley, D. *Continuous Delivery* (2010).
- Forsgren, N., Humble, J. & Kim, G. *Accelerate* (2018).
- Hodgson, P. [*Feature Toggles (aka Feature Flags)*](https://martinfowler.com/articles/feature-toggles.html) (2017).
- Fowler, M. [*BlueGreenDeployment*](https://martinfowler.com/bliki/BlueGreenDeployment.html) (2010).
- Allspaw, J. & the Resilience Engineering literature — the discipline of return paths.

---

Forward link: [`guides/principles/rollback.md`](../../guides/principles/rollback.md). See also: [`migrations.md`](migrations.md), [`modern-software-engineering.md`](modern-software-engineering.md).
