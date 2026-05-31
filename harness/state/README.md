# State

The empirical, temporal map of the codebase. *What is true right now*, not *what should be true*. Reads change as the code changes; sensors and humans both write here.

## What lives here

Files initialized by the **bootstrap** action:

- `CODEBASE_STATE.md` — activation map per region: which idioms apply, coverage, last review.
- `risk-fingerprints.md` — complexity + coupling + coverage patterns per region.
- `traffic-topology.md` — git signals (churn, recency) + business criticality.

Plus migration tracking:

- `migrations/active/<NAME>.md` — in-flight migrations (expand/contract, deprecations, refactors-in-progress).
- `migrations/completed/<NAME>.md` — archived migration records.

## Activation

Ambient, always loaded. The agent reads State before planning a change — it tells the agent which regions are well-covered, which are risky, and what migrations are mid-flight.

## Authorship

**Both** agent and human. Sensors propose updates (drift detection, coverage delta, risk recomputation) as diffs against this layer — the user accepts or edits. Humans edit freely.

Per the scaffolding safety contract, no agent ever silently overwrites a state file. Always diff-then-confirm.

## Update triggers

State updates whenever the state of the code changes:

- After any commit that touches a tracked region — the **learn** or **verify** action proposes a diff.
- After an active migration step — the corresponding `migrations/active/<NAME>.md` updates.
- On demand via the **audit** action.

## Empty by default

This directory ships with templates. The **bootstrap** action writes the initial snapshot from your codebase. The **audit** action reconciles ongoing.

## Conventions

Free-form per file, but every file states the date it was last reconciled, and every migration record states its expected end-state.

## Changes when

Every meaningful code change. This is the most volatile layer in the corpus.
