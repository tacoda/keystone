# Controllers

A controller is a **translator** — it takes input from outside the system, asks the model to do something, and picks a view to render. If a controller is computing business rules, those rules have leaked from the model.

> **Rules extracted:** [`guides/idioms/mvc/controllers.md`](../../../guides/idioms/mvc/controllers.md).

## What belongs in a controller

- Parsing input (params, headers, multipart bodies).
- Authentication and authorization filters (delegated where possible).
- Calling into the model or a domain service.
- Picking a view template and providing the data it needs.
- HTTP-shaped responses (status codes, redirects, content types).

## What does NOT belong

- Business rules. ("Can this customer place an order?" — that's a model question.)
- Multi-step domain orchestration. (Wrap it in a domain service the controller calls.)
- Direct ORM queries scattered across actions. (Encapsulate in repository methods on the model.)
- Conditional rendering logic that depends on business state. (Compute a flag once; pass it to the view.)

## Sizing

A healthy controller has **under ~10 actions** — usually the standard REST set (index, show, new, create, edit, update, destroy) plus one or two custom actions. When a controller has 50 methods, it is hiding several smaller controllers.

## How to apply

- When you write a controller method, ask: "If I removed the framework, would this code make sense?" If the answer is "no, because it's all framework wiring," good — that's controller work. If it's "yes, this is a domain operation," push it to the model.
- When two controllers call into the same multi-step logic, extract a domain service and call it from both.
- When an action grows beyond ~15 lines, look for a method on the model that wants to exist.

## Review checklist

- [ ] Controllers do translation, not business logic.
- [ ] Action methods are short (rough target: under 15 lines).
- [ ] No SQL or raw query construction in the controller — that's the model's job.
- [ ] The controller picks a view; the view does not pick a model.

**Traces to:** [`corpus/principles/mvc.md`](../../principles/mvc.md), [`corpus/principles/separation-of-concerns.md`](../../principles/separation-of-concerns.md).
