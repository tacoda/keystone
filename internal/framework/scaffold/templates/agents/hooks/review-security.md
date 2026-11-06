---
kind: hook
mode: inferential
returns: review-findings
id: review-security
description: 'An agent reads the diff and reports security concerns — injection, auth, secrets, unsafe deserialization, weak crypto, missing authorizat...'
---
# Hook: review-security

An agent reads the diff and reports security concerns — injection, auth, secrets, unsafe deserialization, weak crypto, missing authorization checks, dependency risk.

- **Trigger** — **review** (review phase).
- **Inputs** — the diff, the spec, and any relevant security-sensitive references (e.g., `.agents/references/principles/secrets-management.md`).
- **Exit condition** — no blocking findings, or the user explicitly accepts the residual risk.
- **Output** — findings list keyed by file and line, each tagged blocking / nit / note, with the threat referenced.
- **State writes** — none.

How the agent is invoked is adapter-specific: sub-agent, separate session, MCP tool, or a checklist prompt. See the host's entry in the `AGENTS.md` keystone block. The bootstrap command records whether this hook is available for the project's active adapter.
