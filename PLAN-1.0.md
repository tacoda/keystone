# Keystone 1.0 — Harness Framework Plan

*Draft. Not yet accepted.*

Convert Keystone from a "harness installer with org policy plugins" into a **harness framework**: a small, stable runtime that loads everything else — including its own defaults — as plugins. Every shipped phase ships a forward-compatible migration so existing 0.x installs keep working.

---

## 1. What changes (and what does not)

### Stays
- Markdown-only for *content* (guides, corpus, sensors, playbooks, actions). No central service.
- Single Go binary distributed via brew / curl / scoop.
- The abstractions: guides, corpus, sensors, playbooks, actions, policies, adapters, learning, archive.
- Org → Team → Project cascade semantics. `strict`, `required`.
- The lockfile model (pinned per-source SHAs; per-file hashes for drift detection).

### Changes
- The framework (Go runtime + conventions + loader) is cleanly separated from the content (markdown). The binary embeds default content as **plugins it ships**, not as "the layout."
- The universal engineering corpus/guides, the lifecycle actions, the `task` playbook, and the default sensors all become **first-class policy plugins** — same shape as user-installed plugins, loaded by the same engine.
- The plugin cascade is described by a **nested JSON tree**, not a flat Org/Team/Project enum. Arbitrary depth. Order in JSON = deterministic precedence.
- The framework is **closed to modification**, **open to extension**: new conventions land via plugins, not by patching the runtime.
- **JSON everywhere for config.** Every manifest, the lockfile, every migration descriptor, and the consumer-side `keystone.json` is JSON. The current YAML formats (`keystone-policy.yaml`, `migrations/<v>/NNN-*.yaml`, `.keystone.lock`) are translated by the Phase 1 sweep; YAML loaders are removed at 1.0. Markdown is still the format for content; JSON is the format for config.
- **Three-tier migrations.** Three independent sources flow through one runner: **framework migrations** (shipped in the binary), **plugin migrations** (shipped inside each plugin), and **project migrations** (authored by the consumer team for their own evolving conventions). Each tier tracks its own version in the lockfile. Framework migrations become rare once defaults move to plugins; most ongoing churn migrates with the plugin that owns it, which is exactly where the change originated.

### Naming
Keep the binary name **`keystone`**. Frame the project as **"Keystone — the harness framework."** No CLI rename, no import-path churn. Discussion captured in [`docs/adr/0001-naming.md`](docs/adr/) (Phase 0).

---

## 2. Guiding principles

These are the rules every phase must respect.

1. **Framework / client division is physical.** Framework code lives in `internal/framework/`. Default content (markdown the framework happens to ship) lives in `plugins/<name>/` and is loaded through the same pipeline as user plugins. No special case for "shipped" content.
2. **Ports and adapters, in English.** Each abstraction is a *port* — a named extension point the framework defines with a contract written in plain English in `docs/ports/<port>.md`. Concrete files in plugins are *adapters* for that port. Adding a new port requires a minor-version bump and a port-doc PR; adding an adapter requires only dropping a markdown file in the right place.
3. **Convention over configuration.** Drop a file at the conventional path → it loads. The convention table lives at `docs/conventions.md` and is the canonical reference. Generators (`keystone new ...`) scaffold every shape so authors never have to remember a path.
4. **Determinism in the cascade.** For any `<port>/<name>`, exactly one file wins. The winner is the first occurrence in a pre-order walk of the plugin tree declared in `keystone.json`. `strict` from any ancestor blocks descendants from overriding. The resolver is pure.
5. **Forward compatibility on every release.** Every phase that changes layout, schema, or behavior ships a migration under `migrations/<version>/`. `keystone migrate` brings 0.x installs up to head idempotently. Removals are deprecated for one minor before deletion.
6. **No backwards-compat sludge inside the framework.** The migration *layer* absorbs old shapes. The runtime sees one shape.

---

## 3. Architecture sketch

### The ports (the things plugins extend)

