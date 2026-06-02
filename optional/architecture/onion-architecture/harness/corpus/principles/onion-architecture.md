# Onion Architecture

Articulated by Jeffrey Palermo in a four-part series on jeffreypalermo.com (2008). A direct descendant of [[hexagonal]] and a sibling of [[clean-architecture]] — all three share the same load-bearing rule (dependencies point inward) but differ in how they name and slice the inside. The onion's distinctive emphasis: **the domain model is the absolute center**, and *application services* sit between the model and the outside world as a separate ring.

> **Rules extracted:** [`guides/principles/onion-architecture.md`](../../guides/principles/onion-architecture.md). This file holds the full reasoning, anti-patterns, and references.

## The rings, outermost to innermost

1. **Infrastructure** — UI, persistence, external services, framework. The interchangeable shell.
2. **Application services** — orchestration, transaction boundaries, use-case coordinators. Knows the domain; does not implement domain rules.
3. **Domain services** — operations that span multiple entities or that don't naturally belong to a single one. Pure domain logic, no I/O.
4. **Domain model** — entities, value objects, aggregates, domain events. The heart of the system.

The arrows in Palermo's drawing all point inward. The infrastructure ring depends on application services; application services depend on the domain. The domain depends on nothing.

## Onion vs. Hexagonal vs. Clean

The three are often used interchangeably and the distinctions can feel academic. The most useful framing:

- **Hexagonal** ([[hexagonal]]) emphasizes *symmetry* — every adapter is equal; ports define the seam.
- **Clean** ([[clean-architecture]]) emphasizes *layering* — four named rings with use cases as a distinct concept.
- **Onion** emphasizes the *primacy of the domain model* — the domain is the center; everything else is service to it. Application services are explicitly between domain and infrastructure.

If your team thinks in domain-model-first terms (DDD-leaning, with rich aggregates and domain services), the onion drawing tends to fit best. If your team thinks in use-case-first terms, Clean's four rings tend to fit better. If your team thinks in driving/driven terms, Hexagonal's polygon tends to fit best. They prescribe the same dependency rule; they differ in which concept the drawing emphasizes.

## What it asks of you

- When you place a piece of business logic, ask whether it belongs to a single entity (put it on the entity), to multiple entities (put it in a domain service), or to a use case that orchestrates entities to produce a result (put it in an application service). See [[separation-of-concerns]].
- When you write a domain entity, it should be testable with no application service present, no infrastructure present, no framework. Pure types and pure logic.
- When the application service needs to persist or notify, declare an interface owned by that ring. Implement it in infrastructure.
- When you find yourself reaching from the domain model into infrastructure (an `@Entity` annotation, a logger import, a `DateTime.now()` call), the dependency direction has been violated. Push the dependency outward.

## Anti-patterns

- A domain entity that knows how to save itself (Active Record). Persistence has reached into the center.
- An anemic domain model — entities with only data and getters/setters; all logic in application services. The domain ring is empty; the architecture is layered, not onion. See [[object-oriented-design]] (tell-don't-ask).
- A "domain service" that is really an application service in disguise — it orchestrates use cases and calls repositories. Move it outward.
- Skipping the application-services ring; putting orchestration in controllers. Use cases now live in the framework layer.
- Tooling tags on the domain model (`@JsonProperty`, `@Column`). Each tag is a tiny reach from the outside in. See [[hyrums-law]] — the observable shape of the domain is now coupled to the tool.

## References

- Palermo, J. (2008). *The Onion Architecture: part 1–4*. jeffreypalermo.com.
- Evans, E. (2003). *Domain-Driven Design*. Addison-Wesley. (The domain-centric thinking the onion makes structural.)
- Vernon, V. (2013). *Implementing Domain-Driven Design*. Addison-Wesley.
- Cockburn, A. (2005). "Hexagonal Architecture." (Sibling — see [[hexagonal]].)
- Martin, R. C. (2017). *Clean Architecture*. Prentice Hall. (Sibling — see [[clean-architecture]].)
