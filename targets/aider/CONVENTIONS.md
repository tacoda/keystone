# Conventions — Keystone harness

This project uses a project harness. Read `harness/README.md` before starting work.

## Five layers of the corpus

- `harness/principles/` — universal engineering rules
- `harness/idioms/` — stack-specific patterns
- `harness/domain/` — business rules for this project
- `harness/state/` — empirical map of the codebase right now
- `harness/process/` — six workflow phases (spec → planning → implementation → verification → review → release)

## Lifecycle actions

When asked to run a lifecycle action, read the corresponding phase file from `harness/process/` and follow its activities.

| Action | Phase file |
|---|---|
| spec | `harness/process/spec.md` |
| orient | `harness/process/planning.md` |
| check-drift | `harness/process/implementation.md` |
| verify | `harness/process/verification.md` |
| review | `harness/process/review.md` |
| learn | `harness/process/release.md` |

Aider-specific bindings live in `harness/adapters/aider/`.

## Iron laws

- **No proceeding without explicit acceptance criteria** in the spec.
- **No completion claims without fresh verification evidence** — run sensors via `/run` or `/test` in this turn.
- **No commits with failing sensors.**
- **No AI attribution** in commits.
- **No silent overwrites** of state files.

## Suggested `.aider.conf.yml`

```yaml
read:
  - CONVENTIONS.md
  - harness/README.md
auto-commits: false
```