| Port | What an adapter looks like | Activation |
|---|---|---|
| **Guide** | `guides/<topic>/<name>.md` — rules | Ambient (always loaded) |
| **Corpus** | `corpus/<topic>/<name>.md` — reasoning | On-demand via guide forward-links |
| **Sensor** | `sensors/<name>.md` — automated check | Invoked inside an action |
| **Action** | `actions/<name>.md` — one lifecycle unit | Invoked by name |
| **Playbook** | `playbooks/<name>.md` — ordered actions | Invoked by name |
| **Adapter (agent)** | `adapters/<agent>/{lifecycle,sensors,activation}.md` | Loaded at session start by the agent |
| **Flywheel sink** | `learning/inbox/`, `archive/` | Written by `learn`/`audit` |

Each port has a one-page contract in `docs/ports/<port>.md`: required frontmatter, required sections, forward-link rules, cascade behavior, examples. The contract is the closed-to-modification surface; the markdown is the open-to-extension surface.

### Repo layout (target 1.0 state)

```
keystone/
├── cmd/keystone/           # CLI entrypoint (main.go relocated)
├── internal/framework/     # framework runtime — closed to modification
│   ├── loader/             # plugin discovery + cascade resolution
│   ├── manifest/           # JSON manifest parsing & validation
│   ├── lockfile/           # SHA pinning (JSON)
│   ├── migrate/            # migration runner (reads JSON from all three tiers)
│   ├── budget/             # context budgeting (Phase 5)
│   ├── adapters/           # per-agent codegen for menu files
│   └── scaffold/           # `keystone new ...` generators
├── plugins/                # default plugins shipped embedded in the binary
│   ├── universal/
│   │   ├── keystone-plugin.json
│   │   ├── corpus/  guides/
│   │   └── migrations/<v>/NNN-*.json        # plugin-owned migrations
│   ├── lifecycle/          # task playbook + spec/orient/verify/review/learn
│   ├── sensors/            # default sensor templates
│   └── adapters/           # per-agent bindings (claude-code, codex, cursor, ...)
├── docs/
│   ├── ports/              # port contracts (the framework's English API)
│   ├── conventions.md      # Rails-like convention table
│   ├── adr/                # architecture decision records
│   ├── migration-0.x-to-1.0.md
│   └── schemas/            # JSON Schemas for every config file
├── migrations/             # framework-only migrations (rare after Phase 2)
│   └── <version>/NNN-*.json
└── examples/               # reference plugins, used in CI as fixtures
```

After `keystone init`, a consumer's repo holds **only client content**:

```
<project>/
├── keystone.json                                           # plugin tree (new in 1.0)
├── keystone.lock.json                                      # SHA + file hash pins
├── harness/
│   ├── guides/  corpus/  sensors/  playbooks/  actions/   # project-owned
│   ├── policies/<plugin>/...                              # installed plugin content
│   ├── adapters/<agent>/...                               # adapter overrides if any
│   ├── learning/ archive/                                  # flywheel sinks
│   └── migrations/<v>/NNN-*.json                          # project-owned migrations (optional)
└── <menu file for the active agent>
```

No framework files ship into the consumer repo. The binary stays separately installable.

### Cascade resolution (Phase 3)

`keystone.json` declares the plugin tree:

```json
{
  "version": "1",
  "plugins": [
    { "name": "universal", "source": "embedded" },
    { "name": "acme-org", "source": "git+https://github.com/acme/keystone-policy.git#v2.0.0",
      "strict": { "guides": ["data-handling"] },
      "children": [
        { "name": "acme-platform", "source": "git+https://github.com/acme/platform-policy.git#v1.4.0",
          "strict": { "sensors": ["rubocop"] },
          "children": [
            { "name": "acme-payments", "source": "git+https://github.com/acme/payments-policy.git#v0.9.0" }
          ]
        }
      ]
    },
    { "name": "lifecycle", "source": "embedded" },
    { "name": "sensors",   "source": "embedded" }
  ]
}
```

- **Order = precedence.** A pre-order walk produces a deterministic priority list. Earlier nodes win against later ones for the same `<port>/<name>`. The project's own `harness/<port>/<name>.md` is always at the front (implicit).
- **Strict locks downward.** A `strict` declaration on a node blocks any *deeper* node (its descendants and any later siblings deeper in the tree) from shipping that item. Lockfile/`policy verify` catch violations at install time.
- **One file loads, the rest don't.** When resolving `<port>/<name>`, only the winning file is read into context. The framework never composes overlapping files for the same name.
- **Tier removal.** "Org / team / project" become advisory metadata on a node, not load-order semantics. Depth in JSON replaces the tier enum. Sensors retain the rule that they can't be shipped above a certain depth — encoded as a per-port `max_depth` in the port contract, not a hardcoded enum.

