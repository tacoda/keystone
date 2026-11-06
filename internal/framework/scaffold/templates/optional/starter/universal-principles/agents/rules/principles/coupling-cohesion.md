# Coupling and Cohesion — rules

The non-negotiable and strongly-preferred rules from [`references/principles/coupling-cohesion.md`](../../references/principles/coupling-cohesion.md).
Loaded ambient; enforced by the [drift hook](../../hooks/drift.md). The reference file holds the full reasoning, anti-patterns, and references.

## GOLDEN RULE

- **Aim for the smallest seam.** When two modules must communicate, pass only what is needed — not the whole object.
- **Aim for single-purpose modules.** The cohesion test: can you describe what this module does in one sentence without using "and"?
- **Aim for dependencies that point in one direction.** Cycles between modules guarantee that any change propagates.

---

Traces to: [`references/principles/coupling-cohesion.md`](../../references/principles/coupling-cohesion.md).
