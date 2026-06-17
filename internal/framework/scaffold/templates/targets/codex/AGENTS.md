## Keystone harness

This project uses a **keystone harness**. Framework primitives — guides,
corpus, sensors, actions, playbooks — plus host-native ones (skills,
subagents, commands, rules) all live under
[`.keystone/harness/`](.keystone/harness/). Discover what's available
through the index; open primitive bodies on demand.

**Read first:**
[`.keystone/INDEX.json`](.keystone/INDEX.json) — one entry per
primitive, each with `kind`, `id`, `description`, and `path`. Open
`path` only when you decide to activate the primitive.

**Activate by:**

| Kind         | When to open                                                            |
| ------------ | ----------------------------------------------------------------------- |
| **guide**    | Touched files match the entry's `globs:` (or no globs declared).        |
| **rule**     | Same as guide — host-native flavor.                                      |
| **corpus**   | A guide's `traces:` points at it.                                       |
| **action**   | User's intent matches `description` + `phase`.                          |
| **playbook** | Composed sequence of actions.                                           |
| **sensor**   | Inside an action, per-phase, narrowed by `globs:`.                      |
| **skill**    | Auto-activate by `triggers:` match.                                     |
| **subagent** | Delegated agent — spawn by `id`; the system prompt is the body.         |
| **command**  | Host slash mechanism: user types `/<id>`.                                |

**Lifecycle** — to kick off a unit of work, say "**run task on
`<ticket-id>`**" (runs the **task** playbook). For any single action,
ask in natural language ("run verify", "do a review pass") — the
action's body lives at its INDEX `path`.

**Iron laws** — non-negotiable across every phase:

- No proceeding without explicit acceptance criteria.
- No completion claims without fresh verification — sensors must have
  run this turn.
- No commits with failing sensors. Never `--no-verify`.
- No AI attribution in commits, PRs, or tracker comments.
- No silent overwrites of state files.

**Override** — your project files at `.keystone/harness/<kind>/<id>.md`
always win by default. Among installed policies, policies nested deeper
in `keystone.json` refine outer policies. A policy can mark an item
`strict` to make it absolute.
