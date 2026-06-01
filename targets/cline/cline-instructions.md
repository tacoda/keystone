## Keystone harness

This project uses a **keystone harness**. The corpus at [`harness/`](harness/) defines the engineering knowledge and the six-phase workflow you operate within.

> **Setup:** Cline does not have a rules-file convention. Copy this section into the **Custom Instructions** field of Cline's VS Code settings (or Roo Code's equivalent). This file remains in the repo as a record of what was installed.

**Read first:**
- [`harness/README.md`](harness/README.md) — five layers (principles, idioms, domain, state, process), the lifecycle, and the iron laws.
- [`harness/adapters/cline/`](harness/adapters/cline/) — Cline bindings (shell tool, per-command approval).
- [`harness/domain/`](harness/domain/) — business rules for this project.

**Lifecycle actions:** `spec` · `orient` · `check-drift` · `verify` · `review` · `learn` (plus `bootstrap`, `audit`, `synthesize`, `mode`). Invoke by asking in natural language — see [`harness/adapters/cline/lifecycle.md`](harness/adapters/cline/lifecycle.md).

**Iron laws** — non-negotiable across every phase:

- No proceeding without explicit acceptance criteria in the spec.
- No completion claims without fresh verification evidence — sensors must have run this turn.
- No commits with failing sensors. Never `--no-verify`.
- No AI attribution in commits, PRs, or tracker comments.
- No silent overwrites of state files.
