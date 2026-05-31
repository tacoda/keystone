# Separation of Concerns

Different aspects of a system should be addressed in different parts of the system. Coined by Edsger W. Dijkstra in "On the role of scientific thought" (1974): one should focus attention upon some aspect, and pull it apart from the rest, because the human mind cannot deal with the system as a whole at once.

## The principle

A *concern* is anything that can be reasoned about in isolation: a feature, a piece of business logic, a cross-cutting capability (logging, authorization, persistence), a deployment surface. Two concerns mixed in the same module make both harder to think about, and both harder to change.

Common axes of separation:

- **Layered architectures** — presentation / application / domain / persistence. Each layer addresses one concern; dependencies point inward.
- **Hexagonal architecture (ports and adapters)** — the domain core is isolated from infrastructure concerns by interfaces (ports), with adapters supplying details. The domain has no idea whether it is serving HTTP, gRPC, or a CLI.
- **Cross-cutting concerns** — logging, authorization, tracing — handled by middleware, aspects, or decorators so the domain code is not infested with them.

## What it asks of you

- When a function does *both* a business rule and a database write, ask whether either could change for reasons the other does not share.
- When a controller computes business logic, suspect a missing domain layer.
- When the domain layer imports the database client, the dependency is pointing the wrong way.
- When you need a "presentation-aware" branch in domain code, the concern is leaking.

## GOLDEN RULES

- **Aim for boundaries that match how the system will change.** Concerns that always change together belong together; concerns that change for unrelated reasons belong apart.
- **Aim for dependencies that point toward stability.** Volatile concerns depend on stable ones, never the reverse.

## Anti-patterns

- A controller that knows about database transactions.
- A domain object that knows about HTTP.
- Business logic embedded in templates / views.
- A "service" that does presentation, persistence, and policy in the same method.
- Cross-cutting concerns hardcoded into every domain class instead of factored out.

## References

- Dijkstra, E. W. (1974). "On the role of scientific thought." Reprinted in *Selected Writings on Computing: A Personal Perspective* (1982). Springer.
- Cockburn, A. (2005). "Hexagonal Architecture." alistair.cockburn.us/hexagonal-architecture.
- Evans, E. (2003). *Domain-Driven Design: Tackling Complexity in the Heart of Software*. Addison-Wesley.
- Martin, R. C. (2017). *Clean Architecture*. Prentice Hall.
