# Hexagonal Architecture

Also called **Ports and Adapters.** Articulated by Alistair Cockburn in "Hexagonal Architecture" (alistair.cockburn.us, 2005). The central idea: the **domain core** has no dependencies on the outside world — not on the database, not on the web framework, not on the message bus. Everything outside is reached through **ports** (interfaces owned by the domain) and supplied by **adapters** (implementations owned by the infrastructure).

The hexagon shape in Cockburn's drawings is incidental — what matters is that the inside is a regular polygon: every direction is equal, and the *number* of sides matches the number of ways the application is driven or driven. The application does not know whether it is being driven by HTTP, gRPC, a CLI, a test, or a queue. The application does not know whether persistence is Postgres, an in-memory store, or a fake.

> **Rules extracted:** [`guides/principles/hexagonal.md`](../../guides/principles/hexagonal.md). This file holds the full reasoning, anti-patterns, and references.

## The structure

- **Inside the hexagon** — the domain. Plain language types, business rules, no framework imports. The unit of correctness.
- **Ports** — interfaces the domain declares. *Driving ports* describe what the domain can be asked to do (use cases). *Driven ports* describe what the domain needs from the outside (repositories, clocks, message senders).
- **Adapters** — implementations of those interfaces. *Driving adapters* translate from outside (HTTP, CLI, test) into calls to driving ports. *Driven adapters* translate from driven ports into outside (Postgres, Redis, SMTP).

The compile-time dependency direction is **always inward.** Adapters depend on ports; ports are owned by the domain; the domain depends on nothing.

## What it asks of you

- When you import a framework or driver from inside the domain, you have broken the architecture. The domain only imports its own language's standard library and pure-domain modules. See [[separation-of-concerns]] and [[information-hiding]].
- When you write a port, name it from the *domain's* point of view, not the adapter's. `UserRepository.findByEmail` (domain language), not `PostgresUserDAO.selectByEmailColumn` (adapter language).
- When you test domain logic, do not stand up the database. Inject a fake adapter. The hexagon makes this the cheapest kind of test, not a special one. See [[testing-patterns]] (mock at real boundaries — the port is the boundary).
- When you find yourself adding a domain method that takes an `HttpRequest`, the driving adapter is leaking. The domain method should take parameters in domain terms; the adapter is what translates.

## Anti-patterns

- An ORM entity exposed all the way out to an HTTP handler. The adapter's persistence model is leaking through every layer.
- A "service" class that takes an `HttpRequest` directly. The driving adapter has been pushed into the domain.
- A domain method that calls a vendor SDK directly ("the domain needs to send an email" — yes, through a port).
- A test that requires a running Postgres to exercise a calculation. The calculation is in the wrong place.
- Adapters that talk to each other directly without going through the domain. The hexagon has been short-circuited.

## References

- Cockburn, A. (2005). "Hexagonal Architecture." alistair.cockburn.us/hexagonal-architecture/.
- Vernon, V. (2013). *Implementing Domain-Driven Design*. Addison-Wesley. (Chapter 4 — DDD layering via hexagonal.)
- Evans, E. (2003). *Domain-Driven Design*. Addison-Wesley. (The domain core whose isolation hexagonal preserves.)
- Cockburn, A. (n.d.). *Hexagonal Architecture Explained*. (Cockburn's own restatement; he stresses that the hexagon is *symmetric* — both sides equally important — which the Clean Architecture drawings sometimes obscure.)
