---
name: keystone-new-playbook
description: Scaffold a new playbook (ordered sequence of actions) at the canonical path.
tools:
  - Read
  - Write
  - Edit
  - Glob
  - Bash
model: sonnet
---

# keystone:new-playbook — scaffold a playbook

A **playbook** is a named, ordered sequence of actions that drives one
end-to-end workflow. The shipped `task` playbook orchestrates
`spec → orient → implementation → check-drift → verify → review`.
Custom playbooks compose other workflows the same way.

Playbooks live at `.keystone/harness/playbooks/<name>.md`.

## Run

```
keystone new playbook <name>
```

Example:

```
keystone new playbook release
# writes .keystone/harness/playbooks/release.md
```

## After scaffolding

1. Fill in `## Sequence` — a numbered list of action names that run in
   order.
2. Fill in `## Halt conditions` — when the playbook stops early
   (sensor failure, missing prerequisite, user veto).
3. Run `keystone index` to refresh the descriptor surface.

Full port contract:
[`docs/ports/playbook.md`](../../../../docs/ports/playbook.md).
