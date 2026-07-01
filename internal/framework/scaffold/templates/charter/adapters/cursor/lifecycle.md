# Cursor — Lifecycle binding

How Cursor runs the charter's lifecycle actions.

## Invocation

Every action is invoked via natural language in the chat: "run task on TICKET-123," "run verify," "do a review pass." Cursor's always-applied rule (`.cursor/rules/keystone.mdc`) lists every action with a pointer to its playbook in `charter/actions/<action>.md`. The agent reads the playbook and executes it. No per-action `.mdc` rule files, no `@`-references — the playbooks are agent-agnostic.

The canonical kickoff phrase is **"run task on `<ticket-id>`"** (or "run the task workflow") — `charter/actions/task.md` orchestrates `spec → orient → implementation → check-drift → verify → review`.

## Why one rule, not many

Cursor's rule system supports three kinds of `.mdc` files: always-applied, auto-attached (by glob), and manual (`@<name>`). The charter uses **one always-applied rule** (`keystone.mdc`) as the menu file; idiom guides loaded by region still use auto-attach via `globs:` frontmatter inside `charter/guides/idioms/<stack>/`. Manual `@`-references are no longer used — every action lives in `charter/actions/` and the menu rule lists them.

## Sensor execution

Cursor's agent mode can invoke shell commands with per-call approval. Sensors run there. In chat-only mode the user pastes sensor output back into the chat. See `charter/adapters/cursor/sensors.md` for the per-sensor binding.

## Sub-agent parallelism

Cursor does not have first-class parallel sub-agents. The **review** action runs review concerns **sequentially** within the same chat — each is its own pass over the diff.

## Modes

Cursor effectively has one autonomy level — the user accepts each tool call in agent mode. The charter's `paired`/`solo`/`autopilot` distinction collapses to **paired** in practice.

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
