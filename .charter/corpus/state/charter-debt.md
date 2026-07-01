---
kind: corpus
id: corpus/state/charter-debt
description: 'Paired ledger: [code-debt.'
last_reviewed: 2026-06-18
---
# Charter Debt Ledger

Populated by the **audit** action's Pruning flywheel from the [charter-debt sensor](../../sensors/charter-debt.md). Hand-edits during audit are OK.

Paired ledger: [`code-debt.md`](code-debt.md) — debt in the codebase. Tracked separately on purpose.

The debt the charter itself is carrying. When the charter gives stale or wrong guidance, that cost lands on every task the agent runs against this project. This ledger makes that cost visible so the team can pay it down deliberately.

## Ledger

| ID | Location | Category | Severity | Notes |
|---|---|---|---|---|
| HDEBT-001 | `.charter/archive/`, `.charter/sources/`, `.charter/evals/` | empty-shell | noisy | README-only auxiliary dirs not in the twelve-kind taxonomy. Either drop or mark as auxiliary in each README. |
| HDEBT-002 | `.charter/corpus/state/code-debt.md`, `charter-debt.md` | drifted-state | noisy | Internal links used `sensors/...` (pre-2.0 path); fixed to `../../sensors/...`. Watch for re-introduction in template copies. |
| HDEBT-003 | `internal/framework/scaffold/templates/charter/corpus/state/INSTALL_PROFILE.md` | placeholder | load-bearing | Was missing from templates; fresh installs wouldn't ship the file. Added 2026-06-18. |
| HDEBT-004 | `internal/framework/scaffold/templates/patches/1.0.3/`, `1.0.4/` | stale-rule | stale | 1.x patches removed 2026-06-18 (min supported = 2.0). Verify no upgrader still requests them. |
| HDEBT-005 | `docs/ports/` | placeholder | noisy | Was missing `persona`, `skill`, `subagent`, `command`, `rule`, `computational`. Stubs added 2026-06-18; expand as semantics solidify. |
| HDEBT-006 | `.github/workflows/` | unresolved-gap | load-bearing | Only `release.yml` wired. PR / push CI added 2026-06-18 (`ci.yml`). Verify on next PR. |

## Categories

See [`../../sensors/charter-debt.md`](../../sensors/charter-debt.md) for category and severity definitions.

## How to use it

- **Before any task** → if the touched region has a `load-bearing` charter-debt item, address it first (the agent will be operating on bad guidance otherwise).
- **During audit** → triage `placeholder` and `failing-sensor` items as `load-bearing` by default; they're the most likely to mislead.
- **During synthesize** → before promoting an inbox item into a guide, check that no `stale-rule` already covers the topic. Adding a new rule next to a stale one compounds the noise.

## Pruning

`stale` items are deleted during **audit**, not archived. The charter lives in git — history is already preserved.
