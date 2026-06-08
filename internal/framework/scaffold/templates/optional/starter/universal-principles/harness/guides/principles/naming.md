# Naming — rules

The non-negotiable and strongly-preferred rules from [`corpus/principles/naming.md`](../../corpus/principles/naming.md).
Loaded ambient; enforced by the [drift sensor](../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**Names are part of the contract.** A function, class, or module that does more or less than its name implies has lied — and a reader who acted on the lie will be wrong. If the name no longer fits the behavior, change one of them. See [[least-astonishment]] (names as promises) and [[hyrums-law]] (callers depend on observable properties; the name is one of them).

## GOLDEN RULES

- **Aim for names that survive review without a glossary.** A reviewer who has never seen the code should be able to guess what each name means.
- **Aim for the most specific accurate name.** Not the shortest, not the longest — the most *informative*. `fetchActiveUsers` beats `fetch` and beats `getActiveUsersFromDatabaseAndCacheThem`.
- **Aim for names that match domain language exactly.** If the business says *invoice*, the code does not say *bill*, *receipt*, or *Statement*.
- **Aim to rename early.** A bad name caught in the first week is one rename; caught after a year, it is a multi-file change with a long review.

---

Traces to: [`corpus/principles/naming.md`](../../corpus/principles/naming.md).
