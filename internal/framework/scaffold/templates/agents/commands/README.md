# Commands

Keystone ships no commands: its lifecycle steps (spec, orient, verify, review, …) are `keystone-*` skills. Each file here is a command: a unit of work the user asks for by name, such as `/release-notes`. A command does nothing until someone calls it. If the work should start from the task itself, write a skill instead.

## Shape

```markdown
---
id: release-notes
description: Draft release notes from the commits since the last tag.
args:
  - name: card
    type: string
    required: false
    description: The tracker card id.
---

# release-notes

The prompt the host gives the agent when the user calls the command.
```

- `id` is the name the user calls.
- `args` lists each parameter: `name`, `type`, `required`, and `description`.
- Commands are flat: one file per command, no topic folders.

## Add one

Run `keystone new command <id>`, fill in the body, then run `keystone index && keystone project`. Where the host has slash commands, the project step writes one, such as `.claude/commands/<id>.md`. Elsewhere, the command is listed in `AGENTS.md`. Edit this file, never the copy.
