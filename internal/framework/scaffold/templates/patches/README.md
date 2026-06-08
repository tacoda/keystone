# Patches

This directory holds **framework patches** — small, idempotent operations
that bring a project's `keystone.json` and lockfile up to the binary's
current version. At 1.0 they're scoped to config-schema bumps; project
content lives in your git history, not in patches.

`keystone patch` reads the install's recorded `keystone_version` (from
`<harness-root>/keystone.lock.json`, with a fallback to
`<harness-root>/corpus/state/INSTALL_PROFILE.md` for older installs), then
applies every patch in this tree whose version directory is strictly newer.

## Layout

```
patches/
  <version>/
    NNN-<slug>.json
    NNN-<slug>.json
  <version>/
    NNN-<slug>.json
```

- One directory per release that ships patches. Name = the release version,
  dotted-numeric (`0.16.0`, `1.2.3`). Non-numeric names are ignored.
- Files inside a version dir run in lexical order — prefix with `001-`,
  `002-` to control sequence.
- After every file in a version dir is processed (applied or knowingly
  skipped), `keystone patch` bumps `keystone_version` to that dir's name.

## File format

```json
{
  "id": "001-bump-keystone-json-schema",
  "description": "Add the harness_root field to keystone.json",
  "operations": [
    {
      "type": "ensure_section",
      "path": "keystone.json",
      "after_heading": "...",
      "heading": "...",
      "body": "..."
    }
  ]
}
```

The full operation schema lives at `docs/schemas/patch.json.schema.json`.
Operation types: `add_file`, `frontmatter_set`, `ensure_section`,
`replace_block`, `move_file`, `move_dir`, `delete_file`, `delete_dir`.

## Operation idempotency

Every operation reads the target's current state before writing:

| op                | applied when                                    | no-op when                          | conflict when                          |
| ----------------- | ----------------------------------------------- | ----------------------------------- | -------------------------------------- |
| `add_file`        | target is absent                                | target present, matching content    | target present, different content      |
| `frontmatter_set` | frontmatter has no such key                     | key already present (any value)     | target file missing                    |
| `ensure_section`  | `heading` text not anywhere in file             | `heading` already present           | `after_heading` anchor not found       |
| `replace_block`   | `match` found exactly and `replacement` not yet | `replacement` already present       | `match` not found exactly              |

Conflicts are surfaced but never auto-resolved — the user has to merge by
hand, since a conflict means the file has diverged from the patch's
assumption.

## Authoring a patch

1. Land your config-schema change as a normal commit on the keystone repo.
2. Add a file under `patches/<next-version>/NNN-<slug>.json` that produces
   the same outcome via idempotent operations.
3. Use `keystone patch --dry-run` against a fixture install at the previous
   version to confirm the planned changes match the intended diff.
4. Note the patch in `CHANGELOG.md` under the version's `### Patches` section.

## Read by

- `keystone patch` — the runner (see `internal/framework/patch/`).
