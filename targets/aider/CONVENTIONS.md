## Keystone harness

This project uses a **keystone harness**. The harness at [`harness/`](harness/) defines the engineering knowledge, rules, sensors, and self-update flywheels you operate within.

**Read first:**
- [`harness/README.md`](harness/README.md) — five components (corpus, guides, sensors, policies, flywheels), the lifecycle, and the iron laws.
- [`harness/guides/`](harness/guides/) — rules. **Always loaded.** What you must do and not do.
- [`harness/corpus/`](harness/corpus/) — informational reference. **On-demand.** Reasoning behind the rules; reach via forward-link from a guide.
- [`harness/adapters/aider/`](harness/adapters/aider/) — Aider bindings: `/run` and `/test` for sensors, `auto-commits: false` required.
- [`harness/corpus/domain/`](harness/corpus/domain/) — business rules for this project.

**Lifecycle actions** — invoke by asking in natural language (see [`harness/adapters/aider/lifecycle.md`](harness/adapters/aider/lifecycle.md) for the full table):

- **bootstrap** — one-time scaffold; detect stack, seed state, classify sensors. Run once per project.
- **spec** — capture intent + acceptance criteria. First action on any task.
- **orient** — load `CODEBASE_STATE.md` and idioms for the touched region; sketch a plan.
- **check-drift** — compare the diff against loaded guides; fast pre-verify check.
- **verify** — run lint / type-check / test / build / drift / commit-message sensors via `/run` or `/test`.
- **review** — semantic review (functional / security / risk / deployment + spec-adherence).
- **learn** — capture an inbox candidate from a surprise or incident.
- **audit** — periodic dual-flywheel: Learning + Pruning.
- **synthesize** — promote inbox candidates into the right corpus / guide layer.
- **mode** — switch pacing (paired / solo / autopilot).

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
