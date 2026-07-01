---
kind: sensor
mode: inferential
returns: review-findings
id: review-functional
description: 'An agent reads the diff and reports logic, correctness, and behavior issues.'
---
# Sensor: review-functional

An agent reads the diff and reports logic, correctness, and behavior issues.

- **Trigger** — **review** (review phase).
- **Inputs** — the diff, the spec, and any region-relevant context (idioms, state).
- **Exit condition** — the reviewer reports no blocking issues, or the user accepts the remaining notes.
- **Output** — findings list keyed by file and line, each tagged blocking / nit / note.
- **State writes** — none.

How the agent is invoked is adapter-specific — sub-agent, separate session, MCP tool, or a checklist prompt — see `charter/adapters/<your-agent>/sensors.md`. The bootstrap action records whether this sensor is available for the project's active adapter.
