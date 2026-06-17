# Rules (agent abstraction)

Canonical home for **agent-side rules** — the host-native flavor of a
directive the agent reads at session start. Cursor calls these "rules"
(`.cursor/rules/*.mdc`); CLAUDE.md / AGENTS.md hold the same kind of
content inline.

In keystone's taxonomy, **agent rules** sit alongside (not instead of)
**framework guides** under `.keystone/harness/guides/`. Two different
abstractions, both first-class:

| Layer       | Lives at                                          | When to use it                                                              |
| ----------- | ------------------------------------------------- | --------------------------------------------------------------------------- |
| Framework   | `.keystone/harness/guides/<topic>/<name>.md`      | The default. Structured directive with H2 tiers (iron / golden / regular), paired corpus reasoning, severity-tagged. |
| Agent       | `.keystone/harness/rules/<id>.md`                 | Plain markdown directive the agent reads as-is. Extends what the host already understands; no tier structure, no paired corpus required. |

If you have a choice, **author a guide**. Reach for `rules/` only when:

- The directive matches a host-native rule slot exactly (e.g. a Cursor
  rule with `globs:` that wants to project 1:1 to `.cursor/rules/`).
- You're porting an existing CLAUDE.md / AGENTS.md block into keystone
  and want to keep the original shape.
- The directive does not warrant the framework's full structure
  (paired corpus, tier discipline).

## File shape

```markdown
---
kind: rule
id: <stable-slug>
description: <one line — surfaced in INDEX.json>
globs:                     # optional; narrow activation to code paths
  - "src/billing/**"
---

# <Title>

One-paragraph framing. Then directives.

- Do this.
- Don't do that.
```

## Authoring

```
keystone new rule <id>
```

Scaffolds `rules/<id>.md` with canonical frontmatter.

See [`docs/ports/primitive.md`](../../../docs/ports/primitive.md) for
the full descriptor shape and [`docs/conventions.md`](../../../docs/conventions.md)
for the two-layer taxonomy.
