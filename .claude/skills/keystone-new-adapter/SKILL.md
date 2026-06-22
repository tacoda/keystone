---
kind: skill
id: keystone:new-adapter
description: Scaffold the per-agent adapter triple (lifecycle, sensors, activation) for a new host.
triggers:
  - keystone new adapter
  - keystone:new-adapter
  - /keystone:new-adapter
  - add a host adapter
  - scaffold a new agent target
model: sonnet
tools:
  - Read
  - Write
  - Edit
  - Glob
  - Bash
includes:
  - scaffolds-primitive
tags:
  - scaffold
---

# keystone:new-adapter — scaffold a host adapter

An **adapter** is the per-agent bundle that teaches one host how to
load the harness — three files per agent: `lifecycle.md` (how the host
invokes playbooks and actions), `sensors.md` (how it runs sensors), and
`activation.md` (the host's menu-file content).

Shipped adapters: `claude-code`, `codex`, `cursor`, `aider`, `cline`,
`continue`, `goose`, `github-copilot`, `pi`. Add a new one when
supporting a host not on that list.

Adapters live at `.keystone/harness/adapters/<agent>/`.

## Run

```
keystone new adapter <agent>
```

Example:

```
keystone new adapter mistral
# writes .keystone/harness/adapters/mistral/{activation,lifecycle,sensors}.md
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
