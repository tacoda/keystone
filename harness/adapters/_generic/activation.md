# Generic — Activation

How an unknown agent picks up the corpus.

## The menu

For agents without a specific entry-file convention, the installer drops an `AGENTS.md` at the repo root. This is the de facto cross-agent convention; most modern coding agents will read it on session start.

`AGENTS.md` content is short — a few paragraphs pointing at `harness/README.md` and naming the five layers. The full corpus is read from `harness/` on demand, not autoloaded.

## Where the agent reads from

Every agent that uses the generic adapter is expected to:

1. Read `AGENTS.md` (or whatever the user told it to read) at session start.
2. From there, follow links to `harness/README.md`.
3. Read additional `harness/` files on demand as work progresses.

This is the floor — markdown read on demand. Anything beyond that is an adapter upgrade.

## Lazy-by-region — manual

The generic adapter does not have a mechanism to automatically load `harness/idioms/<stack>/` when editing a particular path. The agent must read the **orient** phase explicitly at task start and load the relevant idioms by hand.

## Context-reset primitive

Document yours here. Common patterns:

- A "clear conversation" button in the UI.
- An explicit `/reset` or `/new` command.
- Starting a new session manually.

After a flywheel write (**synthesize** or **audit**), the user must reset the context themselves so the next turn re-reads the updated corpus.

## Capability gaps

The generic adapter has no opinions about:

- How tracker cards are fetched (probably by the human pasting the card content).
- How sub-agents are spawned (probably they aren't; **review** runs sequentially).
- How autonomy levels work (probably only one mode; default to `paired`).

If your agent has capabilities beyond the floor, write a real adapter and contribute it back.
