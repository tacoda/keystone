# Controllers — rules

The rules from [`corpus/idioms/mvc/controllers.md`](../../../corpus/idioms/mvc/controllers.md).
Loaded ambient; enforced by the [drift sensor](../../../sensors/drift.md).

## IRON LAW

**Controllers translate, they do not decide.** A controller may parse input, call the model, and choose a view. A controller may not contain business rules. Any branching on domain state that is not "did the model succeed or fail" is a violation.

## GOLDEN RULES

- **Aim for controllers under ~10 actions.** Larger means a missing decomposition.
- **Aim for action methods under ~15 lines.** Beyond that, a model method or domain service usually wants to exist.
- **Aim for no SQL or ORM queries in controller methods.** Repository methods on the model take that load.
- **Aim for one model orchestration per action.** Multi-step domain logic across an action body is a service waiting to be extracted.

---

Traces to: [`corpus/idioms/mvc/controllers.md`](../../../corpus/idioms/mvc/controllers.md).
