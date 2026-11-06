# Separation of Concerns — rules

The non-negotiable and strongly-preferred rules from [`references/principles/separation-of-concerns.md`](../../references/principles/separation-of-concerns.md).
Loaded ambient; enforced by the [drift hook](../../hooks/drift.md). The reference file holds the full reasoning, anti-patterns, and references.

## GOLDEN RULE

- **Aim for boundaries that match how the system will change.** Concerns that always change together belong together; concerns that change for unrelated reasons belong apart.
- **Aim for dependencies that point toward stability.** Volatile concerns depend on stable ones, never the reverse.

---

Traces to: [`references/principles/separation-of-concerns.md`](../../references/principles/separation-of-concerns.md).
