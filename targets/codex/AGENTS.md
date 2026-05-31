# AGENTS.md

This project uses a **project harness**. Read [`harness/README.md`](harness/README.md) before starting work.

## Five layers of the corpus

- `harness/principles/` — universal engineering rules
- `harness/idioms/` — stack-specific patterns (lazy-loaded by region)
- `harness/domain/` — business rules for this project
- `harness/state/` — empirical map of the codebase right now
- `harness/process/` — six workflow phases (spec → planning → implementation → verification → review → release)

## Lifecycle actions

When asked to run a lifecycle action, read the corresponding phase file from `harness/process/` and follow its activities. Sensors are executed via shell.

| Action | Phase file |
|---|---|
| spec | `harness/process/spec.md` |
| orient | `harness/process/planning.md` |
| check-drift | `harness/process/implementation.md` |
| verify | `harness/process/verification.md` |
| review | `harness/process/review.md` |
| learn | `harness/process/release.md` |

Codex-specific bindings live in `harness/adapters/codex/`.

## Iron laws

- **No proceeding without explicit acceptance criteria** in the spec.
- **No completion claims without fresh verification evidence.**
- **No commits with failing sensors.**
- **No AI attribution** in commits, PRs, or tracker comments.
- **No silent overwrites** of state files.

## Prerequisites the harness assumes

- A way to track work — a tracker card (Jira / Linear / GitHub Issues / Asana), a `TODO.md`, or a conversation.
- Lint / type-check / test / build commands in `harness/state/CODEBASE_STATE.md`.
- A pull-request workflow.
- A CI pipeline (CD is even better).
