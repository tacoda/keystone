---
kind: corpus
id: corpus/state/code-debt
description: 'Paired ledger: [harness-debt.'
---
# Code Debt Ledger

> **Template.** The **audit** action and **debt-review** action will populate this from the [code-debt sensor](sensors/code-debt.md). Until then, leave as-is or fill in by hand.

Paired ledger: [`harness-debt.md`](harness-debt.md) — debt in the harness itself. Tracked separately on purpose.

The known debt this codebase carries. One row per item. The point of the ledger is to make the cost of debt *visible during planning* — when **orient** runs against a region with load-bearing debt, the plan should account for it.

## Ledger

| ID | Location | Category | Severity | Owner | Trigger to revisit | Notes |
|---|---|---|---|---|---|---|
| `<DEBT-001>` | `<path:line or region>` | `<deliberate|drift|shortcut|discovery>` | `<load-bearing|noisy|stale>` | `<name or team>` | `<event or date>` | `<one line>` |

## Categories

See [`harness/sensors/code-debt.md`](sensors/code-debt.md) for category and severity definitions. Keep them consistent — the planning phase reads this table verbatim.

## How to use it

- **Before planning a change in a region with load-bearing debt** → factor the debt into the plan. Either pay it down first, route around it, or document why you're adding to it.
- **Before adding new debt** → add the row first (with a trigger to revisit). A debt item without a revisit trigger rots into noise.
- **During audit** → sweep `stale` items for removal and `discovery` items for triage.

## Pruning

`stale` items are deleted during **debt-review**, not archived. The whole point of the ledger is signal density; archiving every fix turns it into a museum.
