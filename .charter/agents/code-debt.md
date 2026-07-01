---
kind: agent
id: code-debt
description: 'Surfaces and categorizes debt in the codebase — the known-suboptimal code the team has chosen to live with, plus the shape of any new deb...'
tags:
  - llm-judgment
tools:
  - Read
  - Grep
---
# Sensor: code-debt

Surfaces and categorizes **debt in the codebase** — the known-suboptimal code the team has chosen to live with, plus the shape of any new debt being added in the current diff.

Paired sensor: [`charter-debt`](charter-debt.md) — covers debt in the charter itself (stale rules, placeholder bootstrap, empty idiom dirs, etc.). The two are tracked separately on purpose.

- **Trigger** — **audit** (codebase-wide) and **review** (diff-scoped).
- **Inputs** — `git grep` for debt markers (`TODO`, `FIXME`, `HACK`, `XXX`, `DEPRECATED`), complexity hotspots from [risk-fingerprint](risk-fingerprint.md), and any debt items already listed in `corpus/state/code-debt.md`.
- **Exit condition** — every surfaced item is categorized (see Categories) and either matched to an existing ledger entry or proposed as a new one.
- **Output** — table of debt items: location, category, severity, owner (if known), whether new in this diff.
- **State writes** — proposes a diff to `corpus/state/code-debt.md`. User accepts or edits.

## Categories

| Category | What it means |
|---|---|
| **deliberate** | Known-suboptimal, chosen as a tradeoff. Has an owner and a trigger for revisit. |
| **drift** | Was once fine; the surrounding code moved and left this stale. |
| **shortcut** | A recent shortcut taken under time pressure. |
| **discovery** | Newly noticed; not yet categorized. |

A `TODO` without a category is automatically `discovery` until **debt-review** triages it.

## Severity

- **load-bearing** — change is dangerous; the surrounding region depends on this exact shape.
- **noisy** — adds review burden but doesn't constrain change.
- **stale** — no longer relevant; can be removed without thought.

## Anti-noise rule

Don't surface `TODO` in vendored or generated code (anything excluded from lint). Don't surface debt the ledger already records *unless* its severity changed.