---

## 4. Phases

Each phase: scope → deliverables → migration → exit criteria → risks.

### Phase 0 — Foundation & decision log (target: 0.14.0)

**Scope.** Settle naming, versioning policy, and 1.0 contract surface before shuffling code. Capture ADRs.

**Deliverables.**
- `docs/adr/0001-naming.md` — keep "Keystone"; reposition as framework.
- `docs/adr/0002-framework-client-boundary.md` — what counts as framework, what counts as client.
- `docs/adr/0003-ports-and-adapters.md` — list the ports; freeze the names.
- `docs/adr/0004-cascade-and-json-config.md` — JSON tree + pre-order precedence + strict semantics.
- `docs/adr/0005-versioning-policy.md` — SemVer with explicit compatibility guarantees post-1.0.
- `docs/ports/` — first draft of port contracts (one page each).
- `README.md` reframe: "Keystone — the harness framework."

**Migration.** None. Phase 0 is docs-only — the on-disk shape doesn't change, so there's nothing for `keystone migrate` to do. The Phase 1 sweep replaces the runner and back-converts the existing 0.7.0–0.13.0 migration files from YAML to JSON.

**Exit criteria.**
- All ADRs reviewed and merged.
- No code moved yet. No content moved yet. Pure intent capture.

**Risks.** Bikeshedding on naming. Time-boxed: 3 days.

---

### Phase 1 — Framework / client boundary + YAML→JSON sweep (target: 0.15.0)

**Scope.** Physically separate framework code from default content **without changing what the binary produces for consumers in terms of content**. Move Go sources under `internal/framework/`, `cmd/keystone/`. Convert every config/manifest/lockfile/migration file in the repo from YAML to JSON, including the migration runner itself. Leave embedded markdown content where it sits until Phase 2.

**Deliverables.**
- `cmd/keystone/main.go` — relocated entrypoint.
- `internal/framework/{loader,manifest,lockfile,migrate,scaffold,adapters}/` — code split out from the root-level files (`policy_*.go`, `migrate.go`, `init.go`, etc.).
- A new `Loader` interface that takes a list of `Plugin{name, root fs.FS}` and resolves the cascade. The existing tier-based resolver becomes one of two backends, both feeding the same interface.
- **JSON-only formats.** All in-tree YAML rewritten:
  - `migrations/0.7.0/...0.13.0/*.yaml` rewritten as `*.json` with identical semantics.
  - `keystone-policy.yaml` (in fixtures + docs) replaced with `keystone-plugin.json`.
  - `.keystone.lock` (YAML) replaced with `keystone.lock.json`.
  - The migration runner reads JSON only. The `gopkg.in/yaml.v3` dependency is dropped from `go.mod`.
  - JSON Schemas under `docs/schemas/` for `keystone-plugin.json`, `keystone.lock.json`, `migration.json`.
- Unit tests on the resolver against in-memory fixtures in `examples/`.
- Behavior parity: `keystone init` produces a byte-identical *content* output to 0.13.0 against the same fixture inputs (only config-file extensions/format differ).

**Migration.** `migrations/0.15.0/001-yaml-to-json.json`:
- `add_file` `keystone.lock.json` synthesized from the existing `.keystone.lock` YAML.
- `delete_file` `.keystone.lock` (after the JSON twin is in place).
- For any plugin installed at 0.13.0 whose source-cached manifest was `keystone-policy.yaml`, write a `keystone-plugin.json` translation alongside; the YAML reader behind a deprecation shim still resolves until 1.0 for plugins fetched from older releases.
- `ensure_section` in `INSTALL_PROFILE.md` recording `framework_layout_version: 1` and `config_format: json`.

**Exit criteria.**
- `keystone init` on a fresh dir produces the same content tree as 0.13.0 with JSON configs.
- `keystone migrate` from 0.13.0 → 0.15.0 leaves the consumer with `keystone.lock.json` (no `.keystone.lock`) and no other content change.
- All policy verify and lockfile tests pass; YAML test fixtures replaced with JSON.
- `go.sum` no longer references `yaml.v3`.

