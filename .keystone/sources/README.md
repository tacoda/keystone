# Sources

Framework primitive. Declares an external source the runtime
resolution flow's stage 3 can reach when in-charter rules + corpus
aren't enough. Linear, Confluence, local doc folders, generic
HTTPS endpoints.

## File shape

```markdown
---
kind: source
id: eng-docs
description: Org engineering documentation folder.
type: folder
settings:
  path: ./docs
---

# eng-docs

Optional prose describing what's behind this source, who owns it,
auth requirements, etc.
```

Built-in types: `folder`, `url`. Service adapters (`linear`,
`confluence`, …) land as Phase B.

## Cross-cutting note (2.0)

For backward-compat, the runtime ALSO reads
`.keystone/context.json`'s `sources:` array. Where a source appears
in BOTH the primitive form and context.json, the primitive wins.
Migration plan: in 2.1, `.keystone/context.json`'s `sources:` block
is deprecated; source primitives become the only canonical
declaration.

## Authoring

```
keystone new source <id>
```

## Packaging in policies

Sources can ship via policies — an org policy that wires up its
Linear / Confluence as standard stage-3 sources for every consumer
project.

See [`docs/ports/primitive.md`](../../../docs/ports/primitive.md).
