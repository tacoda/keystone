# AGENTS.md

> **Note:** If pi.dev reads a different file (e.g. `.pirules`, `pi.config`, etc.), rename this file accordingly. `AGENTS.md` is the cross-agent default.

This project uses a **project harness**. Read [`harness/README.md`](harness/README.md) before starting work.

## Five layers of the corpus

- `harness/principles/` — universal engineering rules
- `harness/idioms/` — stack-specific patterns
- `harness/domain/` — business rules for this project
- `harness/state/` — empirical map of the codebase right now
- `harness/process/` — six workflow phases

## Lifecycle actions

When asked to run an action, read the corresponding phase file from `harness/process/` and follow its activities.

- spec → `harness/process/spec.md`
- orient → `harness/process/planning.md`
- check-drift → `harness/process/implementation.md`
- verify → `harness/process/verification.md`
- review → `harness/process/review.md`
- learn → `harness/process/release.md`

pi.dev-specific bindings live in `harness/adapters/pi/`.

## Iron laws

- No proceeding without explicit acceptance criteria in the spec.
- No completion claims without fresh verification evidence.
- No commits with failing sensors.
- No AI attribution in commits, PRs, or tracker comments.
- No silent overwrites of state files.
