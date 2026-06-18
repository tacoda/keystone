# Codex CLI — Lifecycle binding

How each abstract lifecycle action is invoked in OpenAI's Codex CLI. Codex reads `AGENTS.md` at the repo root on every session and runs shell commands autonomously. It has no slash commands and no built-in sub-agents — every lifecycle action is invoked by asking the agent in natural language.

## Invocation

Every action is invoked via natural language: "run task on TICKET-123," "run verify," "do a review pass." The agent reads `AGENTS.md` at session start, finds the action in the bulleted list, follows the link to `harness/actions/<action>.md`, and executes the playbook. Codex has no slash-command primitive — natural language is the only invocation surface, which is why action playbooks are agent-agnostic and live in `harness/actions/`.

The canonical kickoff phrase is **"run task on `<ticket-id>`"** (or "run the task workflow") — `harness/actions/task.md` orchestrates `spec → orient → implementation → check-drift → verify → review`.

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

The **review** action would ideally spawn `review-functional`, `review-security`, `review-risk`, and `review-deployment` in parallel. Codex doesn't support this natively. When invoked, the review action runs them sequentially in one conversation, then combines findings.

## Tracker integration

Codex does not bundle MCP-style tracker integrations. Use the agent's shell access to fetch cards:

- GitHub Issues: `gh issue view <id>`
- Linear: `linear` CLI or paste card content
- Jira: `jira` CLI or paste card content

The **spec** action's tracker-fetcher sensor falls back to "paste the card" if no CLI is configured.
