# Wishlist

Known gaps the team plans to address — candidate rules and guides whose treatment for *this* project has not yet been authored. **Not a TODO list.** Items here become real **inbox** candidates when a real situation triggers them.

The Learning flywheel works best when the agent writes inbox items from *lived experience*. The wishlist is the separate channel for things the team agrees are worth covering eventually, surfaced so the agent can recognize the situation and produce a richer inbox entry when it occurs.

## Active wishes

- **Performance regression rules** — when a change crosses a performance budget, fail it. Establish budgets per region; capture baseline + delta.
- **Accessibility** — keyboard navigation, ARIA, alt text, contrast ratios, focus management. Lives in `guides/idioms/<frontend-stack>/` once bootstrap detects one.
- **Internationalization (i18n)** — never hardcode user-facing strings; locale-safe formatting for dates, numbers, currencies; right-to-left layouts.
- **Test data / fixtures** — when to use builders, factories, snapshots, golden files. Seeded faker patterns per stack.
- **External-service contracts** — fake adapters for external HTTP services; never hit prod APIs from tests; contract tests at boundaries.
- **Time, date, timezone discipline** — beyond what [[determinism]] covers; storing UTC, displaying local, monotonic vs wall clock.

## Promoting from the wishlist

When the team is ready to add one:

1. Move the bullet into a real `learning/inbox/<timestamp>-<slug>.md` candidate. Drop a one-line note linking back to this file.
2. Run **synthesize** to promote it into a guide and/or corpus file.
3. Remove the bullet from this file.

## Adding to the wishlist

Anyone can append a bullet. Keep each one to a sentence; the depth lives in the eventual corpus/guide file. If a wish stays here for a year without ever becoming an inbox item, it probably is not actually a gap — consider removing it.

## Not in scope

- One-off ideas for a specific PR (those go in the PR).
- Stack-specific patterns that bootstrap will add automatically.
- Things already covered by an existing guide.
