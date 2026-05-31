# Claude Code — Lifecycle binding

How each abstract lifecycle action is invoked in Claude Code.

## Action → invocation

| Action | Invocation | Notes |
|---|---|---|
| **spec** | `/<prefix>:spec` slash command | Fetches tracker card via Atlassian / Linear / GitHub MCP server if a card ID is provided. |
| **orient** | `/<prefix>:orient` slash command | Reads `harness/state/CODEBASE_STATE.md`, lazy-loads matching idioms for the touched region. |
| **check-drift** | `/<prefix>:check-drift` slash command | Runs the drift sensor on the current diff. |
| **verify** | `/<prefix>:verify` slash command | Runs lint, type-check, test, build, drift, commit-message sensors via shell. |
| **review** | `/<prefix>:review` slash command | Spawns `review-functional` and `review-security` sub-agents in parallel on the diff. |
| **learn** | `/<prefix>:learn` slash command | Writes a candidate to `harness/learning/inbox/<timestamp>-<slug>.md`. |
| **bootstrap** | `/<prefix>:bootstrap` slash command | One-time initial scaffold. Detects stack, populates `harness/idioms/<stack>/` and `harness/state/`. |
| **audit** | `/<prefix>:audit` slash command | Full dual-flywheel audit (Learning + Pruning). |
| **synthesize** | `/<prefix>:synthesize` slash command | Promotes inbox items into the right corpus layer. |
| **mode** | `/<prefix>:mode <paired\|solo\|autopilot>` | Updates `harness/process/modes.md` in place. |

`<prefix>` is the project's chosen slash-command namespace. If the corpus was installed via a Claude Code plugin, the prefix is the plugin name; if installed directly into the project, the user picks a short prefix (typical: `harness`, `kit`, or the project's short name).

## Where commands live

The slash command files live in one of two places:

- **As a plugin** — installed via the Claude Code plugin marketplace. The commands are owned by the plugin author; users get them by installing the plugin.
- **In the project** — `.claude/commands/*.md` at the consumer's repo root. The commands are owned by the project; they're versioned alongside `harness/`.

Either path works. The plugin path is recommended for organizations that want to ship the same commands across many projects.

## Sub-agent support

Claude Code supports parallel sub-agents via the Agent tool. The **review** action takes advantage of this — multiple review agents run concurrently on the same diff and combine findings.

## Autonomy and modes

All three pacing modes (paired / solo / autopilot) are supported. The agent reads `harness/process/modes.md` to determine current behavior.

## Capability matrix

| Capability | Supported? |
|---|---|
| Autonomous shell sensor execution | ✓ |
| Sub-agent parallelism | ✓ |
| Autonomy levels (paired/solo/autopilot) | ✓ |
| Lazy-by-region idiom loading | ✓ (via slash command logic + Read tool) |
| Context-reset primitive | ✓ (`/clear` and `/compact`) |
| MCP tracker integration | ✓ (Atlassian, Linear, GitHub Issues, Asana servers exist) |