**Risks.** Hidden coupling between root-package files. Mitigation: do the move in one PR per concern (loader first, lockfile second, migrate runner third, YAML deletion last).

---

### Phase 2 — Extract default content into first-class plugins + plugin-level migrations (target: 0.16.0)

**Scope.** Move every piece of "default" markdown into a plugin namespace. The binary still embeds them, but they go through the same loader as user plugins. Introduce **plugin-owned migrations** so that future updates to `lifecycle`, `universal`, etc., migrate consumers via the plugin, not the framework.

**Deliverables.**
- `plugins/universal/` — what's currently under `harness/policies/universal/`, with its own `keystone-plugin.json`.
- `plugins/lifecycle/` — currently `harness/playbooks/task.md`, `harness/actions/*.md`, `harness/guides/process/*.md`. The whole lifecycle (spec → orient → verify → review → learn, plus `bootstrap`, `audit`, `synthesize`, `mode`).
- `plugins/sensors/` — currently `harness/sensors/*.md`, as a default plugin with sensor templates the bootstrap action customizes.
- `plugins/adapters/` — per-agent bindings, one sub-plugin per agent (`claude-code`, `codex`, `cursor`, `aider`, `continue`, `cline`, `goose`, `github-copilot`, `pi`, `_generic`).
- Each plugin has a `keystone-plugin.json` declaring its port adapters, version, and any `strict`/`required` items.
- The loader treats every plugin uniformly. The "embedded" source becomes one transport option alongside `git+`.
- **Plugin migrations.** Each plugin can ship a `migrations/<version>/NNN-*.json` tree. The runner discovers plugin migrations through the resolved tree and applies them scoped to that plugin's installed namespace (`harness/policies/<plugin>/...`). Per-plugin versions are tracked in `keystone.lock.json` under `plugins.<name>.applied_version`; a plugin's `migrate` walks from that version to the manifest's current `version`.
- **Plugin migrations cannot escape the plugin's namespace.** `path` fields in plugin migration operations are validated to live under `harness/policies/<plugin>/`. Same enforcement as the existing content-namespace guard.

**Migration.** `migrations/0.16.0/001-relocate-defaults.json`:
- `move_file` for each `harness/playbooks/task.md` → `harness/policies/lifecycle/playbooks/task.md` (and similarly for actions, process guides, sensors).
- `replace_block` updates in `harness/README.md` and `CLAUDE.md`/agent menus to point at the new locations.
- `ensure_section` in `INSTALL_PROFILE.md` recording the relocation.
- Seeds each embedded plugin's `applied_version` in `keystone.lock.json` to its initial version, so plugin migration runs from 0.16.0 forward have a known starting point.

Consumers who never run `migrate` keep working — the framework's loader resolves the old paths via a deprecation shim that logs a one-time warning naming the migration. Shim is removed at 1.0.

