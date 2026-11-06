---
kind: corpus
id: corpus/state/code-debt
description: 'Paired ledger: [charter-debt.'
last_reviewed: 2026-06-18
---
# Code Debt Ledger

Populated by the **audit** and **debt-review** actions from the [code-debt sensor](../../sensors/code-debt.md). Hand-edits during those actions are OK.

Paired ledger: [`charter-debt.md`](charter-debt.md) — debt in the charter itself. Tracked separately on purpose.

The known debt this codebase carries. One row per item. The point of the ledger is to make the cost of debt *visible during planning* — when **orient** runs against a region with load-bearing debt, the plan should account for it.

## Ledger

| ID | Location | Category | Severity | Owner | Trigger to revisit | Notes |
|---|---|---|---|---|---|---|
| `<DEBT-001>` | `<path:line or region>` | `<deliberate|drift|shortcut|discovery>` | `<load-bearing|noisy|stale>` | `<name or team>` | `<event or date>` | `<one line>` |

(No entries yet. The **code-debt** sensor wires up once `grep` over the diff finds debt markers; sweep with **debt-review** to triage.)

## Categories

See [`../../sensors/code-debt.md`](../../sensors/code-debt.md) for category and severity definitions. Keep them consistent — the planning phase reads this table verbatim.

## How to use it

- **Before planning a change in a region with load-bearing debt** → factor the debt into the plan. Either pay it down first, route around it, or document why you're adding to it.
- **Before adding new debt** → add the row first (with a trigger to revisit). A debt item without a revisit trigger rots into noise.
- **During audit** → sweep `stale` items for removal and `discovery` items for triage.

## Pruning

`stale` items are deleted during **debt-review**, not archived. The whole point of the ledger is signal density; archiving every fix turns it into a museum.
