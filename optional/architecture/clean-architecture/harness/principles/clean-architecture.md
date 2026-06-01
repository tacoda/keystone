# Clean Architecture

Robert C. Martin's synthesis of Hexagonal (Cockburn), Onion (Palermo), DCI, and BCE into a single layered scheme. Articulated in "The Clean Architecture" (cleancoder.com, 2012) and *Clean Architecture* (Prentice Hall, 2017). The recurring drawing — concentric circles with arrows pointing inward — names four layers; the **Dependency Rule** is the load-bearing principle.

## The four layers, outermost to innermost

1. **Frameworks & drivers** — the web framework, the database, the UI toolkit, the message broker. The replaceable substrate.
2. **Interface adapters** — controllers, presenters, gateways. Translate between the substrate and the application's vocabulary.
3. **Application business rules / use cases** — what the system *does*. Orchestrates entities to satisfy a request. Owns no domain rules, no I/O.
4. **Enterprise business rules / entities** — the *what* of the business, independent of any application. The innermost ring.

## The Dependency Rule

> Source code dependencies must point *only inward*, toward higher-level policies. — Martin

Nothing in an inner ring may name anything in an outer ring. Use cases do not import framework types; entities do not import use-case types. The dependency direction is enforced by interfaces: when an inner ring needs something the outer ring provides (a database write, an email send), it declares an interface — and the outer ring implements it. This is [[SOLID]]'s Dependency Inversion as the structural rule of the codebase.

Clean Architecture and [[hexagonal]] agree on the dependency direction. Where Hexagonal draws a symmetric polygon (every adapter is equal), Clean draws concentric rings (some layers are *more central* than others). The two views are compatible; pick the one whose drawing you find clearer.

## What it asks of you

- When you write a use case, take inputs as plain data structures and return plain data structures. Never accept an `HttpRequest`; never return an ORM entity. See [[separation-of-concerns]] and [[information-hiding]].
- When the application needs to reach out (database, queue, third-party API), the use case declares an interface in its own ring. The outer ring writes the implementation. The use case never imports the outer type.
- When you find yourself wanting to "just use the framework's request object" inside a use case, that is the framework leaking inward. Convert at the controller. See [[hexagonal]] — same rule, different geometry.
- When you wire the application at startup, the composition root lives at the outermost ring (the framework layer). It is the only place that knows about every layer.

## IRON LAW

**No imports go outward.** A grep of every inner-ring file for an outer-ring identifier must return zero results. The rule is not "minimize" — it is **zero**. If you cannot enforce it with directory structure, package visibility, or a linter, the architecture is decorative.

## GOLDEN RULES

- **Aim for inner rings that compile and test without any outer ring.** The entities and use cases should build into a library with no framework or database dependencies.
- **Aim for interfaces owned by the consumer.** The use case owns the `UserRepository` interface; the database adapter implements it. The interface lives in the inner ring.
- **Aim for the outermost ring to be as thin as possible.** Frameworks change; controllers should not contain business logic that has to change with them.
- **Aim for crossing the boundary with data structures, not framework types.** The boundary is where the dialect changes.

## Anti-patterns

- A controller that contains business rules ("if the user is admin, …") — the rule belongs in the use case, the controller dispatches.
- Use cases that depend on the ORM. The ORM is in the outer ring; the use case has reached out.
- "Generic" repositories with framework-specific query types in their signatures. The framework has leaked through the interface.
- Entities annotated with framework decorators (`@Entity`, `@Column`, `@Table`). The entity now requires the framework to exist.
- A single Go/Java package containing everything from controller to entity. The dependency direction is unenforced; the rings are imaginary.

## References

- Martin, R. C. (2017). *Clean Architecture: A Craftsman's Guide to Software Structure and Design*. Prentice Hall.
- Martin, R. C. (2012). "The Clean Architecture." blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html.
- Cockburn, A. (2005). "Hexagonal Architecture." (Precursor; see [[hexagonal]].)
- Palermo, J. (2008). "The Onion Architecture." (Sibling pattern; see [[onion-architecture]].)
- Jacobson, I. (1992). *Object-Oriented Software Engineering: A Use Case Driven Approach*. Addison-Wesley. (Boundary / Control / Entity — the BCE precursor to Clean's use-case layer.)
