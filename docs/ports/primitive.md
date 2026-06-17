# Port: Primitive

**Activation:** This port is not a runtime port — it is the *shape* every other port-typed file shares. The agent never "activates a primitive"; it activates a rule, an action, a skill, etc. The primitive shape is what makes those activations cheap.
**Purpose:** One canonical frontmatter schema across every harness file. Lets the agent read a single `harness/INDEX.json` of descriptors and open bodies only on activation, instead of reading every markdown file to discover what is there.

Every other port doc in this directory — `guide.md`, `action.md`, `corpus.md`, `sensor.md`, `adapter.md`, `playbook.md` — refines this shape with its own required fields and activation rules. When those docs and this one disagree on a shared field, **this doc wins**.

## Why a shared shape

`keystone-mcp` already factors harness content into typed primitives (`ContextDoc{kind, id, severity, source, …}`) and exposes a *descriptor* surface (`keystone_list_topics`) separate from body content. The agent reads descriptors; bodies stay cold until needed.

The static harness can give the agent the same affordance without a server: every file declares the same descriptor in frontmatter, the generator emits a single `harness/INDEX.json`, and the menu file (`CLAUDE.md`, `AGENTS.md`, `.cursor/rules/keystone.mdc`) instructs the agent to read the index first and open bodies only on activation.

Same source of truth (markdown). Far less reading per session. Primitives instead of prose wherever the host supports them.

## Canonical frontmatter

Every harness file ships this frontmatter block. Per-kind required and optional fields are listed under [Required by kind](#required-by-kind).

```yaml
---
# Identity (every kind)
kind: rule | action | corpus | skill | subagent | command | sensor
id: <stable-slug>            # globally unique within kind; survives renames
description: <one line>      # surfaced in INDEX.json; agent reads this without opening the body

# Activation — any subset; absent = the kind's default activation
globs: [...]                 # gitignore-style; narrows by code path
phase: bootstrap | orient | implement | verify | learn | synthesize | audit
triggers: [...]              # natural-language cues (skills, slash commands)
tools: [...]                 # subagents only — tool allow-list
args: [...]                  # commands only — declared parameters

# Relationships
traces: [<corpus-id>, ...]   # forward links to reasoning corpus
deps:   [<primitive-id>, ...]# other primitives this one references

# Authority (rules only)
severity: must | should | may   # replaces H2-tier parsing
tier:     iron | golden | regular  # optional human label, derived from severity
---
```

Fields not listed for a kind are ignored by the indexer but valid in the file (forward-compat).

### Identity

- **`kind`** — required. One of the values above. Unknown kinds fail `keystone lint`.
- **`id`** — required. Stable slug, globally unique within `kind`. Use `<topic>/<name>` for rules/corpus to mirror the path; flat slugs for actions/skills/subagents/commands/sensors. The id survives file moves; the indexer uses it as the join key.
- **`description`** — required. One sentence. This is what the agent sees in `INDEX.json` before deciding whether to open the body. If the agent must read the body to know what the file does, the description is wrong.

### Activation

These fields collectively answer "when does the agent open this body?" Absence is meaningful: an omitted field falls back to the kind's default (documented in that kind's port doc).

- **`globs`** — gitignore-style list (`bmatcuk/doublestar/v4` semantics). Repo-relative POSIX. `!`-prefixed entries exclude. Empty list (`globs: []`) is a parse error — omit the key instead. Globs **narrow** activation; they never expand it. See `guide.md` for the touched-files set per action.
- **`phase`** — one of the lifecycle phases. Used by actions to declare when in the workflow they fit; used by sensors/guides to declare when they evaluate.
- **`triggers`** — list of natural-language phrases. Skills and slash commands use this; the host agent (Claude Code's skill mechanism, Cursor's mode triggers) matches user phrasing against the list. Plain strings, no regex.
- **`tools`** — subagent allow-list. Mirrors Claude Code's `.claude/agents/*.md` `tools:` field. The primitive shape simply makes the field declared instead of implicit.
- **`args`** — slash commands. Each entry is `{name, type, required, description}`. The host's command UI surfaces these.

### Relationships

These fields are how the indexer (and the agent) builds a sparse graph across primitives without reading bodies.

- **`traces`** — forward links from a rule to its corpus reasoning. Replaces the prose `For reasoning, see [...]` footer as the *machine-readable* source; the footer stays as the human-readable rendering. Indexer enforces referential integrity.
- **`deps`** — anything else this primitive references: sensors an action runs, subagents it spawns, skills it consults, other rules it depends on. Form: `<kind>/<id>`.

### Authority

Rules only.

- **`severity`** — `must` | `should` | `may`. Replaces the H2-tier parsing (`## IRON LAW(S)` / `## GOLDEN PATH` / `## RULES`). Frontmatter wins; the H2 sections remain as human-readable rendering but the parser no longer reads them. `keystone migrate-frontmatter` lifts the tier into severity for existing 1.0.x installs.
- **`tier`** — optional human label. `iron` ↔ `must`, `golden` ↔ `should`, `regular` ↔ `may`. The indexer ignores it; authors keep it for readability.

## Required by kind

Bodies are unchanged markdown. Only the descriptor surface differs per kind.

