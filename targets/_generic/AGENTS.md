# AGENTS.md

This project uses a **project harness** — a project-scoped corpus of engineering knowledge that drives any coding agent through a structured workflow.

## Where to start

Read [`harness/README.md`](harness/README.md). It describes the five layers of the corpus and the lifecycle actions you'll be invoking.

## The corpus, in one paragraph

- `harness/principles/` — universal engineering rules (always loaded).
- `harness/idioms/` — stack-specific patterns (loaded when editing matching paths).
- `harness/domain/` — what this project ships and the business rules it enforces (always loaded).
- `harness/state/` — empirical map of the codebase right now (always loaded).
- `harness/process/` — six phases: spec → planning → implementation → verification → review → release. Loaded when you enter a phase.

## How you'll be invoked

Lifecycle actions:

- **spec** — capture intent and acceptance criteria from a tracker card or inline.
- **orient** — load idioms and state for the touched region before planning.
- **check-drift** — compare the diff against corpus rules during implementation.
- **verify** — run every sensor (lint, type-check, test, build, drift, commit-message) before commit. **Fresh evidence per turn — no stale claims.**
- **review** — check spec adherence + run review agents (functional, security) on the diff.
- **learn** — capture novel judgment calls to `harness/learning/inbox/` for the next synthesis cycle.

If your agent doesn't have specific bindings yet, see [`harness/adapters/_generic/`](harness/adapters/_generic/) for the fallback model.

## Iron laws

These are non-negotiable across every phase:

- **No proceeding without explicit acceptance criteria** in the spec.
- **No completion claims without fresh verification evidence** — sensors must have run in this turn.
- **No commits with failing sensors.** No `--no-verify`.
- **No AI attribution** in commit messages, PR descriptions, or tracker comments.
- **No silent overwrites** of state files — always diff-then-confirm.

## Prerequisites the harness assumes

- A way to track work — a tracker card (Jira / Linear / GitHub Issues / Asana), a `TODO.md`, or a conversation.
- Lint, type-check, test, and build commands defined in `harness/state/CODEBASE_STATE.md`.
- A pull-request workflow.
- A CI pipeline (CD is even better).

Missing one degrades the corresponding phase but does not break the harness.
