# Port: Command

**Activation:** Host slash mechanism — the user types `/<id>` and the host loads the file body as the prompt.
**Purpose:** Host-native slash-command definition. `command` is the agent-native primitive; framework authors usually reach for `action` (which wraps it) and let `keystone project` write the host counterpart.

## Path convention

```
.charter/commands/<id>.md                                    # project-owned
.charter/policies/<policy>/commands/<id>.md                   # policy-owned (read-only)
```

Flat — commands are global by `id`.

## Required shape

```markdown
---
kind: command
id: <stable id, matches the slash invocation>
description: 'One sentence — what this command does.'
---

# /<id>

<command body — what the agent should do when the user types /<id>>
```

- **`kind: command`** — required.
- **`id:`** — required and must equal the slash invocation (no leading slash).

## Cascade behavior

Project wins; deeper-nested policies refine. `strict.commands: [<id>]` locks absolutely.

`keystone project` writes the host-native form (e.g. `.claude/commands/<id>.md`) on every run.

## Example

See `.charter/commands/audit.md` — the slash-command wrapper for the audit action.

## Authoring

Prefer `action` (framework wrapper) over authoring `command` directly. Author a command only when the host-native slash semantics are needed without an underlying action.
