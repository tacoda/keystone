---
name: keystone-new-guide
description: 'Scaffold a new rule guide + paired corpus stub at the canonical path. Triggers: keystone new guide; keystone:new-guide; /keystone:new-guide; add a new guide; scaffold a rule guide; new keystone rule.'
allowed-tools: Read Write Edit Glob Bash
metadata:
  model: sonnet
  tags: scaffold,composition
---

# keystone:new-guide — scaffold a guide + paired corpus

A **guide** is a markdown file under `.agents/rules/<topic>/`
that declares one or more rules (the *what*); a paired **corpus** entry
under `.agents/references/<topic>/` carries the reasoning (the
*why*). The generator emits both with canonical frontmatter,
charter-root-relative cross-links, and the required tier headings.

## Run

```
keystone new guide <topic>/<name>
```

Example:

```
keystone new guide process/release
# writes .agents/rules/process/release.md
#        .agents/references/process/release.md
```

## Topic selection

The topic directory sets the guide's default activation:

| Topic         | Default activation                                  |
| ------------- | --------------------------------------------------- |
| `domain/`     | Ambient on every action.                            |
| `process/`    | On phase entry.                                     |
| `idioms/<stack>/` | Lazy by region — only when the touched region's stack matches. |
| `principles/` | Ambient on every action.                            |
| `computational/` | Editor / LSP / on-save (tool-driven).            |

See [`docs/ports/guide.md`](../../../docs/ports/guide.md) for the
full guide port contract.

## After scaffolding

1. Fill in the rule body — short, declarative bullets under
   `## RULES` (the default tier). Use `## IRON LAW(S)` or
   `## GOLDEN RULE` only when warranted.
2. Fill in the paired corpus body — long-form reasoning, anti-patterns,
   references.
3. Optionally add a `globs:` frontmatter list to narrow activation to
   specific code paths.
4. Run `keystone index` to refresh the descriptor surface.
