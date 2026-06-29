---
name: keystone-review-risk
description: An agent reads the diff and reports risk concerns — blast radius, reversibility, hot-spot regions, fan-out and shared-state coupling, irr...
tools:
  - Read
  - Grep
---
# Sensor: review-risk

An agent reads the diff and reports risk concerns — blast radius, reversibility, hot-spot regions, fan-out and shared-state coupling, irreversible side effects.

- **Trigger** — **review** (review phase).
- **Inputs** — the diff, the spec, `corpus/state/risk-fingerprints.md` (if present), and any region-relevant context.
- **Exit condition** — no blocking findings, or the user explicitly accepts the residual risk.
- **Output** — findings list keyed by file and line, each tagged blocking / nit / note, with the risk dimension referenced (blast radius, reversibility, coupling, side effect).
- **State writes** — none.

How the agent is invoked is adapter-specific — sub-agent, separate session, MCP tool, or a checklist prompt — see `harness/adapters/<your-agent>/sensors.md`. The bootstrap action records whether this sensor is available for the project's active adapter.
