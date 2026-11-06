# Slash commands

Canonical home for project-authored slash commands. Each command lives
at `commands/<id>.md` and declares the canonical primitive frontmatter —
`kind: command`, `id`, `description`, `args:` — plus a body the host
invokes when the user types `/<id>`.

This is the **source of truth**. `keystone project` regenerates the host
projection at `.claude/commands/<id>.md` by copying the file verbatim.
Hand-edits to the projection are erased on the next run; the drift
sensor reports them.

The same file is served by `keystone-mcp` for hosts that consult MCP
for command discovery. One source, many consumers.

## Authoring

```
keystone new command <id>
```

Scaffolds `commands/<id>.md` with canonical frontmatter. `args:` lists
each parameter the command accepts — name, type, required, description.
Fill in the body to describe what the command does when invoked.

See [`docs/ports/primitive.md`](../../../docs/ports/primitive.md) for the
full descriptor shape.
