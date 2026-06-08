## Keystone harness

This project uses a **keystone harness**. The harness at [`harness/`](harness/) defines the engineering knowledge, rules, sensors, and self-update flywheels you operate within.

**Read first:**
- [`harness/README.md`](harness/README.md) — five components (corpus, guides, sensors, policies, flywheels), the lifecycle, and the iron laws.
- [`harness/guides/`](harness/guides/) — rules. **Always loaded.** What you must do and not do.
- [`harness/corpus/`](harness/corpus/) — informational reference. **On-demand.** Reasoning behind the rules; reach via forward-link from a guide.
- [`harness/adapters/codex/`](harness/adapters/codex/) — Codex bindings (shell execution, sandbox model).
- [`harness/corpus/domain/`](harness/corpus/domain/) — business rules for this project.

**Lifecycle** — to kick off a unit of work, say "**run task on `<ticket-id>`**" (runs the **task playbook**). To invoke any single action, ask in natural language ("run verify," "do a review pass"). Plugins (vendored at `harness/plugins/<name>/`) can ship playbooks, actions, and guides. Override: your project files at `harness/<kind>/<name>.md` always win by default. Among plugins, outer plugins (shallower in `keystone.json`) win over plugins nested inside them. A plugin can mark an item `strict` to make it absolute — nothing else can override a strict item, not the project, not any other plugin.

**Playbooks** ([`harness/playbooks/`](harness/playbooks/)):

- **[task](harness/playbooks/task.md)** — end-to-end workflow: spec → orient → implementation → check-drift → verify → review.

**Actions** ([`harness/actions/`](harness/actions/)):

- **[bootstrap](harness/actions/bootstrap.md)** — one-time scaffold; detect stack, seed state, classify sensors. Run once per project.
- **[spec](harness/actions/spec.md)** — capture intent + acceptance criteria. First action on any task.
- **[orient](harness/actions/orient.md)** — load `CODEBASE_STATE.md` and idioms for the touched region; sketch a plan.
- **[check-drift](harness/actions/check-drift.md)** — compare the diff against loaded guides; fast pre-verify check.
- **[verify](harness/actions/verify.md)** — run lint / type-check / test / build / drift / commit-message sensors.
- **[review](harness/actions/review.md)** — semantic review (functional / security / risk / deployment + spec-adherence).
- **[learn](harness/actions/learn.md)** — capture an inbox candidate from a surprise or incident.
- **[audit](harness/actions/audit.md)** — periodic dual-flywheel: Learning + Pruning.
- **[synthesize](harness/actions/synthesize.md)** — promote inbox candidates into the right corpus / guide layer.
- **[mode](harness/actions/mode.md)** — switch pacing (paired / solo / autopilot).

**Iron laws** — non-negotiable across every phase:

- No proceeding without explicit acceptance criteria in the spec.
- No completion claims without fresh verification evidence — sensors must have run this turn.
- No commits with failing sensors. Never `--no-verify`.
- No AI attribution in commits, PRs, or tracker comments.
- No silent overwrites of state files.
