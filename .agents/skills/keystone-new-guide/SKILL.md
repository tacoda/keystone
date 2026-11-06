---
kind: skill
id: keystone:new-guide
description: Scaffold a new rule guide + paired corpus stub at the canonical path.
triggers:
  - keystone new guide
  - keystone:new-guide
  - /keystone:new-guide
  - add a new guide
  - scaffold a rule guide
  - new keystone rule
model: sonnet
tools:
  - Read
  - Write
  - Edit
  - Glob
  - Bash
includes:
  - scaffolds-primitive
tags:
  - scaffold
---

# keystone:new-guide — scaffold a guide + paired corpus

A **guide** is a markdown file under `.charter/guides/<topic>/`
that declares one or more rules (the *what*); a paired **corpus** entry
under `.charter/corpus/<topic>/` carries the reasoning (the
*why*). The generator emits both with canonical frontmatter,
charter-root-relative cross-links, and the required tier headings.

## Run

```
keystone new guide <topic>/<name>
```

Example:

```
keystone new guide process/release
# writes .charter/guides/process/release.md
#        .charter/corpus/process/release.md
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

See [`docs/ports/guide.md`](../../../../docs/ports/guide.md) for the
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
