# Process

Phase definitions and gate conditions for the development workflow. "What happens at each step" of how this project ships work.

## What lives here

Six phases:

- `spec.md` — capturing intent. Entry: an idea or tracker card. Gate: approved spec with explicit acceptance criteria.
- `planning.md` — turning the approved spec into a plan. Entry: spec approved. Gate: plan approved.
- `implementation.md` — writing the change. Entry: plan approved. Gate: tests pass, sensors clean.
- `verification.md` — mechanical correctness check. Entry: implementation gate passed. Gate: every sensor green this turn.
- `review.md` — semantic correctness check (spec adherence + review agents + PR comments). Entry: verification gate passed. Gate: no blockers, AC met.
- `release.md` — landing and announcing. Entry: review gate passed. Gate: change deployed and recorded.

Two cross-cutting references:

- `modes.md` — pacing modes (paired / solo / autopilot). Orthogonal to the six phases; modes change *how* phases run, not *what* happens.
- `sensors.md` — sensor contracts (lint, type-check, test, build, drift, coverage, risk-fingerprint, traffic-topology, state-region, commit-message, tracker-card-fetcher). Each sensor declares its trigger, inputs, exit condition, output, and any state writes.

Cross-cutting discipline (apply across every phase):

- `sensitive-files.md` — files the agent must never read or write. Sensitive-data hygiene.
- `dangerous-actions.md` — irreversible operations requiring explicit confirmation. Counterpart to `modes.md`.
- `scoping.md` — size limits on commits and PRs.
- `ci-failure.md` — what to do when CI fails. Sibling of `release.md` (the happy path).
- `escalation.md` — when to stop and ask. Counterpart to `modes.md`.

Agent-specific failure modes (loaded ambient; counter the natural pull of how coding agents go wrong):

- `grounding.md` — verify that a function, package, flag, or config key exists before invoking it.
- `pushback.md` — disagree explicitly when the user is wrong; do not collapse to agreement.
- `self-validation.md` — refuse to count the agent's own claim as evidence; tool output is evidence.
- `subagent-trust.md` — a subagent's "done" report is a claim to verify, not evidence to accept.
- `context-budget.md` — read what is relevant to the touched region, no more; grep before reading.
- `surgical-edits.md` — touch only what serves the task; no "while I'm here" cleanups.

Each phase doc enumerates:

- **Entry condition** — what must be true to enter this phase.
- **Activities** — the steps that happen here, including which lifecycle commands fire.
- **Sensors** — what runs during the phase (full contracts in `sensors.md`).
- **Gate condition** — what must be true to exit.
- **Artifacts** — what gets written (specs, plans, ADRs, threat models, commits, postmortems) and where.

## Tracker card threading

The **spec** phase captures a tracker card identifier (Jira, Linear, GitHub Issues, Asana, etc.) when one is provided. The reference lives in the spec's frontmatter and is carried forward by every downstream phase — PR descriptions, postmortems, follow-up cards, learning captures all link back to the original card.

If no tracker card is provided, each phase asks the user for the relevant input (spec content, plan deviations, review approvers, release notes) directly.

## Activation

A phase doc is loaded when the agent enters that phase. Not ambient — the **orient** action reads `planning.md` at task start; the agent stays inside the active phase doc until its gate passes.

## Lifecycle actions

Phases reference lifecycle actions the agent invokes:

| Phase | Entry action | Exit action |
|---|---|---|
| Planning | **orient** | (review plan with user) |
| Implementation | — | **check-drift** |
| Verification | **verify** | (sensors gate the commit) |
| Release | — | **learn** |

How each action is actually invoked is agent-specific — slash command, rules-file trigger, CLI subcommand, etc. See `harness/adapters/<your-agent>/lifecycle.md` for the binding.

## Authorship

Lead engineer drafts; agent refines through Learning flywheel. Process changes when the team's working style changes.

## Pacing modes

Set by invoking the **mode** action with `<paired|solo|autopilot>`. Modes change the user-facing pace; the phases themselves are unchanged.

## Artifact locations (convention)

| Artifact | Location |
|---|---|
| Spec | `docs/specs/YYYY-MM-DD-<topic>.md` |
| Plan | `docs/plans/YYYY-MM-DD-<topic>.md` |
| RFC | `docs/rfcs/NNNN-<slug>.md` |
| ADR | `docs/adrs/NNNN-<slug>.md` |
| Threat model | `docs/threats/<feature>.md` |
| Postmortem | `docs/postmortems/YYYY-MM-DD-<slug>.md` |

## Changes when

The process matures. New phase added, gate tightened, artifact location changed → `process/` files update.
