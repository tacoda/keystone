## Keystone harness

This project uses a **keystone harness**. The harness at [`harness/`](harness/) defines the engineering knowledge, rules, sensors, and self-update flywheels you operate within.

**Read first:**
- [`harness/README.md`](harness/README.md) — five components (corpus, guides, sensors, policies, flywheels), the lifecycle, and the iron laws.
- [`harness/guides/`](harness/guides/) — rules. **Always loaded.** What you must do and not do.
- [`harness/corpus/`](harness/corpus/) — informational reference. **On-demand.** Reasoning behind the rules; reach via forward-link from a guide.
- [`harness/adapters/aider/`](harness/adapters/aider/) — Aider bindings: `/run` and `/test` for sensors, `auto-commits: false` required.
- [`harness/corpus/domain/`](harness/corpus/domain/) — business rules for this project.

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
