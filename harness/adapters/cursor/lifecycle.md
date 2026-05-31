# Cursor — Lifecycle binding

> **Stub.** Fill in the bindings as you wire Cursor up. The shape below is the expected starting point.

## Action → invocation

Cursor activates rules via `.cursor/rules/*.mdc` files with glob frontmatter. Each lifecycle action becomes a rules file that triggers when the human types a matching phrase or edits a matching path.

| Action | Suggested binding |
|---|---|
| **spec** | `.cursor/rules/spec.mdc` — `alwaysApply: false`; user invokes by asking "start the spec phase for `<task>`." |
| **orient** | `.cursor/rules/orient.mdc` — glob `**/*`; loads on every edit; pulls `harness/process/planning.md`. |
| **check-drift** | `.cursor/rules/check-drift.mdc` — user-invoked. |
| **verify** | `.cursor/rules/verify.mdc` — user-invoked; surfaces sensor commands for the user to run. |
| **review** | `.cursor/rules/review.mdc` — user-invoked; loads `review-*` agent prompts sequentially (Cursor has no sub-agent parallelism). |
| **learn** | `.cursor/rules/learn.mdc` — user-invoked at end of work. |
| **bootstrap** | One-time; user asks Cursor to bootstrap by reading `harness/README.md`. |
| **audit** | User-invoked. |
| **synthesize** | User-invoked. |
| **mode** | Edit `harness/process/modes.md` directly. |

## Capability matrix

| Capability | Supported? |
|---|---|
| Autonomous shell sensor execution | partial (Cursor agent mode can run commands; Cursor chat mode cannot) |
| Sub-agent parallelism | ✗ |
| Autonomy levels | ✗ (effectively one mode) |
| Lazy-by-region | ✓ (via `.mdc` glob frontmatter) |
| Context-reset primitive | "New chat" button |
