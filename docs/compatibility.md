# Keystone 1.0 Compatibility Guarantees

What 1.0 promises going forward. Pin a major and trust that:
- Your `keystone.json` layout still parses.
- Your generators still emit conformant output.
- Your CI scripts calling `keystone init`, `keystone install`, `keystone verify`, `keystone doctor` keep working.

This page is the contract that backs those expectations.

## Stable across 1.x

Once you're on 1.x, these surfaces don't break in minor releases. Breaking changes follow the deprecation cycle below and only land in 2.0.

### Configuration formats
- **`keystone.json` schema** — every field documented at [`docs/schemas/keystone.json.schema.json`](schemas/keystone.json.schema.json).
- **`keystone.lock.json` schema** — [`docs/schemas/keystone.lock.json.schema.json`](schemas/keystone.lock.json.schema.json).
- **`keystone-plugin.json` schema** (plugin manifests) — [`docs/schemas/keystone-plugin.json.schema.json`](schemas/keystone-plugin.json.schema.json).
- **`patch.json` schema** (framework patch files) — [`docs/schemas/patch.json.schema.json`](schemas/patch.json.schema.json).

New optional fields can be added; existing fields can't be renamed or have their types changed without a deprecation cycle.

### Port contracts
The ten ports and their on-disk shapes are stable:

| Port | Contract |
|---|---|
| Guide | [`docs/ports/guide.md`](ports/guide.md) |
| Corpus | [`docs/ports/corpus.md`](ports/corpus.md) |
| Sensor | [`docs/ports/sensor.md`](ports/sensor.md) |
| Action | [`docs/ports/action.md`](ports/action.md) |
| Playbook | [`docs/ports/playbook.md`](ports/playbook.md) |
| Adapter (per-agent) | [`docs/ports/adapter.md`](ports/adapter.md) |
| Flywheel sink | [`docs/ports/flywheel-sink.md`](ports/flywheel-sink.md) |
| State ledger | [`docs/ports/state-ledger.md`](ports/state-ledger.md) |
| Patch | [`docs/ports/patch.md`](ports/patch.md) |
| Budget | [`docs/ports/budget.md`](ports/budget.md) |

