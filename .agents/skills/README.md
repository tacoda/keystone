# Skills

Canonical home for project-authored Claude Code skills. Each skill lives
at `skills/<id>/SKILL.md` and declares the canonical primitive
frontmatter — `kind: skill`, `id`, `description`, `triggers:` — plus a
body the host loads when a trigger phrase matches.

This is the **source of truth**. `keystone project` regenerates the host
projection at `.claude/skills/<id>/SKILL.md` by copying the file
verbatim. Hand-edits to the projection are erased on the next run; the
drift sensor reports them.

The same file is read by `keystone-mcp` as a skill resource — the
canonical frontmatter satisfies both interfaces. One source, many
consumers.

## Authoring

```
keystone new skill <id>
```

Scaffolds `skills/<id>/SKILL.md` with canonical frontmatter. Fill in the
body to teach the agent how to do the thing.

See [`docs/ports/primitive.md`](../../../docs/ports/primitive.md) for the
full descriptor shape.
