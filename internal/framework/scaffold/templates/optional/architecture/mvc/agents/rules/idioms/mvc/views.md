# Views — rules

The rules from [`corpus/idioms/mvc/views.md`](../../../corpus/idioms/mvc/views.md).
Loaded ambient; enforced by the [drift sensor](../../../sensors/drift.md).

## IRON LAW

**Views render, they do not compute business rules.** A view may loop, format, and localize. A view may not contain conditionals on raw business state, query the database, or advance a state machine.

## GOLDEN RULE

- **Aim for views that read precomputed flags, not raw model state.** `if can_edit:` not `if user.role == 'admin' and post.author_id == user.id and not post.locked:`.
- **Aim for view-models or presenters when display state is complex.** A stable struct shaped for the view beats an ORM entity with everything attached.
- **Aim for no DB calls in the view layer.** N+1 queries hide behind innocuous-looking template iteration.
- **Aim for templates that pass type-check or template-lint without the runtime.** If the template only works "as long as the controller sets things up right," that's a fragile boundary.

---

Traces to: [`corpus/idioms/mvc/views.md`](../../../corpus/idioms/mvc/views.md).