What "stable" means for a port:
- **Path convention** for the port stays put. Guides at `<harness-root>/guides/<topic>/<name>.md` are forever.
- **Required frontmatter keys** keep their names and meanings. Optional keys can be added.
- **Required sections** (e.g. sensors' `## Command`, `## Interpretation`, `## Remediation`) keep their names.
- **Cascade behavior** is fixed — project beats plugin; pre-order walk of the plugin tree; `strict` blocks downward.

### CLI surface
These commands and their flags don't disappear or change semantics in 1.x:

```
keystone init [<dir>] [--agent <name>] [--reset --i-understand-this-is-destructive]
              [--harness-root <name>] [--starter <packs>]
keystone install [--dir <path>] [--harness-root <name>]
keystone plugin add <shorthand> [--name <n>] [--dir <path>] [--harness-root <name>]
keystone plugin update <name> [@<new-version>] [--dir <path>] [--harness-root <name>]
keystone plugin remove <name> [--dir <path>] [--harness-root <name>]
keystone verify [--dir <path>] [--harness-root <name>]
keystone doctor [--dir <path>] [--harness-root <name>]
                [--paths-only|--plugins-only|--drift-only|--budget] [--fix]
keystone new <port> <name> [--dir <path>] [--harness-root <name>]
keystone target add <agent>[,<agent>...] [--dir <path>] [--harness-root <name>]
keystone patch [<dir>] [--apply|-y] [--dry-run] [--from <version>]
               [--harness-root <name>]
keystone version
keystone help
```

Subcommands of `new` (guide/corpus/sensor/action/playbook/adapter/plugin) and `plugin` (add/update/remove) are part of the contract too.

### Generator output shape
Every `keystone new <port> <name>` produces a skeleton with:
- The right H1 title format for the port.
- The right frontmatter (none for most ports; `kind:` for sensors).
- The right required sections (e.g. `## Entry condition` / `## Activities` / `## Exit condition` for actions).
- Harness-root-relative cross-references — never `../` or `./` segments.

The fine details of skeleton content (placeholder text inside sections) are templates, not contract.

### Conventions table
[`docs/conventions.md`](conventions.md) is the canonical reference. The table's columns (Port / Project path / Plugin path / Activation / Frontmatter / Required shape / Cascade / Strict-able / Generator) stay; rows can be added when new ports land; existing rows don't have their meanings revised silently.

### Project-side paths
- The harness folder name lives in `keystone.json#harness_root` (default `harness`).
- The lockfile is at `<harness-root>/keystone.lock.json`.
- Vendored plugins are at `<harness-root>/plugins/<name>/`, gitignored.
- `keystone.json` is at the project root.

The shape of these paths is part of the contract. The lockfile is permitted to grow new optional fields under the existing schema.

## Free to evolve in 1.x

These surfaces are intentionally not part of the compatibility contract:

- **`internal/framework/` packages** — Go API for keystone's own packages. Treated as internal; external code MUST NOT import them. Refactors land freely.
- **Scaffold template contents** (the markdown skeletons under `internal/framework/scaffold/templates/`) — improvements ship in minor releases. Existing installs aren't auto-updated; `keystone doctor` reports template drift so users can pull updates intentionally.
- **Warning text and log format** — the wording of stderr warnings, info lines, and progress markers. Stable in spirit, not character-for-character.
- **Internal heuristics** — the whitespace-approximate token estimator, the sensor depth limit default, the cache directory layout. Better implementations can land without a deprecation cycle.
- **Patch file contents** — what shipped patches do is up to the framework author. The schema is stable; the contents aren't.

## Deprecation cycle (post-1.0)

To break any stable surface, the framework follows this cycle:

1. **Minor release announces deprecation.** A warning shim continues to honor the old shape, but its use prints a one-line stderr notice naming the new way. The CHANGELOG entry calls it out under `### Deprecations`.
2. **One minor release later, the deprecation is removed.** The old shape stops working; the warning shim is deleted. Removal lands in the next major release.

Example: if `keystone install` ever needed a flag renamed, the schedule would be:
- `v1.4.0` — both old and new flag accepted; old prints a deprecation warning.
- `v2.0.0` — old flag removed; new flag is the only form.

The 0.x → 1.0 cutover is the one-time exception to this cycle, per [ADR 0007](adr/0007-no-backward-compat-at-1.0.md).

## Versioning

SemVer ([ADR 0008](adr/0008-versioning-policy.md)):
- **MAJOR** — breaks a stable surface (only after a deprecation cycle, except 1.0 itself).
- **MINOR** — new commands, new ports, new optional fields, new template content.
- **PATCH** — bug fixes, performance improvements, doc clarifications.

Pre-release tags (`-alpha`, `-beta`, `-rc`) are used during phase rollouts; consumers can pin them but get no stability guarantee.

Plugin authors version their plugins independently. The framework only pins what `keystone.json` declares.

## Plugin compatibility

Plugins follow a parallel contract:
- The `keystone-plugin.json` manifest schema is stable across 1.x.
- The plugin content shape (`guides/<topic>/<name>.md`, etc.) is the same conventions table that governs project content.
- Plugin authors should set `keystone_min` in their manifest to the lowest binary version they support. A future minor will surface a clear error when a consumer runs an older binary than a plugin requires.

## Release cadence

Minor releases ship monthly during active 1.x development; patch releases roll out as fixes accumulate. Major releases (2.0+) are infrequent and only happen after a deprecation cycle on every breaking change.

The cadence is best-effort, not contractual.

## See also

- [`PLAN-1.0.md`](../PLAN-1.0.md) — the 0.x → 1.0 implementation plan.
- [`docs/upgrade-0.x-to-1.0.md`](upgrade-0.x-to-1.0.md) — the consumer-side upgrade narrative.
- [`docs/adr/`](adr/) — architecture decision records.
- [`docs/ports/`](ports/) — per-port contracts.
- [`docs/conventions.md`](conventions.md) — the canonical convention table.
- [`docs/schemas/`](schemas/) — JSON Schemas for every config file.
