---
name: keystone-new-adapter
description: Scaffold the per-agent adapter triple (lifecycle, sensors, activation) for a new host.
tools:
  - Read
  - Write
  - Edit
  - Glob
  - Bash
model: sonnet
---

# keystone:new-adapter — scaffold a host adapter

An **adapter** is the per-agent bundle that teaches one host how to
load the charter — three files per agent: `lifecycle.md` (how the host
invokes playbooks and actions), `sensors.md` (how it runs sensors), and
`activation.md` (the host's menu-file content).

Shipped adapters: `claude-code`, `codex`, `cursor`, `aider`, `cline`,
`continue`, `goose`, `github-copilot`, `pi`. Add a new one when
supporting a host not on that list.

Adapters live at `.charter/adapters/<agent>/`.

## Run

```
keystone new adapter <agent>
```

Example:

```
keystone new adapter mistral
# writes .charter/adapters/mistral/{activation,lifecycle,sensors}.md
```

## After scaffolding

1. Fill in `lifecycle.md` — how this host invokes the lifecycle
   (natural-language actions, slash commands, MCP tools, etc.).
2. Fill in `sensors.md` — sensor invocation conventions for this host.
3. Fill in `activation.md` — the host-native menu-file content the
   installer will project to the consumer's repo root.
4. Run `keystone index` to refresh the descriptor surface.

Full port contract:
[`docs/ports/adapter.md`](../../../../docs/ports/adapter.md).
