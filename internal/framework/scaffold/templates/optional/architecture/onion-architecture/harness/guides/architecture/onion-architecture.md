# Onion Architecture — rules

The rules from [`corpus/onion-architecture.md`](../corpus/onion-architecture.md).
Loaded ambient; enforced by the [drift sensor](../../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**The domain model is the center, and it points nowhere.** Domain types must import nothing from any outer ring — not the database, not the framework, not the logger, not the clock. Time, persistence, and side effects enter the domain through interfaces the domain owns.

## GOLDEN RULE

- **Aim for an explicit application-services layer.** Don't collapse it into either the domain or the controllers. The orchestration concern is real; give it a home.
- **Aim for domain services on operations that span entities.** A "domain service" is not "any service" — it is a domain concept that doesn't belong to one entity. Don't dilute the term.
- **Aim for interfaces owned by the inner ring that needs them.** The domain owns its `Clock`; the application service owns its `UserRepository`. The infrastructure implements both.
- **Aim for testability inward.** Each ring should be testable with the next ring inward as a real dependency and the next ring outward as a fake or interface.

---

Traces to: [`corpus/onion-architecture.md`](../corpus/onion-architecture.md).
