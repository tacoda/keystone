---
name: charter-debt
description: Surfaces debt in the charter itself — content that exists but isn't earning its keep.
---

# Subagent: charter-debt

Surfaces **debt in the charter itself** — content that exists but isn't earning its keep. Distinct from [`code-debt`](code-debt.md), which covers debt in the codebase.

- **Trigger** — **audit** (full sweep). Cheaper subsets can run on demand.
- **Inputs** — `.agents/` tree, `.agents/references/state/CODEBASE_STATE.md`, `.keystone/lockfile.json`, `git log -- .agents/`, and the outputs of the hooks classified as available in `CODEBASE_STATE.md`.
- **Exit condition** — every category below has been swept; each finding is matched to an existing ledger entry or proposed as a new one.
- **Output** — table of charter-debt items: location, category, severity.
- **State writes** — proposes a diff to `.agents/references/state/charter-debt.md`. User accepts or edits.

## Categories

| Category | What it means | How to detect |
|---|---|---|
| **stale-rule** | A rule no diff has touched in N months. | `git log --since=N.months.ago -- .agents/rules/<file>` is empty. |
| **dead-idiom** | An idiom directory for a stack no longer in `CODEBASE_STATE.md`. | Compare `.agents/references/idioms/*` against detected stacks. |
| **placeholder** | Bootstrap placeholders (`<...>`) left unfilled. | `grep -rE '<[^>]+>' .agents/references/state/`. |
| **failing-hook** | Hook classified as available in `CODEBASE_STATE.md` but errors when invoked. | Try the recorded command; record exit code. |
| **empty-shell** | Directory scaffolded by bootstrap (idioms, learning/inbox) that has no real content. | Directory size + non-README file count. |
| **uncited-policy** | Installed policy whose rules were never referenced in the last N reviews. | Cross-reference policy rule paths against review outputs (or `git log` of review reports). |
| **drifted-state** | `CODEBASE_STATE.md` `last_reconciled` is older than N months OR [stack-drift](../hooks/stack-drift.md) hook flagged divergence. | Frontmatter date + stack-drift output. |

## Severity

- **load-bearing** — the charter gives wrong guidance because of this (rules that contradict current code; hooks that claim to run but don't).
- **noisy** — adds review burden, but agents still produce the right behavior.
- **stale** — safe to delete; no longer earning attention.

## Relationship to the audit command

`audit.md`'s Pruning flywheel is the canonical command that *acts on* this subagent's output: it triages the ledger and proposes deletions or fixes. This subagent produces the candidates; **audit** decides what to do with them.

## Anti-noise rule

Don't surface items the ledger already records *unless* their severity changed. Don't flag a rule as `stale-rule` if a recent commit touched any rule in the same `.agents/rules/<topic>/` directory (sibling-rule touches count as activity).
