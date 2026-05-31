# Aider — Lifecycle binding

> **Stub.** Fill in the bindings as you wire Aider up.

## Action → invocation

Aider reads `CONVENTIONS.md` (and any file passed via `--read`) on every session. There are no slash commands. Lifecycle actions are invoked by the human typing the action name into the chat.

| Action | Suggested binding |
|---|---|
| **spec** | "Start the spec phase for `<task>`." Aider reads `harness/process/spec.md` and follows it. |
| **orient** | "Orient for work in `<region>`." Reads `harness/state/CODEBASE_STATE.md` + matching idioms. |
| **check-drift** | "Check the diff for drift." |
| **verify** | "Run the verify action." Aider invokes shell commands directly (`/run` or its lint config). |
| **review** | "Run the review action." Sequential — no sub-agent parallelism. |
| **learn** | "Capture the learnings from this work." |
| **bootstrap** | "Bootstrap the harness." |
| **audit** | "Audit the corpus." |
| **synthesize** | "Synthesize the inbox." |
| **mode** | Edit `harness/process/modes.md` directly. |

## Capability matrix

| Capability | Supported? |
|---|---|
| Autonomous shell execution | ✓ (Aider can run `/run <cmd>` and `/test`) |
| Sub-agent parallelism | ✗ |
| Autonomy levels | ✗ |
| Lazy-by-region | ✗ (Aider reads what you pass it; no auto-load by path) |
| Context-reset primitive | `/clear` |
