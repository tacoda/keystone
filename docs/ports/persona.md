# Port: Persona

**Activation:** Spawned by an action via the host's subagent mechanism, addressed by `id`. The system prompt is the file body.
**Purpose:** A delegated subagent authored at the framework layer — sharpened posture plus a tools allow-list. `persona` wraps the agent-native `subagent` kind; the framework primitive is the canonical authoring surface, the host-native one is the escape hatch.

## Path convention

```
harness/personas/<id>.md                                      # project-owned
harness/policies/<policy>/personas/<id>.md                     # policy-owned (read-only)
```

Flat — personas are global by `id` across the cascade.

## Required shape

```markdown
---
kind: persona
id: <stable id>
description: 'One sentence — what this persona does.'
tools:
  - <tool name>
  - <tool name>
---

# Persona: <name>

<system-prompt body — the persona's posture, instructions, examples>
```

- **`kind: persona`** — required.
- **`id:`** — required, stable across renames.
- **`description:`** — one sentence; surfaces in the host's agent picker.
- **`tools:`** — required allow-list. Mirrors the subagent contract: a persona without a `tools:` list is rejected by `keystone verify`.

## Cascade behavior

Project wins by default; among policies, deeper-nested refines outer. A policy may declare `strict.personas: [<id>]` to lock the persona absolutely.

`keystone project` writes the host-native counterpart (e.g. `.claude/agents/<id>.md`) on every run. Re-running `keystone project` after editing a persona body re-projects.

## Example

```markdown
---
kind: persona
id: security-reviewer
description: 'Security-focused PR reviewer that hunts OWASP-top-10 patterns and unsafe defaults.'
tools:
  - Read
  - Grep
  - Glob
---

# Persona: security-reviewer

You are a security-focused code reviewer. Read the diff; flag injection,
auth, secrets, unsafe deserialization, weak crypto, missing authorization.
One finding per line; severity-tagged; do not propose unrelated cleanups.
```

## Authoring

```
keystone new persona <id>
```
