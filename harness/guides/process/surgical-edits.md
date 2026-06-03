# Surgical Edits

Touch only what serves the task. Counterpart to [[scoping]] (size budget on output) — this is the *scope* discipline: not "how much," but "*what kind* of changes are allowed."

## GOLDEN RULES

- **Aim to make every changed line trace to the user's request.** If a line cannot be justified by the task, it should not be in the diff.
- **Aim to remove only the imports / variables / functions that *your* changes orphaned.** Pre-existing dead code is not in scope. Note it; do not delete it.
- **Aim to match existing style** even when the agent would prefer a different one. Style changes are their own work — they belong in their own commit.
- **Aim to keep formatting changes out of behavior commits.** A behavior commit that re-formats five files hides the actual change inside the formatting.

## RULES

- **No "while I'm here" cleanups.** Renaming a nearby variable, fixing a typo in an unrelated comment, tightening unrelated code — all separate work.
- **No "improving" adjacent code.** If the agent notices something — surface it, do not silently fix it.
- **No mass formatter runs as part of a logic change.** Run the formatter, but in its own commit, before or after.
- **No moving files as a side effect.** A move is a structural change; it gets its own commit. See [[scoping]].
- **No "since I'm refactoring this anyway, let me also..." sequences.** Each "also" is a new commit's worth of work.

## Why this is agent-specific

A coding agent reads more context than a human reviewer does. It sees the typo three files away, the inconsistency in the function naming, the missing docstring. Its natural pull is to fix everything it sees — and the result is a 600-line PR for a 20-line task, with the real change buried in the noise.

Human reviewers cannot review 600 lines of mixed work with the same care they bring to 20 lines of focused work. The user pays for the cleanup with reviewer fatigue and a higher chance of a regression slipping through.

The remedy is to draw a hard boundary: *the diff serves the request*. If something else needs to be done, it gets its own request.

## Sensors

- The **drift sensor** flags changes outside the touched region as candidates for review.
- The **scoping rule** (see [[scoping]]) catches the size symptom; this guide catches the cause.

## Anti-patterns

- Reformatting the entire file the agent edited.
- Renaming variables for "clarity" in code the agent did not otherwise change.
- "Updating the comment to match modern conventions" mid-feature.
- Sorting imports / reordering keys / fixing whitespace alongside logic changes.
- A PR description that says "also cleaned up X, Y, Z."

---

See also: [[scoping]], `release.md`, [[refactoring]].
