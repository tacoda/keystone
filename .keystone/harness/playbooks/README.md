# Playbooks

A **playbook** is a markdown file that runs an ordered set of [actions](actions/README.md). Actions are single units of work; playbooks chain them.

This directory holds **project playbooks**. Plugins can also distribute playbooks — they live at `harness/plugins/<name>/playbooks/` when vendored.

## Action vs. playbook

- **Action** — one unit of work (one markdown file). Read [`harness/actions/`](actions).
- **Playbook** — orchestrates multiple actions in order. This directory.

Most files in `harness/actions/` are single-purpose (e.g., `spec`, `verify`, `review`). When a playbook says "run spec, then orient, then verify," it follow-links into those action files.

## Invocation

The agent reads its menu file (`CLAUDE.md`, `AGENTS.md`, etc.) on session start. The menu lists every playbook and action with a one-line description and a link. When the user says "run task" (or "run the task playbook"), the agent follows the link and executes.

## Playbooks in this project

| Playbook | File | What it chains |
|---|---|---|
| **task** | [`task.md`](task.md) | spec → orient → implementation → check-drift → verify → review (+ optional learn) |

## Override cascade

For any `<name>.md`, the project's `harness/playbooks/<name>.md` always wins by default. Among plugins, plugins nested deeper in `keystone.json` refine the outer plugins they're nested in. A plugin can mark an item `strict` to make it absolute — nothing else can override a strict item, not the project, not any other plugin. `keystone verify` reports a violation if any layer attempts to shadow a strict item.

The same cascade applies to **actions** (`harness/actions/`) and **guides** (`harness/guides/`). **Corpus** is background reference loaded on-demand by forward-link from a guide; it doesn't cascade and is never strict-able.
