# Port: Rule

**Activation:** Host-native rule directive. Loaded ambient by the host (e.g. Cursor's `.cursor/rules/*.mdc`, the directive block of `CLAUDE.md`).
**Purpose:** Plain directive surface for hosts whose rules layer expects a flat list of "do this / don't do that." `rule` is the agent-native primitive; framework authors usually reach for `guide` (which wraps it) and let the cascade project a per-host rule file.

## Path convention

```
harness/rules/<id>.md                                         # project-owned
harness/policies/<policy>/rules/<id>.md                        # policy-owned (read-only)
```

Flat — rules are global by `id`.

## Required shape

```markdown
---
kind: rule
id: <stable id>
description: 'One sentence — what this rule constrains.'
globs:                # optional, narrow-only
  - "<pattern>"
---

# <Rule name>

<directive body — short, imperative, no reasoning>
```

- **`kind: rule`** — required.
- **`globs:`** — optional. Narrows ambient activation to a fileset. Cannot broaden.

## Cascade behavior

Project wins; deeper-nested policies refine. `strict.rules: [<id>]` locks absolutely.

For pointer-style hosts (Claude Code, Codex, Aider, etc.), the cascade uses [`corpus/state/GLOBS_INDEX.md`](../../.keystone/harness/corpus/state/GLOBS_INDEX.md) to gate idiom loading. For Cursor, `keystone project` writes a `.cursor/rules/<id>.mdc` per rule with `globs:` mirrored verbatim.

## Authoring

Prefer `guide` (framework wrapper) — it carries reasoning via a paired corpus entry and the cascade still projects to the rules surface. Author a `rule` directly only for short, freestanding host-native directives that need no corpus pairing.
