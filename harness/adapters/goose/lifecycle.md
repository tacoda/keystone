# Goose — Lifecycle binding

> **Stub.** Fill in the bindings as you wire Goose up.

## Action → invocation

Block's Goose reads `.goosehints` at the repo root (and global config). Lifecycle actions are invoked by asking the agent.

| Action | Invocation |
|---|---|
| **spec** | "Start the spec phase for `<task>`." |
| **orient** | "Orient for work in `<region>`." |
| **check-drift** | "Check the diff for drift." |
| **verify** | "Run the verify action." Goose runs shell commands via its developer toolkit extension. |
| **review** | "Run the review action." Sequential. |
| **learn** | "Capture learnings from this work." |
| **bootstrap** | "Bootstrap the harness." |
| **audit** | "Audit the corpus." |
| **synthesize** | "Synthesize the inbox." |
| **mode** | Edit `harness/guides/process/modes.md` directly. |

## Capability matrix

| Capability | Supported? |
|---|---|
| Autonomous shell execution | ✓ (developer extension) |
| Sub-agent parallelism | ✗ |
| Autonomy levels | ✗ |
| Lazy-by-region | ✗ |
| Context-reset primitive | new session |
