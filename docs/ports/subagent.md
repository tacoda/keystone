# Port: Subagent

**Activation:** Spawned by the host's delegation primitive (e.g. Claude Code's Task tool) by `id`. The file body is the subagent's system prompt.
**Purpose:** Host-native delegated agent definition. `subagent` is the raw agent-native primitive; framework authors usually reach for `persona` (which wraps it) and let `keystone project` write the host counterpart.

## Path convention

```
harness/agents/<id>.md                                        # project-owned
harness/policies/<policy>/agents/<id>.md                       # policy-owned (read-only)
```

Flat — subagents are global by `id` across the cascade.

## Required shape

```markdown
---
kind: subagent
id: <stable id>
description: 'One sentence — what this subagent does.'
tools:
  - <tool name>
---

# Subagent: <name>

<system-prompt body>
```

- **`kind: subagent`** — required.
- **`tools:`** — required allow-list. Subagents without `tools:` fail `keystone verify`.

## Cascade behavior

Project wins; deeper-nested policies refine. `strict.subagents: [<id>]` locks absolutely.

`keystone project` does not re-author subagents — they ARE the host-native form. A `persona` (wrapper) authored in `harness/personas/` *is* projected to `.claude/agents/<id>.md`; authoring a subagent directly bypasses the wrapper layer.

## Example

```markdown
---
kind: subagent
id: drift-reviewer
description: 'Drift reviewer — flags places the current diff drifts from loaded harness rules.'
tools:
  - Read
  - Grep
  - Bash
---

# Subagent: drift-reviewer

You compare the diff against guides under `harness/guides/`. Report only
deviations; do not propose unrelated improvements.
```

## Authoring

Prefer `persona` (framework wrapper) over authoring `subagent` directly. Use `subagent` only when the host's native semantics are needed verbatim.
