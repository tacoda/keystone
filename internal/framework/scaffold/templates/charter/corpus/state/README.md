# State

The empirical, temporal map of the codebase. *What is true right now*, not *what should be true*. Reads change as the code changes; sensors and humans both write here.

## What lives here

Files initialized by the **bootstrap** action:

- `CODEBASE_STATE.md` — activation map per region: which idioms apply, coverage, last review.
- `risk-fingerprints.md` — complexity + coupling + coverage patterns per region.
- `traffic-topology.md` — git signals (churn, recency) + business criticality.
- `quality-radar.md` — five-dimension scorecard, updated by **audit** and **review**.
- `code-debt.md` — ledger of debt in the codebase, updated by **debt-review** and **audit**.
- `charter-debt.md` — ledger of debt in the charter itself, updated by **audit**'s Pruning flywheel.

Plus migration tracking:

- `migrations/active/<NAME>.md` — in-flight migrations (expand/contract, deprecations, refactors-in-progress).
- `migrations/completed/<NAME>.md` — archived migration records.

## Activation

On-demand, like the rest of corpus — but with one always-fires touch-point: the **orient** action loads `CODEBASE_STATE.md` (and any active migrations affecting the touched region) at the start of every planning phase. The agent reads State before planning a change so it knows which regions are well-covered, which are risky, and what migrations are mid-flight.

Beyond **orient**, the agent reads state files when it explicitly needs them (e.g., the [risk-fingerprint sensor](sensors/risk-fingerprint.md) reads `risk-fingerprints.md`).

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

Every meaningful code change. This is the most volatile layer in the charter — corpus pruning is otherwise rare, but `state/` files turn over continuously by design.
