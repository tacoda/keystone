# Generic — Lifecycle binding

Fallback for any coding agent that reads markdown but lacks a specific adapter.

## Action → invocation

For agents without slash commands, custom-command syntax, or rules-file triggers, every lifecycle action collapses to "the user asks the agent to do `<action>`, the agent reads `harness/process/<phase>.md` and follows it."

| Action | Invocation |
|---|---|
| **spec** | "Start the spec phase for `<task or tracker card>`." |
| **orient** | "Orient yourself for work in `<region>` — read state and load idioms." |
| **check-drift** | "Check the current diff for drift against the corpus." |
| **verify** | "Run the verify action — execute every sensor and report." |
| **review** | "Run the review action — check spec adherence and review findings." |
| **learn** | "Capture the learnings from this work to `harness/learning/inbox/`." |
| **bootstrap** | "Bootstrap the harness — populate idioms and state from the project." |
| **audit** | "Audit the corpus against the codebase." |
| **synthesize** | "Synthesize inbox items into the corpus." |
| **mode** | "Set pacing mode to `<paired\|solo\|autopilot>`." |

The agent is expected to read `harness/README.md`, then `harness/process/<phase>.md` for the active phase, then execute the activities described there using whatever tools it has.

## Capability matrix

The generic adapter assumes the **minimum capability** an agent might have:

| Capability | Assumed? |
|---|---|
| Reads markdown files | ✓ (this is the floor — without this, no adapter helps) |
| Autonomous shell execution | ✗ (degrades; see `sensors.md`) |
| Sub-agent parallelism | ✗ (degrades; **review** runs sequential) |
| Autonomy levels | ✗ (single mode; `paired` is the safe default) |
| Lazy-by-region loading | ✗ (agent reads files on demand from `harness/`) |
| Context-reset primitive | varies — document yours in `activation.md` |

## When to use this adapter

- Your agent is not yet listed in `harness/adapters/`.
- Your agent has limited or no extension surface.
- You're prototyping and don't want to write a full adapter yet.

If you write a full adapter for your agent, please contribute it back to the keystone repo so the next user doesn't have to.
