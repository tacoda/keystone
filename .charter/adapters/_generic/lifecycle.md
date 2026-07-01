# Generic — Lifecycle binding

Fallback for any coding agent that reads markdown but lacks a specific adapter.

## Invocation

Every action is invoked via natural language: "run task on TICKET-123," "run verify," "do a review pass." The agent reads `AGENTS.md` at session start, finds the action in the bulleted list, follows the link to `.charter/actions/<action>.md`, and executes the playbook. No slash commands, no rule files, no prompt templates required — markdown reading is the floor.

The canonical kickoff phrase is **"run task on `<ticket-id>`"** (or "run the task workflow") — `.charter/actions/task.md` orchestrates `spec → orient → implementation → check-drift → verify → review`.

## Capability matrix

The generic adapter assumes the **minimum capability** an agent might have:

| Capability | Assumed? |
|---|---|
| Reads markdown files | ✓ (this is the floor — without this, no adapter helps) |
| Autonomous shell execution | ✗ (degrades; see `sensors.md`) |
| Sub-agent parallelism | ✗ (degrades; **review** runs sequential) |
| Autonomy levels | ✗ (single mode; `paired` is the safe default) |
| Lazy-by-region loading | ✗ (agent reads files on demand from `.charter/`) |
| Context-reset primitive | varies — document yours in `activation.md` |

## When to use this adapter

- Your agent is not yet listed in `.charter/adapters/`.
- Your agent has limited or no extension surface.
- You're prototyping and don't want to write a full adapter yet.

If you write a full adapter for your agent, please contribute it back to the keystone repo so the next user doesn't have to.
