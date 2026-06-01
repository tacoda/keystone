## Keystone harness

This project uses a **keystone harness**. The corpus at [`harness/`](../harness/) defines the engineering knowledge and the six-phase workflow you operate within.

**Read first:**
- [`harness/README.md`](../harness/README.md) — five layers (principles, idioms, domain, state, process) and the lifecycle.
- [`harness/adapters/github-copilot/`](../harness/adapters/github-copilot/) — Copilot bindings: native `gh` CLI integration, per-command approval mode.
- [`harness/domain/`](../harness/domain/) — business rules for this project.

**Lifecycle actions:** `spec` · `orient` · `check-drift` · `verify` · `review` · `learn` (plus `bootstrap`, `audit`, `synthesize`, `mode`). Invoked by asking in natural language — see [`harness/adapters/github-copilot/lifecycle.md`](../harness/adapters/github-copilot/lifecycle.md).

**GitHub-native primitives** used by this adapter: `gh issue view` for tracker fetch, `gh run list` / `gh run view` for CI status, `gh pr create` for release. For non-GitHub trackers (Jira, Linear, Asana), paste card content into the chat.

**Iron laws** — non-negotiable across every phase:

- No proceeding without explicit acceptance criteria in the spec.
- No completion claims without fresh verification evidence — sensors must have run this turn.
- No commits with failing sensors. Never `--no-verify`.
- No AI attribution in commits, PRs, or tracker comments — no `Co-Authored-By: Copilot`, no auto-generated footers.
- No silent overwrites of state files — propose a diff, confirm before applying.

Copilot's only autonomy lever is per-command approval. Treat every session as **paired**.
