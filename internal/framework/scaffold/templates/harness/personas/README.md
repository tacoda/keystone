# Personas

Framework abstraction. A persona is a posture-flavored subagent —
the framework wrapper for the agent-side `subagent` primitive, the
same way `action` wraps `command` and `playbook` wraps `skill`.
Authoring a persona is the encouraged path; raw subagents under
`agents/` remain as the escape hatch.

## File shape

```markdown
---
kind: persona
id: security-reviewer
description: Security-focused PR reviewer that hunts OWASP-top-10 patterns.
tools:
  - Read
  - Grep
---

# Security reviewer

System prompt the persona runs under as a delegated subagent. Posture,
priorities, what to ignore, what it returns.

## Output

Markdown table of findings: file:line, severity, issue, suggested fix.
```

## Authoring

```
keystone new persona <id>
```

Scaffolds `personas/<id>.md` with canonical frontmatter.

## Projection

`keystone project` copies the canonical source verbatim to
`.claude/agents/<id>.md` — the host-native subagent path. The
`agents/` escape hatch projects to the same target, so a persona id
and a subagent id cannot collide; the linter rejects.

## When to reach for it

- Capturing a recurring reviewer posture as a first-class primitive
  (security-reviewer, perf-reviewer, accessibility-reviewer).
- Any case where you'd otherwise hand-author a Claude Code subagent
  with a curated tool allow-list.

Do NOT use persona to ship safety / compliance rules — those belong
in `guides/` so they're always loaded, not opt-in. Persona is for
delegated review/inspection postures.

See [`docs/ports/primitive.md`](../../../docs/ports/primitive.md).
