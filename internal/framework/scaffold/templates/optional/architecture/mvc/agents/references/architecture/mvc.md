# Model-View-Controller (MVC)

The oldest GUI architecture pattern still in use. Introduced by Trygve Reenskaug at Xerox PARC in 1978 ("Models-Views-Controllers," internal note) and refined for Smalltalk-80, where it became the canonical pattern for building interactive applications. Forty-plus years later, MVC's vocabulary is everywhere — Rails, Django, ASP.NET, Spring MVC, iOS — though every framework that claims MVC implements it slightly differently.

> **Rules extracted:** [`guides/mvc.md`](../guides/mvc.md). This file holds the full reasoning, anti-patterns, and references.

## The three roles

- **Model** — the application's data and the rules that govern it. Knows nothing about how it is displayed or how the user interacts with it. The unit of correctness.
- **View** — the presentation of the model to the user. Renders state; receives input gestures and forwards them to the controller.
- **Controller** — interprets user input. Translates gestures into operations on the model; may select a new view.

The original Smalltalk MVC was triadic: model, view, and controller were peer objects, communicating through events. The model notified views of changes; the controller interpreted input independent of the view. Modern web MVC (Rails, Spring) is closer to **Model 2 / Web MVC**: the controller is invoked by a router, calls into the model, and selects a view template to render. The peer-to-peer dynamics of the original have largely disappeared on the server side.

## What "Model" actually is

The single most common point of confusion in MVC discussions: *the Model is not just an ORM row.* In Reenskaug's original, the model was the **domain object** — with state, behavior, and validity rules. In framework MVC, "the model" often degenerates into "the row" — a thin object with getters and setters, with logic spread across controllers and helpers. This is the **anemic model** anti-pattern; see [[object-oriented-design]].

A healthy MVC has rich models. The controller is small. The view is dumb. If the controller is large, business logic has leaked upward; if the view contains conditionals on the model's internal shape, presentation has reached into the domain.

## What it asks of you

- When you write a controller, the controller's job is **translation**: parse input, ask the model to do something, pick a view to render. If the controller is computing business rules, those rules belong on the model. See [[separation-of-concerns]] and [[object-oriented-design]] (tell-don't-ask).
- When you write a view, the view's job is **rendering**. No conditionals on business state ("if the user is an admin and the account is overdue and …"). If the view has business conditionals, give the controller (or a view-model) a precomputed flag.
- When you write a model, the model owns its invariants. Validation lives on the model, not on the controller. See [[design-by-contract]].
- When you find logic duplicated across controllers, the abstraction it should be is usually a method on the model or a domain service. See [[pragmatic-principles]] (DRY).

## Anti-patterns

- Business logic in controllers. The classic "fat controller" smell; the controller has absorbed orchestration *and* domain rules.
- Conditional rendering in views based on raw model fields. The view now knows business state shape.
- "Helpers" full of business logic — a pseudo-layer that exists because the controller was too big and the model was too small.
- A model that is an ORM entity passed unchanged to the view. The view is now coupled to the database schema. See [[information-hiding]].
- Cross-controller calls that bypass the model. Two controllers coordinating without going through the model is the model-is-anemic confession.
- Frameworks that call it MVC but route directly from URL to view-template; the controller has been elided. Not strictly an anti-pattern, but the vocabulary is misleading.

## References

- Reenskaug, T. (1979). "Models–Views–Controllers." Xerox PARC, internal note. (The original.)
- Krasner, G. E., & Pope, S. T. (1988). "A Cookbook for Using the Model-View-Controller User Interface Paradigm in Smalltalk-80." *Journal of Object-Oriented Programming*, 1(3).
- Fowler, M. (2003). *Patterns of Enterprise Application Architecture*. Addison-Wesley. ("Model View Controller," p. 330 — the canonical enterprise variant.)
- Reenskaug, T. (2003). *The Model-View-Controller (MVC) — Its Past and Present*. (Reenskaug's own retrospective, distinguishing the original from later framework MVCs.)
