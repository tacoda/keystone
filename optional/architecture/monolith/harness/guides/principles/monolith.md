# Modular Monolith — rules

The rules from [`corpus/principles/monolith.md`](../../corpus/principles/monolith.md).
Loaded ambient; enforced by the [drift sensor](../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**Modules respect each other's public APIs.** A module that reaches into another module's internals — a database table, a private type, a tightly-coupled helper — has dissolved the boundary. The compiler, linter, or code-review process must reject the call. Boundaries enforced only by convention will be crossed under deadline.

## GOLDEN RULES

- **Aim for module boundaries you'd draw as service boundaries.** The eventual split (if it happens) should be cheap because the seams already exist.
- **Aim for in-process calls between modules, not network calls.** That is the monolith's structural advantage — use it.
- **Aim for one database, with table ownership per module.** A shared database with respected ownership is a feature, not a bug; what microservices forbid for distribution reasons doesn't apply in-process.
- **Aim for a clear deployment story.** One deployable means one rollback target — but the deploy must be safe and frequent. A scary deploy is a sign the monolith has grown beyond its discipline.
- **Aim to split a module to a service only when there is *evidence*** — independent scale needs, team-ownership conflicts, regulatory isolation. Splits on speculation are reversed within a year.

---

Traces to: [`corpus/principles/monolith.md`](../../corpus/principles/monolith.md).
