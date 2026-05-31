# Copilot instructions — Project harness

This project uses a project harness. Read `harness/README.md` before starting work.

## Five layers of the corpus

- `harness/principles/` — universal engineering rules
- `harness/idioms/` — stack-specific patterns
- `harness/domain/` — business rules for this project
- `harness/state/` — empirical map of the codebase right now
- `harness/process/` — six workflow phases (spec → planning → implementation → verification → review → release)

## Lifecycle actions

When asked to run an action, read the corresponding phase file from `harness/process/` and follow its activities.

- spec → `harness/process/spec.md`
- orient → `harness/process/planning.md`
- check-drift → `harness/process/implementation.md`
- verify → `harness/process/verification.md`
- review → `harness/process/review.md`
- learn → `harness/process/release.md`

GitHub Copilot CLI-specific bindings live in `harness/adapters/github-copilot-cli/`.

## Iron laws

- No proceeding without explicit acceptance criteria in the spec.
- No completion claims without fresh verification evidence.
- No commits with failing sensors.
- No AI attribution in commits, PRs, or tracker comments.
- No silent overwrites of state files.

## Tracker integration

Use `gh issue view`, `gh pr view`, and `gh pr comment` via the shell. Cards land in the spec's frontmatter and carry through every downstream artifact.