**Exit criteria.**
- Every default markdown file is owned by a plugin.
- `harness/` in a fresh install contains only client-authored directories plus `harness/policies/<plugin>/...` for embedded defaults.
- `keystone migrate` cleanly transforms a 0.13.0 install. Idempotent on re-run.
- A plugin can ship a content change *plus* its own migration in the same release, and consumers pick both up via `keystone migrate` without any framework-binary change.
- Removing `plugins/lifecycle/` from the binary cleanly breaks the lifecycle (proving it's a plugin, not a hardcoded special case).

**Risks.** Doc rot — many internal cross-links point at `harness/actions/...`. Mitigation: a link-check sensor (`harness-debt` extension) reports broken links during the migration's PR.

---

### Phase 3 — Nested plugin tree + project migrations (target: 0.17.0)

**Scope.** Replace flat Org/Team/Project tier semantics with a nested plugin tree declared in `keystone.json` (introduced in Phase 1). Introduce **project-owned migrations** so a team's own evolving conventions can be versioned and reapplied like any other tier.

**Deliverables.**
- `keystone.json` consumer-side schema finalized at `docs/schemas/keystone.json.schema.json`.
- Loader reads `keystone.json` at the project root; falls back to inferring a tree from existing 0.x state (universal + ordered list of policies) when absent.
- Pre-order walk produces the deterministic priority list; `strict` on any node blocks descendants. One winner per `<port>/<name>`.
- `keystone policy verify` upgraded to walk the tree and report violations with path context (`acme-org > acme-platform > acme-payments`).
- `keystone policy add` and `policy update` mutate `keystone.json` rather than the lockfile alone (lockfile still pins SHAs).
- **Project migrations.** A consumer's repo may declare `harness/migrations/<version>/NNN-*.json`. `keystone migrate` runs project migrations after framework and plugin migrations on each pass. A `harness.project_version` field in `keystone.json` declares the current schema for the project's own harness; `keystone.lock.json` records the last-applied project version. Used for: a team renaming an internal idiom folder; bulk-rewriting a guide convention; deleting a deprecated sensor template. The team owns these the same way they own a database migration directory.
- **Single migration runner, three tiers.** One walk of pending work per `keystone migrate`: framework → each plugin (pre-order tree) → project. Each tier's progress recorded independently in the lockfile. A failure in any tier halts the run with a clear "which tier, which file" report.

**Migration.** `migrations/0.17.0/001-introduce-keystone-json.json`:
- `add_file` `keystone.json` synthesized from the existing 0.x state (lockfile + manifests).
- `add_file` `harness/migrations/README.md` documenting the project migration directory and pointing at `docs/migration-authoring.md`.
- `ensure_section` in `harness/policies/README.md` documenting the new tree and pointing at `docs/conventions.md`.
- No content moves. Old layout still works.

**Exit criteria.**
- A 0.13.0 install with two installed policies has, post-migration, an equivalent `keystone.json` and identical resolver output.
- The tier enum is gone from runtime code paths.
- A new fixture under `examples/` exercises a five-deep nested tree with `strict` mid-tree.
- A second fixture demonstrates a project ratcheting forward with two project-owned migrations.
- A third fixture demonstrates a plugin shipping a content change + plugin migration; consumers `keystone migrate` and pick up both with no framework change.

**Risks.** Subtle precedence differences between old tier logic and new pre-order logic. Mitigation: a parity test suite that runs both resolvers on a corpus of fixtures and asserts equal outputs for 0.x-shaped configs.

---

### Phase 4 — Conventions, generators, doctor (target: 0.18.0)

**Scope.** Lock in Rails-like ergonomics. Document conventions exhaustively. Ship scaffolding generators. Add `keystone doctor` to flag convention violations.

**Deliverables.**
- `docs/conventions.md` — the canonical convention table. Path → port → activation.
- Generators:
  - `keystone new plugin <name>` — scaffold a plugin repo.
  - `keystone new guide <topic>/<name>` — scaffold a guide + paired corpus stub.
  - `keystone new sensor <name>` — scaffold a sensor with frontmatter.
  - `keystone new action <name>` — scaffold an action.
  - `keystone new playbook <name>` — scaffold a playbook referencing existing actions.
  - `keystone new adapter <agent>` — scaffold per-agent bindings.
- `keystone doctor` — checks every file under `harness/` against the conventions; reports misplacements, missing frontmatter, broken forward-links.
- A `bridle:`-style autoload model: dropping a file in the conventional path is sufficient — no registry edits required.

**Migration.** `migrations/0.18.0/001-doctor-baseline.json`:
- `add_file` `harness/.keystone-doctor.baseline` recording the result of the first doctor run, so subsequent runs only flag *new* violations until the user opts in to strict mode.

**Exit criteria.**
- Authoring a new sensor / guide / action / playbook / plugin takes one command + filling in the body.
- `doctor` runs in <2s on a fixture project with 200 markdown files.
- No code change required to add a new agent — the adapter generator + a plugin entry suffice.

**Risks.** Generator output drift from hand-authored files. Mitigation: a "regenerate" mode that diffs current scaffolds against canonical templates and reports rot.

---

### Phase 5 — Context budgeting as a first-class feature (target: 0.19.0)

**Scope.** Make context budget a framework concern, not an emergent property. Define per-port budgets, measure load, report, and enforce.

**Deliverables.**
- `internal/framework/budget/` — token estimator (whitespace-tokenizer, approximate) and budget allocator per port.
- `keystone.json` accepts an optional `budgets` block:
  ```json
  "budgets": {
    "guides": { "max_tokens": 6000 },
    "corpus": { "max_tokens_per_load": 4000 },
    "adapters": { "max_tokens": 1000 }
  }
  ```
- A `context-budget` sensor (already stubbed under `harness/guides/process/context-budget.md`) becomes computational: it reports the current ambient load and warns when guides approach the budget.
- `keystone doctor --budget` shows per-port utilization and the top contributors.
- `harness-debt` sensor extended: rules counted against budget pressure.

**Migration.** `migrations/0.19.0/001-budget-defaults.json`:
- `ensure_section` in `INSTALL_PROFILE.md` recording the default budgets.
- `add_file` `harness/corpus/state/context-budget.md` capturing the baseline measurement.

**Exit criteria.**
- A fresh install reports its ambient load on `init`.
- Going over budget produces a doctor warning, not a hard error (warnings can be raised to errors by the project).
- Documented in `docs/ports/budget.md`.

**Risks.** Tokenizer drift versus real model tokenizers. Mitigation: document the heuristic; allow an optional `--tokenizer=tiktoken` mode via an opt-in flag (per the [install-time options](../.claude/projects/-Users-tacoda-tacoda-keystone/memory/feedback_install_time_options.md) preference).

---

### Phase 6 — Hardening, deprecation removal, 1.0 (target: 1.0.0)

**Scope.** Freeze. Remove every deprecation shim that has lived through one full minor cycle. Document compatibility guarantees. Cut the release via tag push.

**Deliverables.**
- All shims gone: legacy YAML manifest reader (the one kept through 0.16→0.19 for plugins fetched from older releases), old `harness/playbooks/task.md` path, tier enum, lockfile-as-source-of-truth fallback.
- `docs/migration-0.x-to-1.0.md` — single-page guide for consumers, with the canonical command sequence (`keystone migrate`).
- `docs/compatibility.md` — what 1.0 promises:
  - **Stable:** port names, port contracts, `keystone.json` schema, lockfile schema, plugin manifest schema, generator output shape.
  - **Free to evolve:** internal/framework/ packages, embedded plugin contents (treated as content, not API), warning/log text.
  - **Deprecation cycle:** one minor with a warning before removal in the next major.
- A full regression suite: golden-file tests for `init`, `migrate`, `policy verify`, `doctor`, generators.
- Release notes summarizing the full 0.13 → 1.0 journey for users who skipped intermediate releases.

**Migration.** `migrations/1.0.0/001-finalize.json`:
- `replace_block` in `harness/README.md`: the "Level 2 / Level 3" language replaced with framework/plugin language.
- `delete_file` for any pre-0.16 fallback files the user never cleaned up (idempotent; missing is fine).
- `ensure_section` recording 1.0 in `INSTALL_PROFILE.md`.

`keystone migrate` from *any* 0.x version to 1.0 is a single command. The runner walks every version dir between the install's recorded version and 1.0.

**Exit criteria.**
- All ADRs accepted.
- All ports documented at `docs/ports/<port>.md`.
- Migration from 0.7.0 (the oldest supported) to 1.0 verified against fixture installs.
- `goreleaser` tag push to `v1.0.0`. (Per [release policy](../.claude/projects/-Users-tacoda-tacoda-keystone/memory/feedback_release_via_tag_push.md) — never `gh release create`.)

**Risks.** Hidden user-authored content that depends on shimmed paths. Mitigation: a one-month pre-1.0 RC with the shims removed, calling for community testing.

---

## 5. Migration policy across phases

`keystone migrate` runs migrations from **three independent sources**, each with its own version tracking in `keystone.lock.json`. One command; three tiers walked in order.

### 5.1 The three tiers

| Tier | Authored by | Lives at | Scope of changes | Version source |
|---|---|---|---|---|
| **Framework** | Keystone maintainers (binary) | `migrations/<v>/NNN-*.json` embedded in the binary | Anything under the framework's contract: `keystone.json` schema, lockfile schema, file relocations *the framework* owns | Binary version |
| **Plugin** | Each plugin's author | `plugins/<name>/migrations/<v>/NNN-*.json` (embedded plugins) or `<git-repo>/migrations/<v>/NNN-*.json` (fetched plugins) | Files inside `harness/policies/<plugin>/...` only — the loader rejects out-of-namespace `path` fields | Plugin manifest `version` |
| **Project** | The consumer team | `harness/migrations/<v>/NNN-*.json` in the consumer repo | Anywhere under `harness/` *except* `harness/policies/<plugin>/...` (that's the plugin's namespace) | `harness.project_version` in `keystone.json` |

A single `keystone migrate` run walks framework migrations first, then each plugin in pre-order tree order, then project migrations last. Each tier's progress is recorded independently in `keystone.lock.json`. A failure in any tier halts the run and reports which tier, which file, which operation.

After Phase 2 most ongoing churn happens at the **plugin** tier — exactly where the change originated. Framework migrations become rare; project migrations show up when a team rewrites its own conventions.

### 5.2 Per-tier rules (apply to every migration file in every tier)

- **Idempotent.** Re-running `keystone migrate` on an already-migrated install is a no-op.
- **Additive first.** Prefer `add_file` and `ensure_section` over `move_file` and `replace_block`. A shim path in the loader covers moves so old installs keep resolving until the user runs `migrate`.
- **One deprecation cycle.** Anything renamed, moved, or replaced lives behind a logging shim for one minor before the file/path is removed. The shim's warning names the migration that fixes it.
- **No data loss.** `move_file` operations preserve content. `delete_file` operations are reserved for files the *same tier* authored and that are no longer used; cross-tier deletes are forbidden — a plugin migration cannot delete project content, a framework migration cannot delete plugin content.
- **`migrate --dry-run` is the contract.** The plan emitted by dry-run is the source of truth for what the migration will do. CI on the keystone repo runs dry-run against fixture installs at every prior version, plus per-plugin fixtures for plugin migrations.

### 5.3 File format

All migration files are JSON. Shape (same operations as today, JSON-serialized):

```json
{
  "id": "002-relocate-task-playbook",
  "description": "Move task playbook into the lifecycle plugin namespace",
  "operations": [
    {
      "type": "move_file",
      "path": "harness/playbooks/task.md",
      "to":   "harness/policies/lifecycle/playbooks/task.md"
    },
    {
      "type": "frontmatter_set",
      "path": "harness/policies/lifecycle/playbooks/task.md",
      "key":  "owner",
      "value": "plugin:lifecycle"
    }
  ]
}
```

A JSON Schema for the migration file format lives at `docs/schemas/migration.json.schema.json` and is the canonical reference for operation types.

### 5.4 Authoring a plugin migration

Plugin authors mirror the framework's own pattern:

1. Land the content change in the plugin repo.
2. Bump the manifest `version`.
3. Add `migrations/<new-version>/NNN-*.json` that takes prior installs forward.
4. Tag and publish the plugin repo. Consumers `keystone policy update <plugin>` followed by `keystone migrate`.

Consumers do not need to know which tier owns which change. `keystone migrate` shows a plan grouped by tier; the consumer reviews and applies.

---

## 6. Open questions (resolve in Phase 0)

1. Do `plugins/adapters/<agent>/` ship as one plugin (with sub-namespaces) or one plugin per agent? Affects whether the user can swap, say, Cursor's adapter for a fork without disabling all others.
2. Do we keep `keystone-policy.yaml` as an alias indefinitely, or remove it at 1.0? Default position: alias through 1.0, remove at 2.0.
3. Should `keystone.json` allow inline overrides (small markdown snippets directly in the config) or only references? Default position: references only, to keep the source of truth in markdown files under cascade rules.
4. Per-agent menu file generation — does it become a port adapter (so a new agent doesn't require Go changes)? Default position: yes, by Phase 4.
5. Naming the embedded "default" plugins: keep `universal` / `lifecycle` / `sensors` / `adapters`, or prefix them (`keystone-universal`)? Default position: unprefixed, since they live under `plugins/` in the framework repo and `harness/policies/` in the consumer.

---

## 7. What to do next

If this plan is accepted in shape, the next concrete step is Phase 0 — draft the five ADRs and the first cut of `docs/ports/`. Everything else is gated on those landing.
