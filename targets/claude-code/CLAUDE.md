## Keystone harness

This project uses a **keystone harness**. The corpus at [`harness/`](harness/) defines the engineering knowledge and the six-phase workflow you operate within.

**Read first:**
- [`harness/README.md`](harness/README.md) — five layers (principles, idioms, domain, state, process) and the lifecycle.
- [`harness/adapters/claude-code/`](harness/adapters/claude-code/) — Claude Code bindings: slash commands, sub-agents, MCP tracker integration.
- [`harness/domain/`](harness/domain/) — business rules for this project.

**Lifecycle actions:** `spec` · `orient` · `check-drift` · `verify` · `review` · `learn` (plus `bootstrap`, `audit`, `synthesize`, `mode`). Invoked via `/<prefix>:<action>` slash commands — see [`harness/adapters/claude-code/lifecycle.md`](harness/adapters/claude-code/lifecycle.md).

**Iron laws** — non-negotiable across every phase:

- No proceeding without explicit acceptance criteria in the spec.
- No completion claims without fresh verification evidence — sensors must have run this turn.
- No commits with failing sensors. Never `--no-verify`.
- No AI attribution in commits, PRs, or tracker comments.
- No silent overwrites of state files — propose a diff, confirm before applying.
