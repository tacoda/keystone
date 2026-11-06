# Subagents

Each file here is a subagent: a role the main agent delegates to, such as a reviewer or a planner. The main agent hands it a narrow job and gets a short report back. Treat that report as a claim to verify, not as evidence.

## Shape

```markdown
---
id: drift-reviewer
description: Flags places the current diff drifts from loaded charter rules.
tools: [Read, Grep, Bash]
model: sonnet          # optional
---

# drift-reviewer

The system prompt. Write it as instructions to the subagent.
```

- `id` is the name the agent spawns it by.
- `tools` is the allow-list. `keystone lint` reports an error when it is missing or empty. Only some hosts enforce it.
- The body is the system prompt.

## Add one

Run `keystone new subagent <id>`, fill in the body, then run `keystone index && keystone project`. The project step writes each host's copy, such as `.claude/agents/<id>.md`. Edit this file, never the copy.
