# Playbooks

A **playbook** is a markdown file that runs an ordered set of [actions](actions/README.md). Actions are single units of work; playbooks chain them.

This directory holds **project playbooks**. Orgs can also distribute playbooks via [policies](policies/README.md) — they live at `harness/policies/<name>/playbooks/`.

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

For any `<name>.md`, the file that wins at runtime is from the highest-priority tier present, in order: **project → team → org**. A project playbook at `harness/playbooks/<name>.md` overrides the same-basename file in any team or org policy unless the higher tier declared that item `strict`.

The same cascade applies to **actions** (`harness/actions/`) and **guides** (`harness/guides/`). **Corpus** is background reference loaded on-demand by forward-link from a guide; it doesn't cascade and is never strict-able. See [`harness/policies/README.md`](policies/README.md) for the full rule and `keystone policy verify` behavior.
