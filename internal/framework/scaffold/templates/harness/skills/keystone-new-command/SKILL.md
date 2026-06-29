---
kind: skill
id: keystone:new-command
description: Scaffold a new lifecycle action (atomic unit of work) at the canonical path.
triggers:
  - keystone new command
  - keystone:new-command
  - /keystone:new-command
  - add a new command
  - scaffold a lifecycle action
---

# keystone:new-command — scaffold a command

An **action** is one atomic unit of lifecycle work, invoked by name
(by a playbook, another action, the menu file, or the user). Default
actions ship with every install: `spec`, `orient`, `verify`, `review`,
`learn`, `audit`, `bootstrap`, `mode`, `synthesize`, `check-drift`.
Custom actions for project-specific lifecycle phases live alongside.

Actions live at `.keystone/harness/actions/<name>.md` (flat, no topic
directory — actions are global by name across the cascade).

## Run

```
keystone new command <name>
```

Example:

```
keystone new command release-notes
# writes .keystone/harness/actions/release-notes.md
```

## After scaffolding

1. Fill in `## Entry condition` — what must be true before this action
   runs.
2. Fill in `## Activities` — a numbered list of verb + artifact steps.
3. Fill in `## Exit condition` — what must be true when the action
   completes.
4. List dependent sensors / subagents / corpus in the `deps:`
   frontmatter so the indexer can build the graph.
5. Run `keystone index` to refresh the descriptor surface.

Full port contract:
[`docs/ports/action.md`](../../../../docs/ports/action.md).
