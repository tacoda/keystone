# Primitives

<!-- voice-register: informational -->
<!-- voice-english: ste -->

A primitive is one typed unit of the charter: one Markdown file under `.agents/`, or one folder for a skill. Each primitive has YAML frontmatter and a body. The frontmatter is the descriptor that keystone indexes. The agent reads the descriptors first and opens a body only when that primitive applies. This page covers what every kind shares. Each kind has its own page for the fields and behavior that only it has.

## The two categories

Every kind belongs to one of two categories. A **constraint** limits, gates, or checks what the agent does. A **capability** adds something the agent can do. `keystone list`, `INDEX.lite.json`, and `AGENTS.md` show constraints first.

| Kind | Category | Folder | Page |
|---|---|---|---|
| `rule` | constraint | `.agents/rules/` | [rule.md](rule.md) |
| `reference` | constraint | `.agents/references/` | [reference.md](reference.md) |
| `hook` | constraint | `.agents/hooks/` | [hook.md](hook.md) |
| `permission` | constraint | `.agents/permissions/` | [permission.md](permission.md) |
| `eval` | constraint | `.agents/evals/<id>/EVAL.md` | [eval.md](eval.md) |
| `skill` | capability | `.agents/skills/<id>/SKILL.md` | [skill.md](skill.md) |
| `command` | capability | `.agents/commands/` | [command.md](command.md) |
| `subagent` | capability | `.agents/agents/` | [subagent.md](subagent.md) |
| `tool` | capability | `.agents/tools/` | [tool.md](tool.md) |
| `document` | capability | `.agents/documents/` | [document.md](document.md) |

Keystone knows only these ten kinds. `keystone lint` reports any other kind as an error, except the retired `concern`, which gets a warning until you inline it with `keystone lint --fix`.

## Kind comes from the folder

Keystone reads the kind from the first folder under `.agents/`. A file in `.agents/hooks/` is a hook, so you can leave `kind:` out. If you write `kind:`, keystone uses it. The same rule applies inside a vendored policy tree under `.keystone/policies/`.

Skills and evals are folders. The indexer reads `SKILL.md` in each skill folder and `EVAL.md` in each eval folder. In the other kind folders, it reads every `.md` file except `README.md`.

## Shared frontmatter

You can set these fields on every kind. The full list is the `Frontmatter` struct in `internal/framework/primitive/primitive.go`. The per-kind JSON Schema is `docs/schemas/primitive.schema.json`, with one `$defs` entry per kind.

```yaml
---
id: process/verification          # stable id; the join key in the index
description: The phase where claims are paired with evidence before commit.
tags: [process]                   # optional, kebab-case
corpus: [process/verification]    # optional: references this primitive cites
supersedes: [process/old-verify]  # optional: ids this primitive replaces
---
```

- **`id`** names the primitive. If you leave it out, `keystone lint` reports an error. Use letters, digits, and `-`, `:`, `_`, `/`, `.`. A spec-shaped skill sets `name:` instead, and keystone uses that value as the id.
- **`description`** must be present. Write one line. The agent reads it in the index to decide whether to open the body. If the agent must read the body to know what the file does, fix the description.
- **`tags`** is an optional list for `keystone list --tag`. Lint requires kebab-case.
- **`corpus`** lists the references that hold the reasoning. The agent opens them on demand.
- **`supersedes`** names the older primitives that this one replaces.

Other fields belong to one kind, and each kind page describes its own fields:

| Kind | Its own fields |
|---|---|
| `rule` | `paths:`, `mode:`, `run:`, `severity:`, `tier:` |
| `hook` | `event:`, `mode:`, `run:`, `agent:`, `returns:` |
| `permission` | `allow:`, `ask:`, `deny:` |
| `skill` | `name:`, `license:`, `compatibility:`, `allowed-tools:`, `metadata:` |
| `command` | `phase:`, `args:`, `tools:`, `model:`, `produces:`, `consumes:`, `stop:` |
| `subagent` | `tools:`, `model:` |
| `tool` | `transport:`, `run:`, `event:` |
| `document` | `type:`, `gates:`, `gate:`, `produced_by:` |

## The index

`keystone index` walks `.agents/` and writes two files under `.keystone/`. The write is atomic, so an agent never reads a partial file.

- **`INDEX.json`** holds the full descriptor of each primitive, plus its `path` and `provenance`. It also has three lookup maps: `by_kind`, `by_category`, and `by_path`. `by_path` maps each `paths:` glob to the primitives it scopes.
- **`INDEX.lite.json`** holds only `category`, `kind`, `id`, and `description` for each primitive. Entries sort by category (constraints first), then kind, then id. Read this file first when you want the whole charter in one view.

The index version is `5.0`. The index holds descriptors only, never bodies. To read a body, the agent opens the file at `path`.

After you add, move, or delete a primitive, run `keystone index`. Or run `keystone watch` in a side terminal, and it re-indexes on every save under `.agents/`.

## Hosts

Hosts read primitives in different ways: some read `.agents/` directly, and some need a linked or translated copy. So `keystone project` writes what each host needs, and [`docs/hosts.md`](../hosts.md) gives the strategy per host and kind. Always edit the source in `.agents/`, never the projected file.

## Override

Project files in `.agents/` override installed policies in `.keystone/policies/`. Among policies, a policy nested deeper in `keystone.json` refines the outer one. To lock an item so that nothing overrides it, list it under `strict:` in the policy.

## 4.x names

Keystone still reads a 4.x tree. It maps the old names when it parses each file, and `keystone lint` warns on each one. `keystone lint --fix` rewrites the keys in place, and `keystone migrate` moves the folders.

| 4.x name | 5.0 name |
|---|---|
| kind `guide`, folder `guides/` | `rule`, `rules/` |
| kind `corpus`, folder `corpus/` | `reference`, `references/` |
| kind `sensor`, folder `sensors/` | `hook`, `hooks/` |
| kind `posture`, folder `posture/` | `permission`, `permissions/` |
| kind `agent` | `subagent` |
| kind `playbook` or `pattern` | `skill` |
| key `globs:` | `paths:` |
| key `on:` | `event:` |

Keystone 5.0 retires the `concern` kind. Lint warns on it, and `keystone lint --fix` inlines each concern into the primitives that include it.

`keystone new` refuses the 4.x kind names and prints the 5.0 name. To add a primitive, run `keystone new <kind> <name>`.
