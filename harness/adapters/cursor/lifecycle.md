# Cursor — Lifecycle binding

How each abstract lifecycle action is invoked in Cursor.

## Action → invocation

Cursor's native invocation surface is the chat — there are no slash commands. Lifecycle actions are invoked by **referencing a `.cursor/rules/<action>.mdc` rule** in the chat (typed as `@<action>` or by asking the agent in plain language) and letting the rule's body drive the work.

| Action | Invocation | Where it lives |
|---|---|---|
| **spec** | `@spec` or "start the spec phase for `<task>`" | `.cursor/rules/spec.mdc` |
| **orient** | `@orient` at task start | `.cursor/rules/orient.mdc` |
| **check-drift** | `@check-drift` after edits | `.cursor/rules/check-drift.mdc` |
| **verify** | `@verify` before commit | `.cursor/rules/verify.mdc` |
| **review** | `@review` after verification gate | `.cursor/rules/review.mdc` |
| **learn** | `@learn` post-merge | `.cursor/rules/learn.mdc` |
| **bootstrap** | "Bootstrap the harness" — one-time | reads `harness/guides/process/` directly |
| **audit** | `@audit` or "audit the corpus" | inline / phase doc |
| **synthesize** | "Synthesize the inbox" | reads `harness/learning/inbox/` |
| **mode** | Edit `harness/guides/process/modes.md` directly | n/a |

## Why rules, not commands

Cursor has no equivalent of Claude Code's slash-command system. Instead, it has **rules** — `.mdc` files in `.cursor/rules/` that can be:

- **Always-applied** (`alwaysApply: true`) — loaded into every chat turn. The harness uses this for the menu pointer (`keystone.mdc`), nothing else.
- **Auto-attached** (matching `globs`) — loaded when the user edits a matching file. Useful for region-scoped idioms.
- **Manual** (referenced via `@<name>`) — loaded only when explicitly named. **Most lifecycle action rules are this kind.**

The lifecycle action's `.mdc` body contains the same instructions a slash command would: which phase docs to read, which sensors to run, which artifacts to produce.

## Sub-agent parallelism

Cursor does not have first-class parallel sub-agents. The **review** action therefore runs review prompts **sequentially** within the same chat — each review concern is its own pass over the diff.

## Modes

Cursor effectively has one autonomy level — the user accepts each tool call in agent mode. The harness's `paired`/`solo`/`autopilot` distinction collapses to **paired** in practice. Treat the `mode` action as informational; the user still confirms each shell command Cursor proposes.

## Capability matrix

| Capability | Supported? |
|---|---|
| Autonomous shell execution | ✓ (in agent mode; ✗ in chat-only mode) |
| Sub-agent parallelism | ✗ |
| Autonomy levels (paired/solo/autopilot) | ✗ — effectively always paired |
| Lazy-by-region rule loading | ✓ (native via `globs:` frontmatter) |
| Context-reset primitive | "New chat" button (also `Cmd+Shift+L`) |
| Tracker integration | none native — paste card content or use the agent's web fetch |
| GitHub integration | partial (via shell-level `gh` commands) |
