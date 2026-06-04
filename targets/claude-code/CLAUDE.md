## Keystone harness

This project uses a **keystone harness**. The harness at [`harness/`](harness/) defines the engineering knowledge, rules, sensors, and self-update flywheels you operate within.

**Read first:**
- [`harness/README.md`](harness/README.md) — five components (corpus, guides, sensors, policies, flywheels) and the lifecycle.
- [`harness/guides/`](harness/guides/) — rules. **Always loaded.** What you must do and not do.
- [`harness/corpus/`](harness/corpus/) — informational reference. **On-demand.** Reasoning behind the rules; reach via forward-link from a guide.
- [`harness/adapters/claude-code/`](harness/adapters/claude-code/) — Claude Code bindings: slash commands, sub-agents, MCP tracker integration.
- [`harness/corpus/domain/`](harness/corpus/domain/) and [`harness/guides/domain/`](harness/guides/domain/) — business knowledge and rules for this project.

**Lifecycle actions:** `spec` · `orient` · `check-drift` · `verify` · `review` · `learn` (plus `bootstrap`, `audit`, `synthesize`, `mode`). Invoked via `/keystone:<action>` slash commands — see [`harness/adapters/claude-code/lifecycle.md`](harness/adapters/claude-code/lifecycle.md).

**Iron laws** — non-negotiable across every phase:

- No proceeding without explicit acceptance criteria in the spec.
- No completion claims without fresh verification evidence — sensors must have run this turn.
- No commits with failing sensors. Never `--no-verify`.
- No AI attribution in commits, PRs, or tracker comments.
- No silent overwrites of state files — propose a diff, confirm before applying.
