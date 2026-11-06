# Patch (retired)

<!-- voice-register: informational -->
<!-- voice-english: ste -->

Keystone 2.1 retired `keystone patch`, and `keystone migrate` replaced it. The binary keeps `patch` as a hidden command, so your old scripts get a pointer instead of "unknown command". The command takes any arguments, prints the `migrate` subcommands, and exits 0. It changes nothing, so replace each call with `keystone migrate`.

```
⚠ `keystone patch` was retired in 2.1.
  Use `keystone migrate` for forward and backward version transforms:
    keystone migrate up      apply pending migrations
    keystone migrate down    roll back the most recent migration
    keystone migrate status  show applied + pending
```

The stub is `patchCmd` in `cmd/keystone/root.go`.

## What to use instead

To upgrade an install, run `keystone migrate`. Add `--dry-run` first if you want to see every step. The registry is in `internal/framework/migrations/`, one file per version (`v5_0.go` moves a 4.x tree to `.agents/` and `.keystone/`).

```
keystone migrate                    # up to the latest version
keystone migrate up [<version>]
keystone migrate down [<version>]   # default: the previous version
keystone migrate status             # applied and pending
```

Both `up` and `down` take `--dir <path>` and `--dry-run`. The lockfile field `migrations_applied` in `.keystone/lockfile.json` records which migrations ran.

## What a migration may change

The package comment in `migrations.go` sets six iron laws. In short:

1. A migration edits only framework-owned files: `keystone.json`, the lockfile, vendored policy manifests, the index, projections, and the scaffolded layout.
2. It never edits a file you wrote, and that includes its frontmatter. `keystone lint` reports frontmatter that no longer fits the schema.
3. A migration may move or rename files and folders, because a move doesn't change content.
4. A migration prints every step before it runs. `--dry-run` prints and writes nothing.
5. If a migration finds an unexpected state, it stops and names the conflict. It doesn't try to recover.
6. After every `up` runs, every command works with no further steps. Until then, each command warns you about pending migrations and still runs.

Each migration has an `up` and a `down` plan, and each step must be safe to run twice.
