# Cline / Roo Code — Lifecycle binding

> **Stub.** Fill in the bindings as you wire Cline up.

## Action → invocation

Cline (and the Roo Code fork) reads a "custom instructions" field configured in the VS Code extension settings. There is no rules-file convention; the entire harness pointer goes in that field.

Lifecycle actions are invoked by asking the agent.

| Action | Invocation |
|---|---|
| **spec** | "Start the spec phase for `<task>`." |
| **orient** | "Orient for work in `<region>`." |
| **check-drift** | "Check the diff for drift." |
| **verify** | "Run the verify action." Cline executes shell commands directly. |
| **review** | "Run the review action." Sequential. |
| **learn** | "Capture learnings from this work." |
| **bootstrap** | "Bootstrap the harness." |
| **audit** | "Audit the corpus." |
| **synthesize** | "Synthesize the inbox." |
| **mode** | Edit `harness/process/modes.md` directly. |

## Capability matrix

| Capability | Supported? |
|---|---|
| Autonomous shell execution | ✓ |
| Sub-agent parallelism | ✗ |
| Autonomy levels | partial (auto-approve toggles) |
| Lazy-by-region | ✗ |
| Context-reset primitive | new task |
