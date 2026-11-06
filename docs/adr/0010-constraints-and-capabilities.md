# 0010: Constraints and capabilities, standard vocabulary (5.0)

**Status:** Proposed
**Date:** 2026-11-05

This ADR is the amendment record for the 5.0 charter changes (see
`GOVERNANCE.md`, charter amendments). Plan: `keystone-5.0.0-plan.md`.

## Context

Keystone 4.0 kept its own kind names (guide, sensor, posture, corpus,
playbook, pattern, concern, ...) and its own canonical directory
(`.charter/`). Every host then needed a full projection, and a reader had to
learn a vocabulary no other tool uses.

Since then the ecosystem has settled on shared names and shared locations:

- `AGENTS.md` as the root instruction file.
- The Agent Skills spec (`<name>/SKILL.md`) for skills.
- `.agents/` as a tool-neutral project directory, starting with `.agents/skills/`.
- Hosts converge on the same nouns for the rest: rules, hooks, permissions,
  commands, subagents, MCP tools.

Keystone's names now cost readers a lookup and cost Keystone a projection
layer that the standards make unnecessary for some hosts.

## Decision

**Two categories.** Every primitive is either a **constraint** (it limits,
gates, or verifies what the agent does) or a **capability** (it adds something
the agent can do). The charter is the union of both. The authorship test from
ADR 0009 is unchanged.

**Standard kind names.** 13 kinds become 10:

| Category | 5.0 kind | Replaces |
|---|---|---|
| constraint | `rule` | `guide` |
| constraint | `reference` (keystone extension) | `corpus` |
| constraint | `hook` | `sensor` (`type: command` for computational, `type: agent` or `prompt` for inferential) |
| constraint | `permission` | `posture` |
| constraint | `eval` (keystone extension) | `eval` |
| capability | `skill` | `skill`, `playbook`, `pattern` |
| capability | `command` | `command` |
| capability | `subagent` | `agent` |
| capability | `tool` | `tool` |
| capability | `document` (keystone extension) | `document` |

`concern` is retired. Its tool lists are inlined into the consuming
primitives' `allowed-tools`.

"Agent" now means only the host's main agent, as in `AGENTS.md`. A role you
author and delegate to is a **subagent**. Its directory stays `agents/`,
because that is the path hosts read.

**The `hook` kind returns, with a different meaning.** ADR 0009 retired `hook`
as an event-to-handler middleman. In 5.0 `hook` is the reacting check itself
(the 4.x sensor), named the way every host names it. Signals stay as the
custom-event namespace, documented as "custom hook events".

**Canonical layout.** `AGENTS.md` is the rendered root entrypoint;
`CHARTER.md` is removed. Standard primitives live in `.agents/<kind>/` in
standard formats, so a host that reads `.agents/` natively needs nothing
generated. Keystone-private state (INDEX, lockfile, learning, work, archive,
sources, per-host docs) moves to `.keystone/`, so nothing non-standard lands in
`.agents/`.

**Skills conform to the Agent Skills spec.** Frontmatter uses spec fields
only. `kind` is inferred from path, `id` becomes `name`, `triggers:` folds
into `description`, `tags:` moves into `metadata`.

**Kinds without a cross-host spec** (rule, command, subagent, hook,
permission) use a superset of the Claude Code frontmatter, which the most
hosts already mirror. Each gets a JSON schema under `docs/schemas/`.

Rules use `paths:` (not `globs:`) for scoping, because Claude Code and Cline
read that key and Claude Code silently ignores any other.

**Host reach.** Each (host, kind) pair resolves to `native`, `symlink`,
`shim`, `index`, or `report`, recorded in `docs/hosts.md` with a source per
cell. Symlinks are per entry and lockfile-owned. `linkMode: copy` is the
default on Windows, and for any host whose symlink support is unverified.

Four rules from the Phase 0 research:

- **Reach each host exactly once.** Cursor, Copilot, opencode, and Goose read
  both `.agents/skills/` and `.claude/skills/`; Copilot reads
  `.claude/settings.json` hooks. The projection planner picks the smallest
  set of written paths that covers every enabled host once.
- **Never degrade a constraint silently.** A host that cannot enforce a `deny`
  permission or a blocking hook gets a `report` entry. Where a host has
  blocking hooks but no permission file (Copilot, Pi), `deny` rules are
  enforced through a `preToolUse` hook bridge.
- **Fail closed.** Blocking hooks set the host's fail-closed option (Cursor
  `failClosed: true`). Shims respect each host's evaluation order (opencode is
  last-match-wins, so `deny` patterns are written last).
- **Host-specific keys live only in shims.** The Agent Skills validator
  rejects unknown frontmatter, so keys like `disable-model-invocation` never
  appear in canonical files.

Commands stay a kind. Hosts are folding commands into user-invoked skills
(Claude Code, Cursor, Codex), so a host without a command format gets the
command as a user-invoked skill shim.

## Alternatives considered

- **Keep `.charter/` canonical and also project into `.agents/`.** Smaller
  migration, but the source stays a keystone-only format and every host still
  needs a projection. Rejected.
- **Keep playbook as its own capability kind.** Its `gates:` are human gates
  read from the body today, not enforced by code. If machine-enforced gates
  arrive, a skill `metadata` key can carry them. Rejected for now.
- **Emit nested `AGENTS.md` for every scoped rule.** Glob-to-directory mapping
  loses information. Kept as an opt-in (`nestedAgentsMd`), off by default.

## Consequences

- **Breaking.** On-disk layout, kind names, frontmatter shape, CLI and MCP
  names all change. The `v5_0` migration performs the moves and rewrites, and
  writes `.keystone/migration-5.0-report.md` for the lossy folds
  (`triggers:` into `description`, playbook gates into a body checklist).
- Installed 4.x policies must be republished in 5.0 shape. The resolver
  refuses a 4.x policy with a message naming the fix.
- Second layout move in two majors. Accepted: following the standards removes
  per-host projection for the hosts that read `.agents/` natively.

## Provenance

Proposed by the maintainer (solo stage). Becomes Accepted when Phase 0 of the
plan closes: every host-matrix cell sourced or marked unsupported.
