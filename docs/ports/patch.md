# Port: Patch

**Activation:** Invoked explicitly by `keystone patch [<dir>]`. Patches are not loaded ambient — they exist to bring an install forward when the binary upgrades.
**Purpose:** Versioned forward-only changes the framework applies when its expected shape evolves — either to config files (`keystone.json`, lockfile) or to the prose it scaffolded (READMEs, sensor/action playbooks). The rules content the user authored stays in their git.

## Path convention

Inside the framework repo:

```
internal/framework/scaffold/templates/patches/<version>/<NNN>-<slug>.json
```

Embedded into the binary; surfaced to consumers via `keystone patch`. No user-side patches at 1.0 (a future minor may add per-project patches under `<harness-root>/patches/`).

## Required shape

A patch is a JSON file with a list of idempotent operations:

```json
{
  "id": "001-bump-keystone-json-schema",
  "description": "Add the harness_root field to keystone.json",
  "operations": [
    {
      "type": "frontmatter_set",
      "path": "keystone.json",
      "key": "harness_root",
      "value": "harness"
    }
  ]
}
```

The full schema is at [`../schemas/patch.json.schema.json`](../schemas/patch.json.schema.json).

## Operations

| Type              | What it does                                                          | Idempotent when                              |
| ----------------- | --------------------------------------------------------------------- | -------------------------------------------- |
| `add_file`        | Create a file if absent.                                              | Target present with matching content.        |
| `frontmatter_set` | Set a YAML frontmatter key if absent.                                 | Key already present (any value).             |
| `ensure_section`  | Append a heading + body if the heading isn't already in the file.     | Heading already present.                     |
| `replace_block`   | Exact-string swap.                                                    | Replacement already present.                 |
| `move_file`       | Relocate one file. Conflicts on diverged destination content.         | Source missing or destination matches.       |
| `move_dir`        | Relocate every file under a directory tree.                           | Source missing or destination tree matches.  |
| `delete_file`     | Remove one file.                                                      | Target already absent.                       |
| `delete_dir`      | Remove an empty directory.                                            | Target already absent.                       |

Conflicts are surfaced but never auto-resolved — a diverged file means the consumer has customized in a way the patch doesn't know about, and a human has to merge.

## How `keystone patch` walks them

1. Reads the install's recorded `keystone_version` (from `<harness-root>/keystone.lock.json`).
2. Lists every patch whose version directory is strictly greater than the recorded version, sorted by `(version asc, filename asc)`.
3. For each patch: plans every operation, previews changes (or applies them with `--apply`), then bumps `keystone_version` to that patch's version after the whole patch succeeds.
4. Conflicts halt the patch run for that version directory — the user resolves, re-runs.

## Scope at 1.0

Patches at 1.0 cover **config-schema bumps** and **framework-scaffolded scaffold updates** — never the rules content the user authored:

- ✅ Add/rename a field in `keystone.json`.
- ✅ Update the lockfile schema (with a corresponding `Version` bump in `internal/framework/lockfile/`).
- ✅ Move a generated file (`INSTALL_PROFILE.md`) to a new location.
- ✅ Update framework-scaffolded scaffold prose — the READMEs under `harness/`, sensor playbooks (`harness/sensors/*.md`), and action playbooks (`harness/actions/*.md`) that ship from `keystone init`. These describe how the harness works; they are framework prose that happens to live in the user's tree.
- ❌ Edit the rules content the user authored — anything they wrote into a guide body, a corpus entry, a domain invariant, or a custom adapter override. That's in their git; they own it.

The bright line is **scaffold prose vs. user rules**: the framework owns the prose it scaffolded; the user owns the rules they wrote. `replace_block` failures (exact-match mismatch) surface user customizations as conflicts so they never get clobbered silently.

The 0.x notion of patches as wholesale content-rewriting "migrations" is still dead. The runner supports content-write operations; their use post-1.0 is bounded by the scaffold/rules line above.

## Authoring a patch

1. Land your config-schema change in the framework code.
2. Add a patch under `templates/patches/<new-version>/<NNN>-<slug>.json` that produces the same outcome via idempotent operations.
3. Verify with `keystone patch --dry-run` against a fixture install at the previous version.
4. Note the patch in `CHANGELOG.md` under the version's `### Patches` section.

There is no `keystone new patch` generator — patches are infrequent, framework-author authored, and best written by hand against the schema.

## CLI

```
keystone patch [<dir>] [--apply|-y] [--dry-run] [--from <version>] [--harness-root <name>]
```

- `--apply` / `-y` — apply every non-conflict change without prompting.
- `--dry-run` — preview every change; write nothing.
- `--from <version>` — override the recorded `keystone_version` (escape hatch for fixture installs and bisecting).

`keystone migrate` (the 0.x name) prints a one-line rename notice and exits non-zero. Anyone with that command in CI scripts gets a clear pointer to the new name.

## Read by

- `keystone patch` — the runner (see `internal/framework/patch/`).
