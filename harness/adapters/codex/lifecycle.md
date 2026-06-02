# Codex CLI — Lifecycle binding

How each abstract lifecycle action is invoked in OpenAI's Codex CLI. Codex reads `AGENTS.md` at the repo root on every session and runs shell commands autonomously. It has no slash commands and no built-in sub-agents — every lifecycle action is invoked by asking the agent in natural language.

## Action → invocation

| Action | Invocation |
|---|---|
| **spec** | "Start the spec phase for `<task or tracker card>`." Codex reads `harness/guides/process/spec.md` and follows it. |
| **orient** | "Orient for work in `<region>`." |
| **check-drift** | "Check the current diff for drift against the corpus." |
| **verify** | "Run the verify action." Codex executes lint / type-check / test / build / coverage via shell. |
| **review** | "Run the review action." Sequential (no native sub-agent parallelism). |
| **learn** | "Capture the learnings from this work." |
| **bootstrap** | "Bootstrap the harness." One-time, seeds corpus (idioms/<stack>/, state/), paired guides (idioms/<stack>/), and confirms sensor commands. |
| **audit** | "Audit the corpus." |
| **synthesize** | "Synthesize the inbox." |
| **mode** | "Set pacing mode to `<paired\|solo\|autopilot>`." Edit `harness/guides/process/modes.md` directly. |

Because Codex has no slash-command convention, lifecycle actions are invoked by the human typing the action name into the chat. The shipped `AGENTS.md` tells Codex which corpus file to read for each action.

## Capability matrix

| Capability | Supported? |
|---|---|
| Autonomous shell execution | ✓ |
| Sub-agent parallelism | ✗ |
| Autonomy levels | ✓ (via `--ask-for-approval`, `--full-auto`, `--auto-edit`, etc.) |
| Lazy-by-region | ✗ (Codex reads files on demand from `harness/`; no glob-based auto-load) |
| Context-reset primitive | new session (no in-session `/clear`) |
| Project-scoped settings | partial (`~/.codex/` global; project AGENTS.md is the per-project surface) |

## Pacing modes ↔ Codex approval modes

Codex's CLI flags map naturally to the harness's pacing modes:

| Keystone mode | Suggested Codex invocation |
|---|---|
| **paired** | `codex` (default — asks before risky actions) or `codex --ask-for-approval` |
| **solo** | `codex --auto-edit` (writes files autonomously; stops on shell commands by default) |
| **autopilot** | `codex --full-auto` (writes files + runs shell without per-action approval) |

`harness/guides/process/modes.md` documents this mapping in the `mode` action body.

## Sub-agent degradation

The **review** action would ideally spawn `review-functional` and `review-security` in parallel. Codex doesn't support this natively. When invoked, the review action runs them sequentially in one conversation, then combines findings.

## Tracker integration

Codex does not bundle MCP-style tracker integrations. Use the agent's shell access to fetch cards:

- GitHub Issues: `gh issue view <id>`
- Linear: `linear` CLI or paste card content
- Jira: `jira` CLI or paste card content

The **spec** action's tracker-fetcher sensor falls back to "paste the card" if no CLI is configured.
