---
kind: hook
mode: inferential
returns: review-findings
id: review-deployment
description: 'An agent reads the diff and reports deployment considerations — schema migration safety (expand-contract), feature-flag gating, environme...'
---
# Hook: review-deployment

An agent reads the diff and reports deployment considerations — schema migration safety (expand-contract), feature-flag gating, environment / config drift, backwards compatibility during rolling deploy, the rollback path.

- **Trigger** — **review** (review phase).
- **Inputs** — the diff, the spec, and relevant deployment-sensitive references (e.g., `.agents/references/principles/migrations.md`, `.agents/references/principles/rollback.md`).
- **Exit condition** — no blocking findings, or the user explicitly accepts the deployment risk.
- **Output** — findings list keyed by file and line, each tagged blocking / nit / note, with the deployment concern referenced (schema, flag, config, compatibility, rollback).
- **State writes** — none.

How the agent is invoked is adapter-specific: sub-agent, separate session, MCP tool, or a checklist prompt. See the host's entry in the `AGENTS.md` keystone block. The bootstrap command records whether this hook is available for the project's active adapter.
