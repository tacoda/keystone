# Claude Code — Lifecycle binding

How each abstract lifecycle action is invoked in Claude Code.

## Action → invocation

| Action | Invocation | Notes |
|---|---|---|
| **spec** | `/keystone:spec` slash command | Fetches tracker card via Atlassian / Linear / GitHub MCP server if a card ID is provided. |
| **orient** | `/keystone:orient` slash command | Reads `harness/corpus/state/CODEBASE_STATE.md`, lazy-loads matching idioms for the touched region. |
| **check-drift** | `/keystone:check-drift` slash command | Runs the drift sensor on the current diff. |
| **verify** | `/keystone:verify` slash command | Runs lint, type-check, test, build, drift, commit-message sensors via shell. |
| **review** | `/keystone:review` slash command | Spawns `review-functional`, `review-security`, `review-risk`, and `review-deployment` sub-agents in parallel on the diff. |
| **learn** | `/keystone:learn` slash command | Writes a candidate to `harness/learning/inbox/<timestamp>-<slug>.md`. |
| **bootstrap** | `/keystone:bootstrap` slash command | One-time initial scaffold. Detects stack, frameworks, and libraries; seeds corpus (idioms/<stack>/, state/), paired guides (idioms/<stack>/); confirms sensor commands; **inventories computational guides** (LSPs, formatters, editor enforcement) into `guides/computational/`; **classifies sensors** (records which inferential sensors — review-functional, review-security, review-risk, review-deployment, spec-adherence — this adapter can run, and which computational sensors have a working command). Post-bootstrap, `corpus/state/CODEBASE_STATE.md` lists every guide and sensor wired up. |
| **audit** | `/keystone:audit` slash command | Full dual-flywheel audit (Learning + Pruning). |
| **synthesize** | `/keystone:synthesize` slash command | Promotes inbox items into the right corpus layer. |
| **mode** | `/keystone:mode <paired\|solo\|autopilot>` | Updates `harness/guides/process/modes.md` in place. |

`keystone:` is the slash-command namespace. The shipped commands live under that prefix so they don't collide with project-defined or other-plugin commands.

## Where commands live

The slash command files live in one of two places:

- **As a plugin** — installed via the Claude Code plugin marketplace. The commands are owned by the plugin author; users get them by installing the plugin.
- **In the project** — `.claude/commands/*.md` at the consumer's repo root. The commands are owned by the project; they're versioned alongside `harness/`.

Either path works. The plugin path is recommended for organizations that want to ship the same commands across many projects.

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
