# Port: Skill

**Activation:** Host-native auto-activation. The host matches the user's input against the skill's `triggers:` and loads its body.
**Purpose:** Surface a workflow as a host-native skill so the agent can invoke it without the user typing a slash command. `skill` is the agent-native primitive; framework authors usually reach for `playbook` (which wraps it) instead.

## Path convention

```
harness/skills/<slug>/SKILL.md                                # project-owned
harness/policies/<policy>/skills/<slug>/SKILL.md               # policy-owned (read-only)
```

A skill is a folder (`<slug>/`) holding `SKILL.md`. The slug becomes the host-native skill id.

## Required shape

```markdown
---
kind: skill
id: <stable id, conventionally `<namespace>:<name>`>
description: 'One sentence — what this skill does.'
triggers:
  - <phrase the user might type>
  - <slash command form>
---

# <Skill Name>

<body — what the agent should do when this skill activates>
```

- **`kind: skill`** — required.
- **`id:`** — required, conventionally namespaced (e.g. `keystone:audit`).
- **`triggers:`** — required, non-empty list. The host substring-matches user input against this set.

## Cascade behavior

Project wins by default; among policies, deeper-nested refines outer. `strict.skills: [<id>]` locks absolutely.

`keystone project` projects the skill into the host's native skills directory (e.g. `.claude/skills/<slug>/SKILL.md`).

## Example

See `harness/skills/keystone-bootstrap/SKILL.md`.

## Authoring

Skills are usually authored as `playbook` (the framework wrapper) and projected. Author a skill directly only when the host-native semantics are needed and no playbook fits.
