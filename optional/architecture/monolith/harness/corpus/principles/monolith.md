# Modular Monolith

A single deployable application, organized internally into **strongly-bounded modules** that respect each other's interfaces. Articulated as a distinct pattern by Simon Brown ("Modular Monoliths", 2015) and championed at scale by Shopify, Basecamp, and GitHub. The modular monolith is the **counterpattern** to premature microservices — it captures most of what teams *think* they want from microservices (clear boundaries, independent reasoning) without the operational cost of a distributed system.

The phrase emphasizes both halves. *Monolith* — one deployable, one process, one database (or carefully shared persistence), the in-process calls fast and reliable. *Modular* — the internal module boundaries are real, enforced by language features or code review, and respected by every team.

> **Rules extracted:** [`guides/principles/monolith.md`](../../guides/principles/monolith.md). This file holds the full reasoning, anti-patterns, and references.

## What "modular" requires

A codebase organized into folders is *not* a modular monolith. The modules must be **enforced boundaries**:

- **Explicit public API per module.** Every module exposes a small surface; everything else is implementation. Other modules call only the public surface. See [[information-hiding]].
- **No circular dependencies.** Module A may call module B, or B may call A, never both. The dependency graph is a DAG.
- **Mechanical enforcement.** Internal packages (Go), `internal` visibility (Rust), package-private classes (Java), ESLint boundary rules (TypeScript), Sorbet/Rubocop boundary checks (Ruby/Shopify). The boundary the compiler enforces is the boundary that survives a rushed deadline.
- **Data ownership per module.** Each module owns its tables. Other modules query through the owning module's API, not by joining tables. The "shared database" smell from [[microservices]] applies here too.

When these are honored, the monolith *splits cleanly* later if scale demands. When they are not, the monolith becomes a "big ball of mud" — the architecture that has no architecture.

## When the monolith is the right answer

- The team is small enough that one deploy serves everyone (roughly: under 50 engineers shipping to the same codebase). See [[microservices]] golden rules — service count should be staffed.
- The product is still finding fit. Boundaries chosen too early ossify in the wrong places.
- Operational maturity (observability, paging, runbooks) is not in place yet. A distributed system pre-mature for the team that runs it produces incident frequency the team cannot keep up with.
- Local-call latency matters. A monolith does in microseconds what a distributed call does in milliseconds; multiply by request fanout. See [[distributed-systems-fallacies]] (latency is zero).

The most successful microservice migrations start from a healthy modular monolith. The least successful start from a hurried decision to "go microservices from day one."

## What it asks of you

- When you draw module boundaries, draw them along the same axis you would draw service boundaries: **business capabilities**, not technical layers. The boundaries you choose now will be the seams along which the monolith eventually splits — if you split it.
- When module A needs data from module B, A calls B's API. *No* SQL join across module boundaries; *no* direct table access. See [[information-hiding]].
- When you find yourself adding a circular dependency between modules, stop. Either A should depend on B or B on A, never both — split the shared concern into a third module C that both depend on.
- When you reach for a microservice, ask first whether a new module would solve the problem. A new module is hours; a new service is months. See [[premature-optimization]] applied to architecture.

## Anti-patterns

- A "modular monolith" whose modules import each other's internals at will. The architecture is decorative.
- Database tables owned by everyone and no one. Schema changes break unrelated modules; ownership confusion guarantees coupling.
- A monolith with no internal seams, called a monolith *as an excuse* for not drawing any. The "ball of mud" — see [[refactoring]].
- A "modular" codebase whose only enforcement of boundaries is README files. Boundaries by convention are crossed under pressure.
- A monolith and a few services *next to each other*, neither cleanly monolithic nor cleanly microservice. The distributed monolith — see [[microservices]] anti-patterns.
- Pre-splitting modules into services because "we'll need to scale later." See [[premature-optimization]]; needs that did not arrive are debts paid without benefit.

## References

- Brown, S. (2015). "Modular Monoliths." simonbrown.je. (The original phrasing.)
- Larrey, K. (2020). *Shopify's Modular Monolith*. Shopify Engineering blog. (Production-scale case study; module boundaries via Sorbet and Packwerk.)
- Fowler, M. (2015). "MonolithFirst." martinfowler.com/bliki/MonolithFirst.html. (The "start with a monolith" argument from the microservices camp itself.)
- Newman, S. (2019). *Monolith to Microservices*. O'Reilly. (Counterpoint and migration patterns.)
- Tornhill, A. (2018). *Software Design X-Rays*. Pragmatic Bookshelf. (How to discover the right modular boundaries empirically from change history.)
