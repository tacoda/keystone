---
name: keystone-harness-debt
description: "Surfaces debt in the harness itself — content that exists but isn't earning its keep."
tools:
  - Read
  - Grep
---
# Sensor: harness-debt

Surfaces **debt in the harness itself** — content that exists but isn't earning its keep. Distinct from [`code-debt`](code-debt.md), which covers debt in the codebase.

- **Trigger** — **audit** (full sweep). Cheaper subsets can run on demand.
- **Inputs** — `harness/` tree, `harness/corpus/state/CODEBASE_STATE.md`, `harness/.keystone.lock`, `git log -- harness/`, and the outputs of the sensors classified as available in `CODEBASE_STATE.md`.
- **Exit condition** — every category below has been swept; each finding is matched to an existing ledger entry or proposed as a new one.
- **Output** — table of harness-debt items: location, category, severity.
- **State writes** — proposes a diff to `corpus/state/harness-debt.md`. User accepts or edits.

## Categories

| Category | What it means | How to detect |
|---|---|---|
| **stale-rule** | A guide no diff has touched in N months. | `git log --since=N.months.ago -- harness/guides/<file>` is empty. |
| **dead-idiom** | An idiom directory for a stack no longer in `CODEBASE_STATE.md`. | Compare `harness/corpus/idioms/*` against detected stacks. |
| **placeholder** | Bootstrap placeholders (`<...>`) left unfilled. | `grep -rE '<[^>]+>' harness/corpus/state/`. |
| **failing-sensor** | Sensor classified as available in `CODEBASE_STATE.md` but errors when invoked. | Try the recorded command; record exit code. |
| **empty-shell** | Directory scaffolded by bootstrap (idioms, learning/inbox) that has no real content. | Directory size + non-README file count. |
| **uncited-policy** | Installed policy whose guides were never referenced in the last N reviews. | Cross-reference policy guide paths against review outputs (or `git log` of review reports). |
| **unresolved-gap** | `harness/adapters/<agent>/<topic>.md` left as a TODO placeholder past install. | `grep -r 'TODO' harness/adapters/`. |
| **drifted-state** | `CODEBASE_STATE.md` `last_reconciled` is older than N months OR [stack-drift](stack-drift.md) sensor flagged divergence. | Frontmatter date + stack-drift output. |

## Severity

- **load-bearing** — the harness gives wrong guidance because of this (rules that contradict current code; sensors that claim to run but don't).
- **noisy** — adds review burden, but agents still produce the right behavior.
- **stale** — safe to delete; no longer earning attention.

## Relationship to the audit action

`audit.md`'s Pruning flywheel is the canonical action that *acts on* this sensor's output — it triages the ledger and proposes deletions or fixes. This sensor produces the candidates; **audit** decides what to do with them.

## Anti-noise rule

Don't surface items the ledger already records *unless* their severity changed. Don't flag a guide as `stale-rule` if a recent commit touched any rule in the same `guides/<topic>/` directory (sibling-rule touches count as activity).
