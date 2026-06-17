# Personas

Agent abstraction. A persona is the identity the agent ADOPTS — a
system-prompt overlay that shapes how it presents and what posture
it takes. Different from `subagent` (a delegated agent invoked via
the Task tool); the persona configures the main agent in-place.

## File shape

```markdown
---
kind: persona
id: senior-reviewer
description: Senior staff engineer doing rigorous PR review.
triggers:
  - review as senior
  - rigorous review
---

# Senior reviewer persona

System-prompt overlay the agent reads when the user invokes this
persona. Describe voice, posture, what to prioritize, what to ignore.
```

## Authoring

```
keystone new persona <id>
```

Scaffolds `personas/<id>.md` with canonical frontmatter.

## When to reach for it

- Switching the agent into a different stance for one task
  (security-paranoid, customer-support-empathy, performance-obsessed).
- Capturing a recurring "act like X" prompt as a first-class
  primitive — discoverable in INDEX.json, callable by name.

Do NOT use persona to ship safety / compliance rules — those belong
in `guides/` so they're always loaded, not opt-in. Persona is for
voice + posture, not for non-negotiables.

See [`docs/ports/primitive.md`](../../../docs/ports/primitive.md).
