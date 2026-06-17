# Separation of Concerns — rules

The non-negotiable and strongly-preferred rules from [`corpus/principles/separation-of-concerns.md`](../../corpus/principles/separation-of-concerns.md).
Loaded ambient; enforced by the [drift sensor](../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## GOLDEN PATH

- **Aim for boundaries that match how the system will change.** Concerns that always change together belong together; concerns that change for unrelated reasons belong apart.
- **Aim for dependencies that point toward stability.** Volatile concerns depend on stable ones, never the reverse.

---

Traces to: [`corpus/principles/separation-of-concerns.md`](../../corpus/principles/separation-of-concerns.md).
