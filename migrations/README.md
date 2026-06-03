# Migrations

This directory holds **harness migrations** — small, idempotent patch files
that bring an existing harness install up to the binary's version. They are
the file-system analog of database schema migrations: each one declares a
small set of changes that move the harness forward by one release.

`keystone migrate` reads `harness/corpus/state/INSTALL_PROFILE.md` to find the
current `keystone_version`, then applies every migration in this tree whose
version directory is strictly newer.

## Layout

```
migrations/
  <version>/
    NNN-<slug>.yaml
    NNN-<slug>.yaml
  <version>/
    NNN-<slug>.yaml
```

- One directory per release that ships migrations. Name = the release version,
  dotted-numeric (`0.6.0`, `1.2.3`). Non-numeric names are ignored.
- Files inside a version dir run in lexical order — prefix with `001-`, `002-`
  to control sequence.
- After every file in a version dir is processed (applied or knowingly
  skipped), `keystone migrate` bumps `keystone_version` to that dir's name.

## File format

```yaml
id: 001-add-kind-frontmatter            # informational; defaults to the filename
description: Add kind: inferential to existing sensor files

operations:
  - type: add_file                      # create a new file (skip if it already exists)
    path: harness/sensors/review-functional.md
    content: |
      ---
      kind: inferential
      ---
      # Review – Functional concerns
      ...

  - type: frontmatter_set               # set a YAML frontmatter key if absent
    path: harness/sensors/drift.md
    key: kind
    value: inferential

  - type: ensure_section                # append a heading+body if heading not present
    path: harness/corpus/state/CODEBASE_STATE.md
    after_heading: "## Frameworks & libraries"
    heading: "## Sensors inventory"
    body: |
      <markdown content for the new section>

  - type: replace_block                 # exact-string swap; conflict if match not found
    path: harness/README.md
    match: |
      <exact prior text>
    replacement: |
      <new text>
```

## Operation idempotency

Every operation reads the target's current state before writing:

| op                | applied when                                       | no-op when                                  | conflict when                          |
| ----------------- | -------------------------------------------------- | ------------------------------------------- | -------------------------------------- |
| `add_file`        | target is absent                                   | target present with matching content        | target present with different content  |
| `frontmatter_set` | frontmatter has no such key                        | key already present (any value)             | target file missing                    |
| `ensure_section`  | `heading` text not anywhere in file                | `heading` already present                   | `after_heading` anchor not found       |
| `replace_block`   | `match` found exactly and `replacement` not yet    | `replacement` already present               | `match` not found exactly              |

Conflicts are surfaced but never auto-resolved — the user has to merge by hand,
since a conflict means the file has diverged from the migration's assumption.

## Authoring a migration

1. Land your harness change as a normal commit on the keystone repo.
2. Add a file under `migrations/<next-version>/NNN-<slug>.yaml` that produces
   the same outcome via idempotent operations.
3. Use `keystone migrate --dry-run` against a fixture install at the previous
   version to confirm the planned changes match the intended diff.
4. Note the migration in `CHANGELOG.md` under the version's `### Migration` section.

## Read by

- `keystone migrate` — the runner (see `migrate.go`, `migration.go`,
  `migration_ops.go`).
