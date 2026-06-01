# Microservices

A style in which a system is built as a suite of **small, independently-deployable services** organized around business capabilities. Articulated by Martin Fowler and James Lewis in "Microservices" (martinfowler.com, 2014) and given a book-length treatment in Sam Newman's *Building Microservices* (2015; 2nd ed., 2021). The pattern's claim: smaller services, evolved by smaller teams, ship faster and stay manageable as the system grows.

The promise is real; the cost is high. Microservices trade the complexity of a large codebase for the complexity of a distributed system — and the second complexity is materially harder than the first. Choose this architecture only if the value of independent deploy and team autonomy exceeds the cost of operating a distributed system at scale. See [[distributed-systems-fallacies]].

## What "microservice" actually means

The word is fuzzy. Fowler & Lewis listed nine characteristics; the most load-bearing:

- **Componentization via services** — each capability is a separate deployable, communicating over the network.
- **Organized around business capabilities** — service boundaries follow domain boundaries, not technical layers. *No* "the database team", *no* "the API team" — services slice vertically.
- **Decentralized data management** — each service owns its data. No shared database. (The single most violated rule of microservices, and the one whose violation guarantees most of the pain without the corresponding benefit.)
- **Smart endpoints and dumb pipes** — logic in services, not in the broker. A message bus that routes; not an ESB that decides.
- **Design for failure** — every call across the network can fail. Timeouts, retries, circuit breakers, fallbacks at every seam. See [[distributed-systems-fallacies]] and [[error-handling]].

A service that ships with a "small" codebase but reads the same Postgres table as five other "small" codebases is a distributed monolith — the worst of both worlds.

## What it asks of you

- When you draw service boundaries, draw them **along business capabilities** — customers, orders, billing — not along technical layers. The most predictive question: *can a single team own this end-to-end, from database to API?* If yes, the boundary works.
- When two services need to read each other's data, ask whether they are actually one service. Frequent cross-service reads are a boundary smell.
- When you publish an API consumed by other services, treat its shape as a contract evolving under [[hyrums-law]]. Versioning, compatibility tests, deprecation discipline — every microservice owes its callers a stable surface.
- When you provision a new service, the *operational* cost arrives immediately: a deployment pipeline, a runbook, monitoring, alerts, on-call. Be willing to pay it. If you are not, you don't have a service — you have an outage waiting. See [[observability]].
- When you write across services, plan for **eventual consistency**. Two-phase commit across services is not coming back. Use sagas, idempotent operations, compensating actions. See [[idempotency]].

## IRON LAW

**No shared database.** Two services that read or write each other's tables are not two services; they are one service with two deployment artifacts. The single biggest predictor of microservice-architecture failure is shared state in a shared schema. Each service owns its data; cross-service data exchange goes through APIs or events.

## GOLDEN RULES

- **Aim for service boundaries that match team boundaries.** Conway's law is descriptive — the architecture *will* mirror the org chart eventually. Pick boundaries the org can actually staff and own.
- **Aim for synchronous calls minimized; asynchronous events preferred** at service boundaries. Synchronous fan-out cascades tail latency; events isolate. See [[event-driven]].
- **Aim for one deployment pipeline per service, fully owned by the service team.** Shared pipelines reintroduce the coupling microservices were meant to eliminate.
- **Aim for observability built in from day one.** Distributed tracing, structured logging with correlation IDs, SLOs per service. Without these, multi-service incidents are unsolvable. See [[observability]].
- **Aim to start with a monolith.** Most successful microservice architectures came from a monolith that needed to scale. Premature microservices are how teams ship one feature in a quarter that a monolith would have shipped in a week. See [[premature-optimization]].

## Anti-patterns

- A "microservice" that shares a database with three other services. The architecture is microservice in shape, monolith in coupling.
- A new feature that requires changes in five services. The service boundaries are wrong; the change should have been one.
- Synchronous chains five services deep. Tail latency dominates; a slow downstream takes the whole request down. See [[distributed-systems-fallacies]].
- A "shared library" that every service imports, evolving without version discipline. Microservices are now coupled at the source-code level.
- A service per noun — `UserService`, `OrderService`, `OrderItemService`, `OrderItemTagService`. Boundaries should be **capabilities**, not entities.
- An organization with five engineers running thirty microservices. The operational overhead exceeds the headcount.

## References

- Lewis, J., & Fowler, M. (2014). "Microservices." martinfowler.com/articles/microservices.html.
- Newman, S. (2015, 2021). *Building Microservices*, 1st and 2nd eds. O'Reilly.
- Newman, S. (2019). *Monolith to Microservices: Evolutionary Patterns to Transform Your Monolith*. O'Reilly.
- Richardson, C. (2018). *Microservices Patterns: With Examples in Java*. Manning. (The pattern catalog — sagas, API composition, transactional outbox.)
- Vernon, V. (2016). *Domain-Driven Design Distilled*. Addison-Wesley. (Bounded contexts as the unit of decomposition.)
- Skelton, M., & Pais, M. (2019). *Team Topologies*. IT Revolution Press. (The org-chart side of the equation.)
