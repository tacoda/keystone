## Keystone harness

This project uses a **keystone harness**. The corpus at [`harness/`](harness/) defines the engineering knowledge and the six-phase workflow you operate within.

**Read first:**
- [`harness/README.md`](harness/README.md) — five layers (principles, idioms, domain, state, process) and the lifecycle.
- [`harness/adapters/aider/`](harness/adapters/aider/) — Aider bindings: `/run` and `/test` for sensors, `auto-commits: false` required.
- [`harness/domain/`](harness/domain/) — business rules for this project.

**Lifecycle actions:** `spec` · `orient` · `check-drift` · `verify` · `review` · `learn` (plus `bootstrap`, `audit`, `synthesize`, `mode`). Invoked by asking in natural language — see [`harness/adapters/aider/lifecycle.md`](harness/adapters/aider/lifecycle.md).

**Suggested `.aider.conf.yml`:**

```yaml
read:
  - CONVENTIONS.md
  - harness/README.md
auto-commits: false
```

**Iron laws** — non-negotiable across every phase:

- No proceeding without explicit acceptance criteria in the spec.
- No completion claims without fresh verification evidence — run sensors via `/run` or `/test` in this turn.
- No commits with failing sensors.
- No AI attribution in commits, PRs, or tracker comments.
- No silent overwrites of state files.
