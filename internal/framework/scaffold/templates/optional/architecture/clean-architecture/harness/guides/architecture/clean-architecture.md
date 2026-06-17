# Clean Architecture — rules

The rules from [`corpus/clean-architecture.md`](../corpus/clean-architecture.md).
Loaded ambient; enforced by the [drift sensor](../../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**No imports go outward.** A grep of every inner-ring file for an outer-ring identifier must return zero results. The rule is not "minimize" — it is **zero**. If you cannot enforce it with directory structure, package visibility, or a linter, the architecture is decorative.

## GOLDEN PATH

- **Aim for inner rings that compile and test without any outer ring.** The entities and use cases should build into a library with no framework or database dependencies.
- **Aim for interfaces owned by the consumer.** The use case owns the `UserRepository` interface; the database adapter implements it. The interface lives in the inner ring.
- **Aim for the outermost ring to be as thin as possible.** Frameworks change; controllers should not contain business logic that has to change with them.
- **Aim for crossing the boundary with data structures, not framework types.** The boundary is where the dialect changes.

---

Traces to: [`corpus/clean-architecture.md`](../corpus/clean-architecture.md).
