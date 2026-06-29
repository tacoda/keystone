## Keystone harness

This project uses a **keystone harness**. Its primitives — guides,
sensors, hooks, agents, commands, skills, playbooks, patterns, corpus,
documents, concerns, posture, tools — all live under
[`.harness/`](.harness/). Discover what's available through the index;
open primitive bodies on demand.

**Read first:**
[`.harness/INDEX.json`](.harness/INDEX.json) — one entry per primitive,
each with `kind`, `id`, `description`, and `path`. Open `path` only when
you decide to activate the primitive.

**Activate by:**

| Kind         | When to open                                                            |
| ------------ | ----------------------------------------------------------------------- |
| **guide**    | Touched files match the entry's `globs:` (or no globs declared). Inferential → a directive; computational → a host hook (LSP). |
| **corpus**   | A guide's `corpus:` (or a prose forward-link) points at it — the *why*. |
| **command**  | User's intent matches `description` + `phase`; a unit of work. Host slash: `/keystone-<id>`. |
| **playbook** | A composed sequence of commands with human `gates:`.                    |
| **sensor**   | An inferential review at a gate — dispatched as an agent, returns a `returns:` verdict. |
| **hook**     | Fires deterministically on an `event:` (host phase or framework event) → `run:` shell or `agent:` dispatch. |
| **skill**    | Claude Code auto-activates by `triggers:` match.                        |
| **agent**    | Spawn via the Task tool by `id`. The system prompt is the body.         |
| **pattern**  | A reusable documentation pattern (Diátaxis) — apply when writing docs.  |

**Lifecycle** — to kick off a unit of work, say "**run task on
`<ticket-id>`**" (runs the **task** playbook). For any single command,
ask in natural language ("run verify", "do a review pass") — the
command's body lives at its INDEX `path`.

**Iron laws** — non-negotiable across every phase:

- No proceeding without explicit acceptance criteria.
- No completion claims without fresh verification — checks must have run
  this turn, against the post-edit state, with cited output.
- No commits with failing checks. Never `--no-verify`.
- No AI attribution in commits, PRs, or tracker comments.
- No silent overwrites of state files.

**Override** — your project files at `.harness/<kind>/<id>.md` always win
by default. Among installed policies, policies nested deeper in
`keystone.json` refine outer policies. A policy can mark an item `strict`
to make it absolute — nothing else can override a strict item.
