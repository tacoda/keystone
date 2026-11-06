# Keystone compatibility

<!-- voice-register: informational -->
<!-- voice-english: ste -->

This page says what 5.0 keeps working from 4.x, what it stops, and how an older install reaches 5.0. The short version: 5.0 still reads every 4.x name in your frontmatter, permanently, but it reads only `.agents/`. So run `keystone migrate up` once to move your files, and then adopt the new names when you're ready. [`upgrade-4.x-to-5.0.md`](upgrade-4.x-to-5.0.md) walks through the upgrade step by step.

## What 5.0 still reads from 4.x

Keystone maps each 4.x name when it parses a file. `keystone lint` warns once per file with the 5.0 spelling, and `keystone lint --fix` rewrites it in place. Nothing here expires.

| 4.x | 5.0 |
|---|---|
| `kind: guide` | `kind: rule` |
| `kind: corpus` | `kind: reference` |
| `kind: sensor` | `kind: hook` |
| `kind: posture` | `kind: permission` |
| `kind: agent` | `kind: subagent` |
| `kind: playbook`, `kind: pattern` | `kind: skill` |
| `globs:` | `paths:` |
| `on:` | `event:` |

The 4.x folder names also still map to a kind: `guides/`, `corpus/`, `sensors/`, `posture/`, `playbooks/`, and `patterns/`. That matters most for a 4.x policy tree under `.keystone/policies/`, which keeps loading without a change.

A few other 4.x shapes still work:

- **`concern`** still parses, but 5.0 retires it, so lint warns until you inline its fields. `keystone lint --fix` copies them into each primitive that included it.
- **Skill `triggers:`** still parses. Agent Skills has no such field, though, so lint warns. Move the triggers into `description`.
- **Budget keys.** The 4.x `budgets` keys (`guides`, `corpus`, `sensors`, `actions`, `playbooks`) map to their kind. See [`internals/budget.md`](internals/budget.md).
- **MCP filter.** `keystone_list_primitives` takes `paths`, and it still accepts `glob`.

## What 5.0 stops

- **`.charter/`.** The 5.0 binary reads `.agents/` and `.keystone/` only. Until you run the migration, it finds no charter.
- **`keystone new <4.x kind>`.** `keystone new guide` and the other 4.x kind names exit with an error that names the 5.0 kind.
- **`keystone init --reset`.** It deleted `.agents/` and wrote it again. Use `keystone init --overwrite` instead, which replaces keystone's own scaffold and host files in place and deletes nothing.
- **Host file names.** 4.x put a `keystone-` prefix on each projected name, but 5.0 uses the primitive's own id. So `.claude/agents/keystone-reviewer.md` becomes `.claude/agents/reviewer.md`. An id that starts with `keystone` keeps its prefix.
- **Unowned host files.** `keystone project` replaces only files that the lockfile says keystone wrote. To take over another file, pass `--overwrite`.
- **`CHARTER.md`.** Keystone no longer writes it. `AGENTS.md` gets a managed block instead, and `CLAUDE.md` gets a block with `@AGENTS.md`.
- **MCP and JSON names.** `keystone_get_corpus` is now `keystone_get_reference`. The `keystone_new_<kind>` tools take the 5.0 kinds, and `keystone show --json` names the field `paths`, not `globs`.

## The migration chain

`keystone migrate` runs versioned migrations in order: 2.0, 2.1, 2.2, 2.3, 2.4, 3.0, 4.0, and 5.0. It records each applied step in `.keystone/lockfile.json`, so an install from any of those versions reaches 5.0 in one run.

```bash
keystone migrate status            # current version, applied, pending
keystone migrate up --dry-run      # print every step, write nothing
keystone migrate up                # apply every pending migration
keystone migrate down [<version>]  # revert, one step by default
```

A migration edits only the files keystone owns: config, the lockfile, the index, host projections, and the folder layout. It never rewrites what you wrote in a primitive. So a frontmatter change reaches you through `keystone lint --fix`, where you see each diff first.

The 5.0 step moves `.charter/` to `.agents/` and `.keystone/`, removes the 4.x entrypoints, and records the 4.x host files as keystone's. It also archives each 4.x scaffold skill you never edited. If two files map to one place, the step stops before it moves anything and names both files. Section 3 of the upgrade guide shows how to clear that conflict.

Install the 5.0 binary first, because a 4.x binary doesn't know the 5.0 migration. Then install it on every machine that runs agents, including CI. The host hook bridges call the `keystone` on your `PATH`, and a 4.x binary rejects the `--host` flag they pass.

## Stable surfaces

These surfaces don't break in a minor release:

- **Config formats.** Each has a JSON Schema in [`docs/schemas/`](schemas/): `keystone.json`, `keystone.lock.json`, `keystone-policy.json`, `patch.json`, and `primitive.schema.json` for frontmatter. A minor release may add an optional field, but it can't rename one or change its type. Known gap: the `keystone.json` schema doesn't list `adapters` or `signals` yet.
- **Kind contracts.** The ten kinds, their folders, and their required frontmatter keys. [`kinds/README.md`](kinds/README.md) lists them.
- **CLI commands and flags** that [`conventions.md`](conventions.md) and the kind pages document.
- **Generator output.** `keystone new` writes the right folder, frontmatter, and headings for each kind. The placeholder text inside a section is template, not contract.

## Free to change

These can change in any release:

- The Go packages under `internal/framework/`. Don't import them from other code.
- Scaffold template contents. An existing install doesn't update by itself, but `keystone doctor` reports template drift.
- Warning text and log format.
- Internal heuristics, such as the word-count token estimate and the cache layout.

## Deprecation cycle

To break a stable surface, keystone follows two steps:

1. A minor release keeps the old shape working and prints a one-line notice that names the new way. The CHANGELOG lists it under `### Deprecations`.
2. The next major release removes the old shape.

5.0 goes further than the cycle needs for names: the 4.x names in frontmatter keep working with no end date. The 0.x to 1.0 cutover was the one exception to the cycle, per [ADR 0007](adr/0007-no-backward-compat-at-1.0.md).

## Versioning

Keystone uses SemVer, per [ADR 0008](adr/0008-versioning-policy.md):

- **Major** breaks a stable surface, and it ships with a migration.
- **Minor** adds commands, kinds, optional fields, or template content.
- **Patch** fixes bugs and docs.

You can pin a pre-release tag (`-alpha`, `-beta`, `-rc`), but it carries no stability promise.

## Policy compatibility

A policy uses the same folders and kinds as project content, and a 4.x policy tree keeps loading through the legacy mapping above. Policy authors version their policies on their own schedule, and keystone pins only what `keystone.json` declares. A policy manifest can set `keystone_min`, but keystone doesn't check it yet.

## See also

- [`upgrade-4.x-to-5.0.md`](upgrade-4.x-to-5.0.md): the upgrade, step by step.
- [`conventions.md`](conventions.md): the convention table.
- [`hosts.md`](hosts.md): how each host gets each kind.
- [`docs/kinds/`](kinds/README.md): one contract per kind.
- [`docs/adr/`](adr/): architecture decisions.
