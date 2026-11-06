# References

Each file here is a reference: the reasoning behind a rule. It says why the rule exists, what it guards against, and where it comes from. The agent opens a reference only when it needs the why, so a reference costs no context until then.

Nothing loads a reference for you. A rule must list it under `corpus:`, or the agent never finds it.

## Shape

```markdown
---
id: idioms/go/stdlib-first
description: Why the Go code prefers the stdlib over new dependencies.
---

# Stdlib first, reasoning

The reasoning, then `## Anti-patterns` and `## Sources`, then a link back to the rule.
```

Use the same `<topic>/<name>` as the rule it explains, so the pair is easy to find. A reference has no `## Iron law` or `## Rules` sections. Those belong in the rule.

## Topics

| Folder | What it answers |
|---|---|
| [`idioms/`](idioms/README.md) | How does this stack express good practice? |
| [`domain/`](domain/README.md) | What does the product do, and which business rules must hold? |
| [`state/`](state/README.md) | What is true about the codebase right now? |
| `process/` | Why the shipped process rules work the way they do. |

The folder only groups files. It doesn't change when the agent reads them.

## Add one

Run `keystone new reference <topic>/<name>`. `keystone new rule <topic>/<name>` scaffolds a rule and its paired reference together.