| Kind       | Required frontmatter                                                                           | Default activation                                              |
| ---------- | ---------------------------------------------------------------------------------------------- | --------------------------------------------------------------- |
| `rule`     | `kind`, `id`, `description`, `severity`                                                        | Ambient by topic + `globs` filter (see `guide.md`)              |
| `action`   | `kind`, `id`, `description`, `phase`                                                           | Invoked by name                                                 |
| `corpus`   | `kind`, `id`, `description`                                                                    | On-demand via a guide's `traces`                                |
| `skill`    | `kind`, `id`, `description`, `triggers`                                                        | Host-native trigger match (e.g. Claude Code skill auto-load)    |
| `subagent` | `kind`, `id`, `description`, `tools`                                                           | Invoked by name via host's Task/agent mechanism                 |
| `command`  | `kind`, `id`, `description`, `args`                                                            | Host slash-command invocation                                   |
| `sensor`   | `kind`, `id`, `description`, `phase`                                                           | Per-phase, narrowed by `globs`                                  |

## `harness/INDEX.json`

The indexer (`keystone index`) walks every primitive file under `harness/` and `.claude/`, parses frontmatter, and emits a single artifact:

```json
{
  "version": "1.1",
  "generated": "<iso8601>",
  "primitives": [
    {
      "kind": "rule",
      "id": "idioms/rails/migrations",
      "path": "harness/guides/idioms/rails/migrations.md",
      "description": "Migrations are reversible and small.",
      "globs": ["db/migrate/**"],
      "severity": "must",
      "traces": ["corpus/idioms/rails/migrations"]
    },
    {
      "kind": "action",
      "id": "verify",
      "path": "harness/actions/verify.md",
      "description": "Run pre-commit + spawn review agents on current diff.",
      "phase": "verify",
      "deps": ["sensor/lint", "subagent/code-reviewer"]
    },
    {
      "kind": "skill",
      "id": "review-code",
      "path": ".claude/skills/review-code/SKILL.md",
      "description": "Review a PR for logic, style, security.",
      "triggers": ["review this PR", "code review", "/review"]
    }
  ],
  "by_kind": { "rule": ["..."], "action": ["..."] },
  "by_glob": { "db/migrate/**": ["rule/idioms/rails/migrations"] }
}
```

Entries are descriptors only — no body content. `path` is the only body pointer the agent needs. `GLOBS_INDEX.md` (shipped in 1.0.4) becomes a human-readable projection of the `by_glob` map.

## Adapter projections

For hosts that have a native primitive, keystone *emits* the projection from the canonical file so the agent loads through its own machinery:

| Host             | Primitive            | Projection                                                       |
| ---------------- | -------------------- | ---------------------------------------------------------------- |
| Claude Code      | Skill                | `.claude/skills/<id>/SKILL.md`                                   |
| Claude Code      | Subagent             | `.claude/agents/<id>.md`                                         |
| Claude Code      | Slash command        | `.claude/commands/<id>.md`                                       |
| Cursor           | Rule with globs      | `.cursor/rules/keystone-<id>.mdc` (shipped in 1.0.4)             |
| Codex / Aider …  | Pointer              | `AGENTS.md` instructs the agent to read `INDEX.json`             |

MCP prompts are deliberately not a primitive kind — they require an MCP
server plus a client-side UI for the human to pick from, neither of
which the CLI-driven harness can guarantee. Human-triggered templated
workflows belong to slash commands; agent-self-triggered patterns
belong to skills.

Projections are *generated*, never hand-edited. `keystone project` regenerates them from the canonical primitives. When a primitive lives natively in `.claude/` (skills, subagents, commands), the canonical file *is* the host's native file — keystone owns its frontmatter shape so the file is simultaneously a valid Claude Code primitive and a valid INDEX entry. No duplication.

## Cascade behavior

The primitive shape inherits the cascade rules already documented per kind (`guide.md`, `action.md`, `corpus.md`). The shared frontmatter does not change override semantics:

1. Project's file at the canonical path always wins by default.
2. Among plugins, plugins nested deeper in `keystone.json` refine outer plugins.
3. A `strict.<kind>: [<id>]` declaration locks the primitive absolutely.

Exactly one file loads per `(kind, id)`.

## Linting and verification

`keystone lint` enforces:

- Required frontmatter per kind (see table above).
- `kind` is one of the known values.
- `id` is unique within `kind`.
- Every `deps` / `traces` reference resolves to a primitive that exists.
- `globs` parse cleanly (no `globs: []`).
- `severity` and `tier` agree when both are present.

`keystone verify` (existing cascade verifier) additionally requires `harness/INDEX.json` to be newer than every primitive file — stale indexes fail the check, and `bootstrap` / `synthesize` / `learn` actions are responsible for calling `keystone index` on their way out.

## Authoring

```
keystone new <kind> <id>
```

Scaffolds a primitive of the given kind with canonical frontmatter pre-populated. Replaces today's `keystone new guide` / `keystone new corpus` / `keystone new action` with one dispatch (the kind-specific scaffolds remain, dispatched by `<kind>`).

## Out of scope

- **MCP prompts as a primitive kind.** Prompts require an MCP server *and* a client-side UI for a human to pick one; the CLI harness has neither guarantee. Use slash commands (human-triggered) or skills (agent-self-triggered) instead. A keystone-mcp deployment that wants to ship templated prompts can keep them in an MCP-specific location outside the canonical harness contract.
- **Runtime dispatch.** No `keystone serve`, no `keystone get|list|exec`. The CLI authors and maintains the harness; the agent reads the artifacts directly.
