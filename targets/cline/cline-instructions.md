# Cline / Roo Code — custom instructions

> **How to use:** Cline does not have a rules-file convention. Copy the content below into the **Custom Instructions** field of Cline's VS Code settings (or Roo Code's equivalent). This file itself stays in the repo as a record of what was installed.

---

This project uses a project harness. Read `harness/README.md` before starting work.

**Corpus layers:**

- `harness/principles/` — universal engineering rules
- `harness/idioms/` — stack-specific patterns
- `harness/domain/` — business rules for this project
- `harness/state/` — empirical map of the codebase right now
- `harness/process/` — six workflow phases (spec → planning → implementation → verification → review → release)

**Lifecycle actions** — when asked to run one, read the corresponding phase file from `harness/process/` and follow its activities:

- spec → `harness/process/spec.md`
- orient → `harness/process/planning.md`
- check-drift → `harness/process/implementation.md`
- verify → `harness/process/verification.md`
- review → `harness/process/review.md`
- learn → `harness/process/release.md`

**Iron laws:**

- No proceeding without explicit acceptance criteria in the spec.
- No completion claims without fresh verification evidence.
- No commits with failing sensors.
- No AI attribution in commits, PRs, or tracker comments.
- No silent overwrites of state files.

Cline-specific bindings: `harness/adapters/cline/`.
