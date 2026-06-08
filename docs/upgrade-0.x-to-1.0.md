# Upgrading from 0.x to 1.0

1.0 is a clean break ([ADR 0007](adr/0007-no-backward-compat-at-1.0.md)). Old shapes are not shimmed, the 0.x patch chain is gone, and the policy CLI has been replaced. This page is the canonical command sequence for getting an existing 0.x install onto 1.0.

The whole upgrade is **destructive by design** — you re-scaffold the harness from 1.0 templates and manually port anything you customized.

## What's changing

| 0.x | 1.0 |
|---|---|
| `keystone-policy.yaml` manifests | `keystone-plugin.json` manifests |
| `.keystone.lock` (YAML) | `<harness-root>/keystone.lock.json` (JSON) |
| Three-tier cascade (Org / Team / Project) | Nested plugin tree in `keystone.json` (pre-order precedence) |
| Policies installed under `harness/policies/<name>/` (read-write) | Plugins vendored under `<harness-root>/plugins/<name>/` (read-only, gitignored, hash-verified, drift-reset) |
| `keystone policy add|update|verify` | `keystone plugin add|update|remove`; `keystone install`; `keystone verify` |
| `keystone migrate` | `keystone patch` (config-schema bumps only) |
| `--force` overwrites | `--reset --i-understand-this-is-destructive` |
| Universal engineering principles scaffolded by default | Opt-in via `--starter universal-principles` |
| Harness folder hardcoded as `harness/` | Configurable via `--harness-root <name>`, recorded in `keystone.json` |
| Inter-harness links written as `../../corpus/...` | Harness-root-relative (`corpus/...`); `keystone doctor --paths-only --fix` rewrites |

## Before you start

1. You have an existing 0.x install with content under `harness/` you've customized.
2. You can install the new binary (brew/curl/scoop) — see the [README](../README.md#install).
3. `git status` is clean. The upgrade tags your current state for reference; it can't tag dirty trees.

## The upgrade

### Step 1 — Tag your 0.x state

```bash
git add harness/ targets/ migrations/ keystone-policy.yaml || true
git commit -m "snapshot pre-1.0"   # if needed
git tag pre-keystone-1.0
git push origin pre-keystone-1.0
```

This is your only safety net. The tag becomes the reference point for diffing whatever the upgrade overwrites.

### Step 2 — Upgrade the binary

Pick your install path; see the [README](../README.md#install).

```bash
brew upgrade keystone
# or: curl -fsSL https://raw.githubusercontent.com/tacoda/keystone/main/install.sh | sh
keystone version    # confirm v1.0.0 or later
```

### Step 3 — Reset the harness from 1.0 templates

```bash
keystone init --reset --i-understand-this-is-destructive
```

The reset wipes the existing harness folder and rewrites it from the 1.0 templates. It also writes `keystone.json` at the project root if missing and adds `<harness-root>/plugins/` to `.gitignore`.

Optional flags worth knowing:
- `--harness-root <name>` — if you want a folder name other than `harness`. Recorded in `keystone.json`; downstream commands inherit it.
- `--starter universal-principles` — opt back into the universal engineering-principles content. Default install no longer ships these.
- `--starter ...` — see `keystone init --help` for the full list of starter packs.

### Step 4 — Install plugins from `keystone.json`

If you had org/team policies in 0.x, you re-declare them as plugins in `keystone.json` and run install. Use the `tacoda/repo@version` shorthand:

```bash
keystone plugin add tacoda/sample-policies@v1.0.0
keystone plugin add gitlab.com/acme/policies@v2.0.0
keystone install                  # if you edited keystone.json by hand
```

Each plugin is fetched into a content-addressable cache, copied to `<harness-root>/plugins/<name>/`, hashed in `keystone.lock.json`, and marked read-only on POSIX. `keystone verify` runs the hash check; drifted plugins are reset, and you re-run `keystone install` to repopulate.

### Step 5 — Port your customizations from the tag

```bash
git diff pre-keystone-1.0 -- harness/
```

For everything in that diff:

- **Project guides/corpus/sensors/actions/playbooks** you authored — copy them back into the new harness folder at the same paths.
- **Universal engineering content** (SOLID, TDD, BDD, naming, error-handling, etc.) — install with `keystone init --starter universal-principles` and reapply any edits.
- **Organization-specific policies** you used to install via `--policy git+...` — translate to plugins and `keystone plugin add tacoda/repo@v1.0.0`.
- **Stale or migrated files** — `keystone doctor` reports them; safe to ignore for content that has moved upstream.

### Step 6 — Verify

```bash
keystone doctor                              # paths + plugin integrity + template drift
keystone doctor --budget                     # per-port token usage report
keystone verify                              # cascade + plugin drift detection
```

Doctor will flag:
- `../`-style inter-harness links from your ported content. Run `keystone doctor --paths-only --fix` to rewrite automatically.
- Plugin drift (any edits to vendored plugins) — auto-reset, re-run `keystone install`.
- Template drift between your scaffolded defaults and the binary's current templates — informational, decide whether to refresh.

## Common surprises

- **The `policy` command is gone.** Running `keystone policy add ...` prints a one-line migration notice and exits non-zero. Use `keystone plugin add ...`.
- **The `migrate` command is gone.** It's `keystone patch` now. The runner is reserved for config-schema bumps — there are no patches shipped at 1.0.
- **`keystone-policy.yaml` and `.keystone.lock` won't be read.** Both formats are JSON in 1.0 (`keystone-plugin.json` and `<harness-root>/keystone.lock.json`).
- **The `policies/` directory is gone from the harness shape.** Vendored content lives under `<harness-root>/plugins/<name>/` and is gitignored.
- **Universal engineering content is opt-in.** Re-add via `--starter universal-principles` at init time, or `keystone init --reset --i-understand-this-is-destructive --starter universal-principles` to re-scaffold.

## Rolling back

If 1.0 doesn't work for your project, the rollback is straightforward thanks to the tag from Step 1:

```bash
git checkout pre-keystone-1.0 -- harness/
brew install tacoda/tap/keystone@0.13.0    # or your pinned 0.x version
```

Then file an issue describing what failed — the upgrade narrative is the surface we want to harden across the 1.x line.

## See also

- [`PLAN-1.0.md`](../PLAN-1.0.md) — the full 0.x → 1.0 plan, with per-phase deliverables.
- [`docs/compatibility.md`](compatibility.md) — what 1.0 promises going forward.
- [`docs/conventions.md`](conventions.md) — the canonical Rails-style convention table.
- [`docs/ports/`](ports/) — per-port contracts.
