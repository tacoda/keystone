---
name: review-functional
description: An agent reads the diff and reports logic, correctness, and behavior issues.
---

# Subagent: review-functional

An agent reads the diff and reports logic, correctness, and behavior issues.

- **Trigger** — **review** (review phase).
- **Inputs** — the diff, the spec, and any region-relevant context (idioms, state).
- **Exit condition** — the reviewer reports no blocking issues, or the user accepts the remaining notes.
- **Output** — findings list keyed by file and line, each tagged blocking / nit / note.
- **State writes** — none.

How the agent is invoked is host-specific: a subagent, a separate session, an MCP tool, or a checklist prompt. The bootstrap command records whether this subagent is available on the project's active host.
