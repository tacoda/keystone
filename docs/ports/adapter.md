# Port: Adapter (agent)

**Activation:** Loaded at session start by the coding agent. The adapter wires the charter to a specific agent's menu / activation surface.
**Purpose:** Per-agent bindings. Each supported agent (`claude-code`, `codex`, `cursor`, `aider`, `continue`, `cline`, `goose`, `github-copilot`, `pi`, `_generic`) has its own adapter directory describing how that agent invokes the charter.

## Path convention

```
.charter/adapters/<agent>/lifecycle.md                       # project-owned
.charter/adapters/<agent>/sensors.md
.charter/adapters/<agent>/activation.md
.charter/policies/<policy>/adapters/<agent>/...               # policy-owned (read-only)
```

`<agent>` is the agent's canonical short name. New agents land by creating a new `<agent>/` directory and a generator entry.

## Required shape

An adapter is three paired files inside `<agent>/`:

- **`lifecycle.md`** — how this agent invokes playbooks and actions.
- **`sensors.md`** — how this agent invokes sensors and consumes their output.
- **`activation.md`** — what this agent reads at session start (the "menu").

Each file:

```markdown
# Adapter — <agent> — <lifecycle|sensors|activation>

<agent-specific instructions, in the voice and conventions the agent expects>
```

- **H1 title** — required. Format above.
- **Frontmatter** — none required.
- **Three files** — all three are required for an adapter to load. Missing files fail `keystone verify`.

## Cascade behavior

Same as other ports: project's `.charter/adapters/<agent>/...` always wins by default. Policies can ship adapters too (e.g., an org-specific Claude Code adapter); among policies, policies nested deeper in `keystone.json` refine the outer policies they're nested in.

`strict.adapters: ["<agent>"]` on any policy makes that adapter absolute — nothing else (project or any other policy) can ship a competing adapter for that agent.

## Per-agent menu generation

The framework's `internal/framework/adapters/` generates the on-disk menu file for the active agent (`CLAUDE.md`, `AGENTS.md`, `.cursor/rules/*.mdc`, `CONVENTIONS.md`, etc.) from the resolved adapter content.

Adding a new agent at 1.0 requires: a new adapter directory in templates + a menu-file mapping. By Phase 4 the menu-file mapping becomes itself a port adapter, so a new agent needs only adapter markdown — no Go change (see PLAN §6 open question 1).

## Example

```markdown
# Adapter — claude-code — activation

At session start, Claude Code reads `CLAUDE.md` from the project root.
The generated CLAUDE.md lists every ambient guide, the active playbook,
and the agent's sensor invocations.
```

## Authoring

```
keystone new adapter <agent>
```
