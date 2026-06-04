# Claude Code — Lifecycle binding

How each abstract lifecycle action is invoked in Claude Code.

## Action → invocation

Each action ships as a project-scoped **skill** under `.claude/skills/keystone/<action>/SKILL.md`. Invoke by name (e.g. `keystone:bootstrap`) or by asking the agent to "run the bootstrap action" — Claude Code auto-discovers skills under `.claude/skills/` and surfaces them.

| Action | Skill | Notes |
|---|---|---|
| **bootstrap** | `keystone:bootstrap` | One-time initial scaffold. Detects stack, frameworks, and libraries; seeds corpus (idioms/<stack>/, state/), paired guides (idioms/<stack>/); confirms sensor commands; **inventories computational guides** (LSPs, formatters, editor enforcement) into `guides/computational/`; **classifies sensors** (records which inferential sensors — review-functional, review-security, review-risk, review-deployment, spec-adherence — this adapter can run, and which computational sensors have a working command). Post-bootstrap, `corpus/state/CODEBASE_STATE.md` lists every guide and sensor wired up. |
| **spec** | `keystone:spec` | Fetches tracker card via Atlassian / Linear / GitHub MCP server if a card ID is provided. |
| **orient** | `keystone:orient` | Reads `harness/corpus/state/CODEBASE_STATE.md`, lazy-loads matching idioms for the touched region. |
| **check-drift** | `keystone:check-drift` | Runs the drift sensor on the current diff. |
| **verify** | `keystone:verify` | Runs lint, type-check, test, build, drift, commit-message sensors via shell. |
| **review** | `keystone:review` | Spawns `review-functional`, `review-security`, `review-risk`, and `review-deployment` sub-agents in parallel on the diff. |
| **learn** | `keystone:learn` | Writes a candidate to `harness/learning/inbox/<timestamp>-<slug>.md`. |
| **audit** | `keystone:audit` | Full dual-flywheel audit (Learning + Pruning). |
| **synthesize** | `keystone:synthesize` | Promotes inbox items into the right corpus layer. |
| **mode** | `keystone:mode` | Updates `harness/guides/process/modes.md` in place. Arg: `paired`, `solo`, or `autopilot`. |

`keystone:` is the namespace prefix. The shipped skills live under that prefix so they don't collide with project-defined or other-plugin skills.

## Where the skills live

The skill files live at `.claude/skills/keystone/<action>/SKILL.md` in the consumer repo. `keystone init` (or `keystone target add claude-code`) lays them down; they're versioned alongside `harness/`.

Organizations that prefer a single source of truth across many repos can lift the same files into a Claude Code plugin and distribute via the plugin marketplace instead — the on-disk shape is identical.

## Sub-agent support

Claude Code supports parallel sub-agents via the Agent tool. The **review** action takes advantage of this — multiple review agents run concurrently on the same diff and combine findings.

## Autonomy and modes

All three pacing modes (paired / solo / autopilot) are supported. The agent reads `harness/guides/process/modes.md` to determine current behavior.

## Capability matrix

| Capability | Supported? |
|---|---|
| Autonomous shell sensor execution | ✓ |
| Sub-agent parallelism | ✓ |
| Autonomy levels (paired/solo/autopilot) | ✓ |
| Lazy-by-region idiom loading | ✓ (via slash command logic + Read tool) |
| Context-reset primitive | ✓ (`/clear` and `/compact`) |
| MCP tracker integration | ✓ (Atlassian, Linear, GitHub Issues, Asana servers exist) |
