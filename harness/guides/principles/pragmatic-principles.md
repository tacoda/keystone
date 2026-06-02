# Pragmatic Principles — rules

The non-negotiable and strongly-preferred rules from [`corpus/principles/pragmatic-principles.md`](../../corpus/principles/pragmatic-principles.md).
Loaded ambient; enforced by the [drift sensor](../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**Knowledge has exactly one home.** If a business rule, a magic number, a configuration value, or an algorithm lives in two places, the system already lies — one of the two copies is wrong, you just don't know which yet.

## GOLDEN RULES

- **Aim for code where a typical change touches one place.** Measure your designs by blast radius, not by line count.
- **Aim for the *first* fix, not the *complete* fix.** A repaired window beats a planned renovation.
- **Aim for orthogonality between layers and between features.** A change to the payment provider should not touch the shopping cart UI.

---

Traces to: [`corpus/principles/pragmatic-principles.md`](../../corpus/principles/pragmatic-principles.md).
