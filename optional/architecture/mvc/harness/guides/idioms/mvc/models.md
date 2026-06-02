# Models — rules

The rules from [`corpus/idioms/mvc/models.md`](../../../corpus/idioms/mvc/models.md).
Loaded ambient; enforced by the [drift sensor](../../../sensors/drift.md).

## IRON LAW

**The model is not a row.** A class with only fields and accessors is not a model — it is a persistence stub. Business logic that belongs on the model must live on the model. If logic lives in a controller, helper, or "service" only because the model is anemic, that is a violation.

## GOLDEN RULES

- **Aim for behavior on the model.** A method like `Customer#suspend!` beats a controller doing the same steps.
- **Aim for validation on the model.** Controllers may surface errors, but the model decides what is valid.
- **Aim for domain language in model methods.** `Order#cancel!` not `Order#set_status(7)`.
- **Aim for no HTTP / rendering / framework concepts in model methods.** If the model imports a request object, something is wrong.

---

Traces to: [`corpus/idioms/mvc/models.md`](../../../corpus/idioms/mvc/models.md).
