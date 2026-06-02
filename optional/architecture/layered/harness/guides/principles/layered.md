# Layered Architecture — rules

The rules from [`corpus/principles/layered.md`](../../corpus/principles/layered.md).
Loaded ambient; enforced by the [drift sensor](../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**Calls go downward through layers, never upward and never skipping.** A presentation handler that invokes a repository directly has bypassed the application layer; an application service that emits an HTTP response has reached upward. Either is a violation. The whole point of layering is that each layer knows only the one beneath it.

## GOLDEN RULES

- **Aim for one layer per concern, with clear names.** Presentation, application, domain, persistence. Resist new "helper" layers that don't earn their place.
- **Aim for the application layer to own transactions.** Transactions span multiple repository calls; the application layer is the only place that knows the unit of work.
- **Aim for DTOs at layer boundaries.** Pass plain data structures across layers, not framework or ORM types. The boundary is where the dialect changes. See [[information-hiding]].
- **Aim to invert the persistence seam if the domain needs to be testable in isolation.** Decide once, apply consistently.

---

Traces to: [`corpus/principles/layered.md`](../../corpus/principles/layered.md).
