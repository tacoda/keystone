# Layered Architecture

The default architecture of most enterprise applications written between 1995 and today: code is organized into horizontal layers, each calling only the layer beneath it. Articulated formally in Martin Fowler's *Patterns of Enterprise Application Architecture* (2003) and earlier by Buschmann et al. in *Pattern-Oriented Software Architecture* (1996). Older than [[hexagonal]] and [[clean-architecture]]; the **parent pattern** they reformed.

## The canonical layers

- **Presentation** — UI, HTTP handlers, view templates, CLI surface. Translates external input/output to/from the layer below.
- **Application** — orchestrates use cases, manages transactions, coordinates domain operations.
- **Domain** — business rules, entities, domain services.
- **Persistence** — repository implementations, ORM mappings, database access.

Layers above call layers below. **Skipping** a layer ("presentation reaches into persistence") is the canonical violation.

## How it differs from Hexagonal / Clean / Onion

The structural difference is small but consequential: in a strict layered architecture, the **dependency direction is purely downward** — the application layer depends on the domain layer, *and the domain layer depends on the persistence layer below it*. In [[hexagonal]], [[clean-architecture]], and [[onion-architecture]], persistence is *outside* the domain; the domain owns an interface that persistence implements.

The practical consequence: in classic layered architecture, the domain knows about the database (sometimes via repository interfaces in the persistence layer; sometimes directly). The domain is harder to test without infrastructure. The reformed architectures invert just that one edge.

Many modern "layered" codebases are actually a *hybrid* — layered overall, but with the persistence interface owned by the domain (i.e. the dependency-inversion of [[SOLID]] applied at one critical seam). This hybrid is fine and common; the principles below describe both the strict and the hybrid form.

## What it asks of you

- When you cross a layer, cross it in the canonical direction (downward). A presentation handler calling a repository directly is the most common layered violation. See [[separation-of-concerns]].
- When a feature spans layers, name the operation at the *application* layer; the application layer is the only layer authorized to know multiple other layers. Presentation talks to application; application talks to domain and persistence.
- When you find domain logic in a controller, hoist it into the application or domain layer. The controller is for translating, not deciding. See [[object-oriented-design]] (tell-don't-ask).
- When you find SQL or ORM imports in the domain layer, decide: keep it strict-layered (acceptable, with caveats) or invert just that seam (the hybrid). Either is correct; mixing both within the same codebase is not.

## IRON LAW

**Calls go downward through layers, never upward and never skipping.** A presentation handler that invokes a repository directly has bypassed the application layer; an application service that emits an HTTP response has reached upward. Either is a violation. The whole point of layering is that each layer knows only the one beneath it.

## GOLDEN RULES

- **Aim for one layer per concern, with clear names.** Presentation, application, domain, persistence. Resist new "helper" layers that don't earn their place.
- **Aim for the application layer to own transactions.** Transactions span multiple repository calls; the application layer is the only place that knows the unit of work.
- **Aim for DTOs at layer boundaries.** Pass plain data structures across layers, not framework or ORM types. The boundary is where the dialect changes. See [[information-hiding]].
- **Aim to invert the persistence seam if the domain needs to be testable in isolation.** Decide once, apply consistently.

## Anti-patterns

- A "service layer" that has become a god — every endpoint maps to one method; the method does presentation, validation, business logic, persistence, and emails in fifty lines.
- A controller that opens a database transaction. Transaction control belongs to the application layer.
- A repository that contains business logic ("`UserRepository.activateAndNotify()`") — persistence has absorbed orchestration.
- Skipping layers "for performance" without measurement. See [[premature-optimization]].
- A "domain layer" that is empty — all logic in the application layer, all data in the persistence layer. The architecture is two-tier wearing a four-tier costume. See [[object-oriented-design]] (anemic domain model).

## References

- Fowler, M. (2003). *Patterns of Enterprise Application Architecture*. Addison-Wesley. (The definitive enterprise-layering treatment.)
- Buschmann, F., Meunier, R., Rohnert, H., Sommerlad, P., & Stal, M. (1996). *Pattern-Oriented Software Architecture, Volume 1*. Wiley. (The pre-enterprise articulation of layering as a fundamental pattern.)
- Evans, E. (2003). *Domain-Driven Design*. Addison-Wesley. (The DDD layering — presentation, application, domain, infrastructure — is the most cited modern variant.)
- Martin, R. C. (2017). *Clean Architecture*. Prentice Hall. (The reformed cousin; see [[clean-architecture]].)
