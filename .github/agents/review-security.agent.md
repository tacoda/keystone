---
name: review-security
description: An agent reads the diff and reports security concerns — injection, auth, secrets, unsafe deserialization, weak crypto, missing authorizat...
---

# Subagent: review-security

An agent reads the diff and reports security concerns — injection, auth, secrets, unsafe deserialization, weak crypto, missing authorization checks, dependency risk.

- **Trigger** — **review** (review phase).
- **Inputs** — the diff, the spec, and any relevant security-sensitive references (e.g., `.agents/references/principles/secrets-management.md`).
- **Exit condition** — no blocking findings, or the user explicitly accepts the residual risk.
- **Output** — findings list keyed by file and line, each tagged blocking / nit / note, with the threat referenced.
- **State writes** — none.

How the agent is invoked is host-specific: a subagent, a separate session, an MCP tool, or a checklist prompt. The bootstrap command records whether this subagent is available on the project's active host.
