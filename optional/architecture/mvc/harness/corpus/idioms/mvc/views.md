# Views

A view's job is **rendering**. Show the user what the model says. Receive the user's gestures and forward them to the controller. Views should be the dumbest layer in the stack.

> **Rules extracted:** [`guides/idioms/mvc/views.md`](../../../guides/idioms/mvc/views.md).

## What belongs in a view

- Template markup (HTML, JSX, ERB, Blade, whatever the stack uses).
- Loops over collections that the controller prepared.
- Display formatting (date format, currency, plural inflection) — usually via helpers or a view-model.
- Localization tokens.

## What does NOT belong

- Business conditionals. ("If the user is an admin AND the account is overdue AND…") — compute a flag in the controller (or on a view-model) and pass it in. The view reads the flag.
- Direct access to the database or ORM. The view does not query.
- Multi-step state machines. The view shows state; it does not advance state.
- Logic that depends on the internal shape of a model. If the view checks `order.status_code == 7`, the view has reached into the domain.

## View-models and presenters

When a view needs computed display values (formatted, sorted, decorated), introduce a **view-model** or **presenter** — a stable struct shaped for the view's needs. The model stays domain-focused; the view consumes presenter properties; the controller wires the presenter.

A view-model is worth the extra type when:

- The same view appears across multiple controllers and needs consistent shape.
- The model contains fields the view shouldn't see (sensitive data, internal IDs).
- The view requires aggregations or derived values that don't belong on the model.

## How to apply

- When a view contains an `if` over business state, treat it as a smell. Push the condition upstream.
- When the same display computation appears in two templates, hoist it to a helper or a presenter.
- When a view's tests are hard to write because the setup involves "build a real model," that's the view-model wanting to exist.

## Review checklist

- [ ] Views render; they do not compute business rules.
- [ ] No DB / ORM access in the view layer.
- [ ] Conditional rendering reads a precomputed flag, not raw model state.
- [ ] Views work against a stable presenter or view-model, not an ORM entity.

**Traces to:** [`corpus/principles/mvc.md`](../../principles/mvc.md), [`corpus/principles/information-hiding.md`](../../principles/information-hiding.md).
