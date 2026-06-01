## Keystone harness

This project uses a **keystone harness**. The corpus at [`harness/`](harness/) defines the engineering knowledge and the six-phase workflow you operate within.

**Read first:**
- [`harness/README.md`](harness/README.md) — five layers (principles, idioms, domain, state, process), the lifecycle, and the iron laws.
- [`harness/adapters/pi/`](harness/adapters/pi/) — pi.dev bindings; also see prompts under `.pi/prompts/`.
- [`harness/domain/`](harness/domain/) — business rules for this project.

**Lifecycle actions:** `spec` · `orient` · `check-drift` · `verify` · `review` · `learn` (plus `bootstrap`, `audit`, `synthesize`, `mode`). Invoke by asking in natural language or via the matching `.pi/prompts/<action>.md` — see [`harness/adapters/pi/lifecycle.md`](harness/adapters/pi/lifecycle.md).

**Iron laws** — non-negotiable across every phase:

- No proceeding without explicit acceptance criteria in the spec.
- No completion claims without fresh verification evidence — sensors must have run this turn.
- No commits with failing sensors. Never `--no-verify`.
- No AI attribution in commits, PRs, or tracker comments.
- No silent overwrites of state files.
