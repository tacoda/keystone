---
kind: corpus
id: corpus/state/charter-debt
description: 'Paired ledger: [code-debt.'
---
# Charter Debt Ledger

> **Template.** The **audit** action's Pruning flywheel populates this from the [charter-debt sensor](sensors/charter-debt.md). Until then, leave as-is or fill in by hand.

Paired ledger: [`code-debt.md`](code-debt.md) — debt in the codebase. Tracked separately on purpose.

The debt the charter itself is carrying. When the charter gives stale or wrong guidance, that cost lands on every task the agent runs against this project. This ledger makes that cost visible so the team can pay it down deliberately.

## Ledger

| ID | Location | Category | Severity | Notes |
|---|---|---|---|---|
| `<HDEBT-001>` | `<charter/path>` | `<stale-rule\|dead-idiom\|placeholder\|failing-sensor\|empty-shell\|uncited-policy\|unresolved-gap\|drifted-state>` | `<load-bearing\|noisy\|stale>` | `<one line>` |

## Categories

See [`charter/sensors/charter-debt.md`](sensors/charter-debt.md) for category and severity definitions.

## How to use it

- **Before any task** → if the touched region has a `load-bearing` charter-debt item, address it first (the agent will be operating on bad guidance otherwise).
- **During audit** → triage `placeholder` and `failing-sensor` items as `load-bearing` by default; they're the most likely to mislead.
- **During synthesize** → before promoting an inbox item into a guide, check that no `stale-rule` already covers the topic. Adding a new rule next to a stale one compounds the noise.

## Pruning

`stale` items are deleted during **audit**, not archived. The charter lives in git — history is already preserved.
