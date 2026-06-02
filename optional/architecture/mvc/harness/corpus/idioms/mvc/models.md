# Models

In MVC, the model is the **domain object** — state, behavior, validity rules. Not a row. The model owns its invariants; the controller asks it to do things; the view shows what it says.

> **Rules extracted:** [`guides/idioms/mvc/models.md`](../../../guides/idioms/mvc/models.md).

## What belongs on the model

- The data the domain cares about (and the data only the domain cares about — not display formatting).
- The rules that govern when that data is valid.
- The operations that transition the data from one valid state to another.
- The events the model emits when state changes.

## What does NOT belong on the model

- HTTP-specific knowledge (params, headers, request/response shapes). That's controller territory.
- Rendering / formatting / locale concerns. That's view territory.
- Cross-cutting concerns that span aggregates (auth, audit). Those belong in domain services or controller filters.

## The anemic-model anti-pattern

A model that is mostly getters and setters with no behavior is **anemic**. The business logic has leaked into controllers ("fat controller") or helpers ("util ghetto"). The cure is to push logic back onto the model. Ask: *if I were reading this code with no framework context, where would I expect this rule to live?* That's almost always the model.

See [[corpus/principles/object-oriented-design]] (tell-don't-ask).

## How to apply

- When you write a query method on the model, ask whether it's really domain language (`Customer#in_good_standing?`) or display language (`Customer#display_status` — better to compute in a presenter).
- When a controller does multi-step orchestration over a model (validate, mutate, save, notify), consider a domain service or model method that does the whole thing atomically.
- When validation appears in both the controller and the model, the model's version is canonical; remove the controller copy.
- When the model has a method whose name contains a framework concept (`render`, `params`, `request`), that's a smell.

## Review checklist

- [ ] Models have behavior, not just fields.
- [ ] Domain rules live on the model, not in controllers or helpers.
- [ ] No HTTP / rendering / view concepts in model methods.
- [ ] Validation is on the model.

**Traces to:** [`corpus/principles/mvc.md`](../../principles/mvc.md), [`corpus/principles/object-oriented-design.md`](../../principles/object-oriented-design.md).
