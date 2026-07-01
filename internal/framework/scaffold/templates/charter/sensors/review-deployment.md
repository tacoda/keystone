---
kind: sensor
mode: inferential
returns: review-findings
id: review-deployment
description: 'An agent reads the diff and reports deployment considerations — schema migration safety (expand-contract), feature-flag gating, environme...'
---
# Sensor: review-deployment

An agent reads the diff and reports deployment considerations — schema migration safety (expand-contract), feature-flag gating, environment / config drift, backwards compatibility during rolling deploy, the rollback path.

- **Trigger** — **review** (review phase).
- **Inputs** — the diff, the spec, and relevant deployment-sensitive corpus (e.g., `corpus/principles/migrations.md`, `corpus/principles/rollback.md`).
- **Exit condition** — no blocking findings, or the user explicitly accepts the deployment risk.
- **Output** — findings list keyed by file and line, each tagged blocking / nit / note, with the deployment concern referenced (schema, flag, config, compatibility, rollback).
- **State writes** — none.

How the agent is invoked is adapter-specific — sub-agent, separate session, MCP tool, or a checklist prompt — see `charter/adapters/<your-agent>/sensors.md`. The bootstrap action records whether this sensor is available for the project's active adapter.
