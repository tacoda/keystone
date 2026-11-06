---
kind: subagent
id: review-deployment
description: 'An agent reads the diff and reports deployment considerations — schema migration safety (expand-contract), feature-flag gating, environme...'
tags:
  - llm-judgment
tools:
  - Read
  - Grep
---
# Subagent: review-deployment

An agent reads the diff and reports deployment considerations — schema migration safety (expand-contract), feature-flag gating, environment / config drift, backwards compatibility during rolling deploy, the rollback path.

- **Trigger** — **review** (review phase).
- **Inputs** — the diff, the spec, and relevant deployment-sensitive references (e.g., `.agents/references/principles/migrations.md`, `.agents/references/principles/rollback.md`).
- **Exit condition** — no blocking findings, or the user explicitly accepts the deployment risk.
- **Output** — findings list keyed by file and line, each tagged blocking / nit / note, with the deployment concern referenced (schema, flag, config, compatibility, rollback).
- **State writes** — none.

How the agent is invoked is host-specific: a subagent, a separate session, an MCP tool, or a checklist prompt. The bootstrap command records whether this subagent is available on the project's active host.
