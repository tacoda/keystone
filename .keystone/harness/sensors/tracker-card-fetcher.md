---
kind: sensor
id: tracker-card-fetcher
description: 'Fetches a tracker card from Jira / Linear / GitHub Issues / Asana via whatever tracker integration the agent has (MCP server, CLI, policy...'
---
# Sensor: tracker-card-fetcher

Fetches a tracker card from Jira / Linear / GitHub Issues / Asana via whatever tracker integration the agent has (MCP server, CLI, policy), if one is referenced.

- **Trigger** — **spec** (when an ID or URL is provided), occasionally **learn** and **review** to surface card metadata in artifacts.
- **Inputs** — a card identifier (e.g., `PROJ-123`, a Linear URL, a GitHub Issue URL); the corresponding tracker integration.
- **Exit condition** — card fetched, or "card not reachable" message produced if the integration is offline or the user lacks access.
- **Output** — title, description, acceptance criteria (if present), labels, links. The agent never edits the card unless the user explicitly asks.
- **State writes** — none. Card metadata lands in `docs/specs/<file>.md` instead.
