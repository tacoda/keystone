# GitHub Copilot CLI — Lifecycle binding

> **Stub.** Fill in the bindings as you wire GitHub Copilot CLI up.

GitHub Copilot CLI reads project context from `.github/copilot-instructions.md` (the same file VS Code's Copilot reads). Lifecycle actions are invoked by asking the agent.

## Action → invocation

| Action | Invocation |
|---|---|
| **spec** | "Start the spec phase for `<task>`." |
| **orient** | "Orient for work in `<region>`." |
| **check-drift** | "Check the diff for drift." |
| **verify** | "Run the verify action." Copilot CLI runs shell commands directly. |
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
| Autonomy levels | partial (per-command approval prompts) |
| Lazy-by-region | ✗ |
| Context-reset primitive | new session |
| GitHub integration | ✓ (native — `gh` CLI, repo/PR/issue context) |
