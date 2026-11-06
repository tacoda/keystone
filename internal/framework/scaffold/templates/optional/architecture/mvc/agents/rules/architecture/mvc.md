# Model-View-Controller (MVC) — rules

The rules from [`corpus/mvc.md`](../corpus/mvc.md).
Loaded ambient; enforced by the [drift sensor](../../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**The model is not a row.** The model is the place where domain behavior lives. A codebase whose "models" only have fields and accessors has implemented persistence-with-controllers, not MVC. The pattern's value comes from the model being substantial.

## GOLDEN RULE

- **Aim for fat models, skinny controllers, dumb views.** The phrase is decades old and still right. Most "MVC done badly" inverts the proportions.
- **Aim for controllers under ~10 actions.** A controller with 50 methods is hiding several smaller controllers.
- **Aim for views that render without knowing the model's internal shape.** A view-model or presenter — a stable struct shaped for the view — is often worth the extra type.
- **Aim for controllers that test as integration tests of routing and translation, not as integration tests of business logic.** The business-logic tests are unit tests on the model.

---

Traces to: [`corpus/mvc.md`](../corpus/mvc.md).
