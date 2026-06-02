# MVC idioms

Concern-specific patterns and rules for working in an MVC codebase. Seeded when the project selects `--architecture mvc`. The conceptual overview lives at [`corpus/principles/mvc.md`](../../principles/mvc.md); this directory captures the per-concern detail (models, views, controllers, routing, helpers).

## Files

- [`models.md`](models.md) — what belongs on a model; the anemic-model trap.
- [`controllers.md`](controllers.md) — controllers as translators; size limits.
- [`views.md`](views.md) — views are dumb; presentation vs. business state.

Paired rules live under [`guides/idioms/mvc/`](../../../guides/idioms/mvc/).

## Adding more

If the team's stack adds a layer (helpers, decorators, presenters, view-models), drop a file here and a paired rule file under `guides/idioms/mvc/`.
