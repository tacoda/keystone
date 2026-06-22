# Coupling and Cohesion — rules

The non-negotiable and strongly-preferred rules from [`corpus/principles/coupling-cohesion.md`](../../corpus/principles/coupling-cohesion.md).
Loaded ambient; enforced by the [drift sensor](../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## GOLDEN RULE

- **Aim for the smallest seam.** When two modules must communicate, pass only what is needed — not the whole object.
- **Aim for single-purpose modules.** The cohesion test: can you describe what this module does in one sentence without using "and"?
- **Aim for dependencies that point in one direction.** Cycles between modules guarantee that any change propagates.

---

Traces to: [`corpus/principles/coupling-cohesion.md`](../../corpus/principles/coupling-cohesion.md).
