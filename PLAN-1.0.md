# Keystone 1.0 — Harness Framework Plan

**Status:** Approved · Phases 0–3 complete · Phase 4 next
**Last updated:** 2026-06-08

Convert Keystone from a "harness installer with org policy plugins" into a **harness framework**: a small, stable Go runtime with Rails-style conventions for project content, and a vendored, read-only plugin system for sharing policy across projects. 1.0 is a clean break from 0.x — no backward-compatibility shims.

## Progress

| Phase | Target | Status |
|---|---|---|
| 0 — Foundation & decision log | 0.14.0 | ✅ Complete (2026-06-08, commit `6b6295f`) |
| 1 — Framework/client boundary + YAML→JSON sweep | 0.15.0 | ✅ Complete (2026-06-08, 6 sub-commits + 1 fixup) |
| 2 — Convention-scaffolded defaults | 0.16.0 | ✅ Complete (2026-06-08, 5 sub-commits) |
| 3 — Vendored read-only plugins | 0.17.0 | ✅ Complete (2026-06-08, 8 sub-commits) |
| 4 — Conventions, generators, doctor | 0.18.0 | ⏳ Pending |
| 5 — Context budgeting | 0.19.0 | ⏳ Pending |
| 6 — Hardening, upgrade guide, 1.0 | 1.0.0 | ⏳ Pending |

---

## 1. What changes (and what does not)

### Stays
- Markdown-only for *content* (guides, corpus, sensors, playbooks, actions). No central service.
- Single Go binary distributed via brew / curl / scoop.
- The abstractions: guides, corpus, sensors, playbooks, actions, adapters, learning, archive.
- The cascade idea (project wins over plugin; `strict`, `required` semantics).
- The lockfile model (pinned per-source SHAs; per-file hashes for drift detection).

### Changes
- **Built-in features are not plugins.** Defaults (universal guides/corpus, lifecycle playbook and actions, default sensors, per-agent adapters) are scaffolded into the consumer's `harness/<port>/` at conventional paths on `keystone init`. The user owns those files in git and edits them like any other markdown. "Editing the harness" = editing markdown at a known path.
- **Plugins exist only to share policies across projects.** A plugin is an external, read-only dependency — never a vehicle for the binary's own defaults.
- **Plugins are vendored** at `harness/plugins/<name>/`, gitignored, re-fetched on `keystone install` from `keystone.json` + `keystone.lock.json`. The model mirrors `node_modules`.
- **Plugins are read-only and integrity-checked.** The lockfile pins per-file hashes; before each cascade resolution the runtime walks `harness/plugins/<name>/` and compares hashes to the lockfile. **Any drift — extra file, missing file, modified content — triggers `rm -rf` on the plugin's directory and a fresh reinstall from the pinned tag.** Users cannot edit plugin files; edits are silently reverted on the next run. Filesystem-level read-only marking (chmod 0444 on POSIX) is best-effort UX; the hash check is the real enforcement.
- **The only folders that may be actively updated are project folders.** `harness/guides/`, `harness/corpus/`, `harness/sensors/`, `harness/playbooks/`, `harness/actions/`, `harness/adapters/`, `harness/learning/`, `harness/archive/`. All other policy layers come in via plugin and are read-only.
- **The plugin cascade is a nested JSON tree** in `keystone.json` — arbitrary depth, order = deterministic precedence. The Org / Team / Project tier enum is gone.
- **JSON everywhere for config.** `keystone.json`, `keystone.lock.json`, plugin manifests, framework migrations. Markdown stays the format for content. The current YAML (`keystone-policy.yaml`, `.keystone.lock`) is removed at 1.0; YAML loaders are dropped.
- **Migrations are framework-only.** Project content is git-tracked — the user is the source of truth and uses git tags for versioning, not a migration runner. Plugins are atomic at their pinned tag — updating a plugin is a re-vendor, never a migration. A framework migration tier survives only for binary-side concerns (e.g., `keystone.json` schema bumps). It never touches user-authored content.
- **No backward compatibility with 0.x.** 1.0 is a clean break. The 0.x → 1.0 jump is a one-time `keystone init --reset` in the consumer repo, documented in an upgrade guide, performed by the user.

### Naming
Keep the binary name **`keystone`**. Frame the project as **"Keystone - the agent harness framework for any project."** No CLI rename, no import-path churn. Captured in `docs/adr/0001-naming.md` (Phase 0).

---

## 2. Guiding principles

These rules every phase must respect.

1. **Framework / client division is physical.** Framework code lives in `internal/framework/`. Default content lives as templates in `internal/framework/scaffold/templates/` and is *copied* into the consumer's `harness/<port>/` on `keystone init`. There is no "embedded plugin" intermediate concept and no `plugins/` directory inside the framework repo.
2. **Ports and adapters, in English.** Each abstraction is a *port* — a named extension point the framework defines with a contract in `docs/ports/<port>.md`. Concrete files at `harness/<port>/...` (project) or `harness/plugins/<plugin>/<port>/...` (plugin) are *adapters*. Adding a new port requires a minor-version bump and a port-doc PR; adding an adapter requires only dropping a markdown file at the conventional path.
3. **Open to extension, closed to modification — applied to *structure*.** The structure (port names, conventions, paths) is closed-to-modification: it changes only via deliberate framework releases. Behavior is open-to-extension via markdown: users add, edit, or remove markdown adapters at the conventional paths to change what the harness does. **It must be unnecessary for a user to modify framework files (Go source, embedded templates) to change behavior.** If a behavior change requires editing the framework, the port surface has a gap and the right answer is a new port or a richer contract — not a fork.
4. **Convention over configuration.** Drop a file at the conventional path → it loads. The convention table lives at `docs/conventions.md` and is the canonical reference. Generators (`keystone new ...`) scaffold every shape so authors never have to remember a path.
5. **Project content belongs to the user.** Anything under `harness/` *except* `harness/plugins/` is owned by the user, tracked in their git, freely editable. The framework writes to it only on `keystone init` (and refuses to overwrite without `--reset`).
6. **Plugins are read-only and atomic.** Plugins under `harness/plugins/<name>/` exist exactly as they did at the pinned tag. The runtime verifies hashes on every cascade resolution; drift triggers wipe-and-reinstall. Editing plugin files is unsupported and silently reverted.
7. **Determinism in the cascade.** For any `<port>/<name>`, exactly one file wins. Project files (`harness/<port>/<name>.md`) always beat plugin files. Among plugins, the first occurrence in a pre-order walk of the `keystone.json` tree wins. `strict` from any ancestor blocks descendants from overriding. The resolver is pure.
8. **No backward-compat sludge inside the framework.** 1.0 is a clean break. Old 0.x layouts are not silently shimmed. The 0.x → 1.0 path is a documented one-time `init --reset`.

---

## 3. Architecture sketch

### The ports (the things the framework can be extended at)

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

### Repo layout (keystone framework — target 1.0 state)

```
keystone/
├── cmd/keystone/                      # CLI entrypoint
├── internal/framework/                # framework runtime — users never edit this
│   ├── loader/                        # convention walk + cascade resolution
│   ├── manifest/                      # JSON manifest parsing & validation
│   ├── lockfile/                      # SHA + per-file hash pins
│   ├── plugins/                       # fetch, install, integrity-check, drift-reset
│   ├── migrate/                       # framework migration runner (config-schema only)
│   ├── budget/                        # context budgeting (Phase 5)
│   ├── adapters/                      # per-agent codegen for menu files
│   └── scaffold/
│       ├── generators/                # `keystone new ...` generators
│       └── templates/                 # default content (guides/, corpus/, sensors/, ...)
│                                      #   embedded via go:embed; copied into
│                                      #   harness/ on `keystone init`
├── docs/
│   ├── ports/                         # port contracts (the framework's English API)
│   ├── conventions.md                 # Rails-like convention table
│   ├── adr/                           # architecture decision records
│   ├── upgrade-0.x-to-1.0.md          # one-time upgrade narrative for consumers
│   └── schemas/                       # JSON Schemas for every config file
├── migrations/<version>/NNN-*.json    # framework-only migrations (config schemas)
└── examples/                          # reference fixtures used in CI
```

There is no `plugins/` directory in the framework repo — defaults are *templates*, not plugins. Plugins are an external concept owned by other repos.

### Repo layout (consumer — after `keystone init`)

```
<project>/
├── keystone.json                      # plugin tree + framework version
├── keystone.lock.json                 # SHA + per-file hash pins for vendored plugins
├── .gitignore                         # includes `harness/plugins/`
└── harness/
    ├── guides/  corpus/  sensors/  playbooks/  actions/  adapters/
    │                                  # project-owned; edit freely; lives in git;
    │                                  # versioned with the project's git tags
    ├── learning/  archive/            # flywheel sinks; user-owned
    └── plugins/                       # gitignored; read-only; hash-verified
        └── <plugin-name>/
            └── guides/ corpus/ sensors/ ...
```

The framework binary stays separately installed (brew/curl/scoop). The consumer repo contains no framework code.

### Cascade resolution

`keystone.json` declares the plugin tree:

```json
{
  "version": "1",
  "framework_version": "1.0.0",
  "plugins": [
    { "name": "acme-org",
      "source": "git+https://github.com/acme/keystone-policy.git",
      "version": "v2.0.0",
      "strict": { "guides": ["data-handling"] },
      "children": [
        { "name": "acme-platform",
          "source": "git+https://github.com/acme/platform-policy.git",
          "version": "v1.4.0",
          "strict": { "sensors": ["rubocop"] },
          "children": [
            { "name": "acme-payments",
              "source": "git+https://github.com/acme/payments-policy.git",
              "version": "v0.9.0" }
          ]
        }
      ]
    }
  ]
}
```

Only external plugins appear here. Built-in defaults are scaffolded into `harness/<port>/` on `init` and are not declared anywhere — they're just project content.

Resolution rules:
- **Project always wins.** Anything under `harness/<port>/` is at the front of the priority list. The framework never lets a plugin override a project file.
- **Among plugins: order = precedence.** A pre-order walk of `plugins[]` produces a deterministic priority list. Earlier nodes win against later ones for the same `<port>/<name>`.
- **Strict locks downward.** A `strict` declaration on a node blocks any deeper node (its descendants and any later siblings deeper in the tree) from shipping that item. Lockfile / `keystone verify` catches violations at install time.
- **One file loads, the rest don't.** When resolving `<port>/<name>`, only the winning file is read into context. The framework never composes overlapping files.
- **Per-port depth limits.** Sensors and any other ports that should not appear "above" a certain plugin depth declare `max_depth` in their port contract. Replaces the hardcoded tier enum.

### Plugin integrity: vendoring, hashing, drift-reset

Plugins live under `harness/plugins/<name>/`, gitignored. The flow:

1. **Resolve.** Read `keystone.json`. For each plugin node, `source` + `version` resolves to a git tag (the only supported source initially; local-path may follow — see open questions).
2. **Fetch.** `keystone install` git-clones each plugin tag into a content-addressable cache at `~/.cache/keystone/plugins/<sha>/`. Re-installs are cache hits.
3. **Lock.** Compute SHA-256 per file under the plugin tree. Write into `keystone.lock.json` keyed by `plugins.<name>.files.<relative-path>`. The lockfile is committed.
4. **Materialize.** Copy from cache into `harness/plugins/<name>/`. Mark files read-only at the filesystem level (chmod 0444 on POSIX) where possible.
5. **Verify on load.** Before every command that reads the cascade (`verify`, agent invocations of the loader, etc.), the runtime walks `harness/plugins/<name>/`, recomputes hashes, and compares to the lockfile.
6. **Drift = reset.** Any mismatch — extra file, missing file, modified content — triggers a hard reinstall: the plugin's vendor directory is removed and re-materialized from the pinned tag. No partial recovery, no merge, no warning prompt. The runtime logs which plugin was reset and which file(s) drifted.

The hash check is layered on top of the existing SHA pinning: the SHA pin guarantees the *source* tag hasn't moved; the per-file hash + drift-reset guarantees the *vendored copy* hasn't been edited locally.

`keystone install` is implicit on first command if `harness/plugins/` is empty or any plugin is missing. It can also be invoked explicitly.

---

## 4. Phases

Each phase: scope → deliverables → exit criteria → risks.

### Phase 0 — Foundation & decision log (target: 0.14.0)

**Status:** ✅ Complete (2026-06-08, commit `6b6295f`).

**Scope.** Settle naming, contract surface, and the major design decisions before shuffling code. Capture ADRs.

**Deliverables.**
- `docs/adr/0001-naming.md` — keep "Keystone"; reposition as framework.
- `docs/adr/0002-framework-client-boundary.md` — what counts as framework, what counts as client.
- `docs/adr/0003-ports-and-adapters.md` — list the ports; freeze the names.
- `docs/adr/0004-cascade-and-json-config.md` — JSON tree + pre-order precedence + strict semantics.
- `docs/adr/0005-conventions-not-plugins.md` — defaults are scaffolded at conventional paths in `harness/<port>/`; plugins are external only.
- `docs/adr/0006-vendored-readonly-plugins.md` — plugins live under `harness/plugins/`, are gitignored, hash-verified on every cascade resolution, and wiped-and-reinstalled on any drift.
- `docs/adr/0007-no-backward-compat-at-1.0.md` — 0.x → 1.0 is a one-time `init --reset`; no shims.
- `docs/adr/0008-versioning-policy.md` — SemVer with explicit compatibility guarantees post-1.0.
- `docs/ports/` — first draft of port contracts (one page each).
- `README.md` reframed: "Keystone - the agent harness framework for any project."

**Exit criteria.**
- All ADRs reviewed and merged.
- No code moved yet. No content moved yet. Pure intent capture.

**Risks.** Bikeshedding on naming. Time-boxed: 3 days.

---

### Phase 1 — Framework / client boundary + YAML→JSON sweep (target: 0.15.0)

**Status:** ✅ Complete (2026-06-08). Six sub-commits: CLI relocation (`c967b23` + fixup `1791f5c`), loader + manifest extraction (`0e598fc`), lockfile + cascade verify (`8779834`), migrate runner with JSON-only ops (`309ebbe`), YAML→JSON sweep + 0.x migration deletion (`5b30b67`), yaml.v3 drop + JSON Schemas (this commit).

**Scope.** Physically separate framework Go code from default content. Move sources under `internal/framework/`, `cmd/keystone/`. Convert every config/manifest/lockfile in the repo from YAML to JSON. Drop YAML loaders entirely. The 0.x migration files are removed (no backward compat to preserve).

**Deliverables.**
- `cmd/keystone/main.go` — relocated entrypoint.
- `internal/framework/{loader,manifest,lockfile,plugins,migrate,scaffold,adapters}/` — code split out from the root-level files (`policy_*.go`, `migrate.go`, `init.go`, etc.).
- A new `Loader` interface that takes a resolved cascade (project layer + plugin tree) and returns per-port resolved files.
- **JSON-only formats.** All in-tree YAML rewritten:
  - `keystone-policy.yaml` (fixtures + docs) → `keystone-plugin.json`.
  - `.keystone.lock` (YAML) → `keystone.lock.json`.
  - JSON Schemas under `docs/schemas/` for `keystone.json`, `keystone-plugin.json`, `keystone.lock.json`, `migration.json`.
  - `gopkg.in/yaml.v3` dropped from `go.mod`.
- The existing `migrations/0.7.0/...0.13.0/*.yaml` chain is **removed**, not converted — those migrations served the 0.x → 0.x path that is not preserved at 1.0.
- Unit tests on the loader against in-memory fixtures in `examples/`.

**Exit criteria.**
- Build and tests pass with the new layout.
- No YAML references in `go.sum`.
- Loader resolves a fixture project + plugin tree deterministically.

**Risks.** Hidden coupling between root-package files. Mitigation: split in one PR per concern (loader first, lockfile second, JSON conversion third, YAML deletion last).

---

### Phase 2 — Convention-scaffolded defaults (target: 0.16.0)

**Status:** ✅ Complete (2026-06-08). Five sub-commits: relocate templates under `internal/framework/scaffold/templates/` (`aabc012`), make universal-principles opt-in via a `starter` category (`8ebda8d`), configurable harness root via `--harness-root` (`07efbdc`), idempotent default `init` + `--reset --i-understand-this-is-destructive` (`b2a87eb`), docs + plan status (this commit).

Diverged from the originally drafted scope in two ways, both at user direction:
- **Universal content is opt-in**, not always-scaffolded. It lives at `templates/optional/starter/universal-principles/harness/{guides,corpus}/principles/` and is installed only when the user passes `--starter universal-principles` (or selects it in the `huh` menu). Treats engineering-principles content as opinion, not as a built-in default.
- **Harness root is configurable.** `--harness-root <name>` at init picks the folder name (default `harness`); downstream commands accept the same flag. `internal/framework/config.DefaultHarnessRoot` is the single source of truth.

**Scope (original).** Move every piece of default content into the scaffold templates directory. `keystone init` copies the templates into the consumer's `harness/<port>/` at conventional paths. The user owns those files after init.

**Deliverables.**
- `internal/framework/scaffold/templates/` — the canonical default content, embedded via `go:embed`:
  - `guides/` — universal engineering guides (currently `harness/policies/universal/guides/...`).
  - `corpus/` — paired reasoning files (currently `harness/policies/universal/corpus/...`).
  - `sensors/` — default sensor templates (currently `harness/sensors/...`).
  - `actions/` — lifecycle actions (spec, orient, verify, review, learn, bootstrap, audit, synthesize, mode).
  - `playbooks/` — the `task` playbook.
  - `adapters/<agent>/` — per-agent bindings for `claude-code`, `codex`, `cursor`, `aider`, `continue`, `cline`, `goose`, `github-copilot`, `pi`, `_generic`.
- `keystone init` walks the templates and writes them into the consumer's `harness/<port>/`. Existing files are not overwritten (idempotent — the user's edits win).
- `keystone init --reset` overwrites. This is the documented one-time path for 0.x → 1.0 upgrade; it requires a `--i-understand-this-is-destructive` confirm flag.
- The loader treats `harness/<port>/` uniformly — there is no special "default" namespace. Defaults look like project content because they *are* project content after `init`.
- Removing a template from the binary affects new installs only; existing installs are unaffected (the file already belongs to the user).

**Exit criteria.**
- `keystone init` on a fresh dir produces a project with all defaults at conventional paths under `harness/<port>/`.
- The runtime has no concept of "embedded plugin." `plugins/` does not exist in the framework repo.
- A fresh install's `harness/` (minus `harness/plugins/`) is fully editable and git-tracked end to end.

**Risks.** Template drift over framework versions — a user's scaffolded copy diverges from upstream defaults. Mitigation: `keystone doctor` (Phase 4) diffs the user's copy against current templates and reports rot; the user opts in to applying upstream changes by hand or via a generator overwrite.

---

### Phase 3 — Vendored read-only plugins (target: 0.17.0)

**Status:** ✅ Complete (2026-06-08). Eight sub-commits:
- `80fdcc9` — `keystone.json` data model + shorthand source format (`tacoda/tacoda-org@0.2.0`).
- `2cd8de4` — `lockfile.Plugins` field; drop transitional `KeystoneInfo.HarnessRoot` (now lives in `keystone.json`).
- `7da637f` — `internal/framework/plugins/` (Fetch + content-addressable cache, Install with per-file hashes, Verify, Reset).
- `0d9e78c` — refactor: centralize harness-root resolution (`keystone.json` → flag → default); remove policy CLI.
- `36894af` — rename `migrate` → `patch` across the codebase (cmd, package, schema, templates, embed comments). `keystone migrate` prints a rename notice.
- `d4db3b5` — `keystone install`, `keystone plugin add|update|remove`. Shorthand parser, default-name derivation, top-level tree append, idempotent re-install.
- `af92007` — loader cascade refactored against the plugin tree (no more tier enum); new `keystone verify` runs drift detection + strict-cascade reporting with path context (`acme-org > acme-platform`).
- `167ed05` — `keystone init` writes `keystone.json` and ensures `<harness-root>/plugins/` is in `.gitignore`.

Diverged from the originally drafted scope in two user-directed ways:
- **Shorthand source format**: `[<host>/]<owner>/<repo>@<version>` instead of the prior `git+<url>#<rev>`. Default host is `github.com`; per-source override by writing the host explicitly.
- **`migrate` → `patch` rename**: the framework's notion of a versioned forward-only change is now called a `patch`, reflecting its narrowed 1.0 scope (config-schema bumps; project content is git-tracked).

**Scope (original).** Replace the flat Org/Team/Project tier semantics with the nested plugin tree declared in `keystone.json`. Build the vendor flow: fetch → hash → write → verify-on-load → drift-reset.

**Deliverables.**
- `keystone.json` consumer-side schema finalized at `docs/schemas/keystone.json.schema.json`.
- `internal/framework/plugins/`:
  - `fetch` — git+tag clone of each plugin into a content-addressable cache (`~/.cache/keystone/plugins/<sha>/`).
  - `install` — copy from cache into `harness/plugins/<name>/`, compute per-file hashes, write `keystone.lock.json`, mark files read-only (chmod 0444 on POSIX, best-effort on Windows).
  - `verify` — walk `harness/plugins/<name>/`, recompute hashes, compare to lockfile, return list of drifted plugins.
  - `reset` — `rm -rf harness/plugins/<name>/`, re-run `install` for that plugin.
- Loader integrates `verify` as a precondition on cascade resolution: any drift triggers `reset` for the affected plugins before resolution proceeds. The verification result is cached for the duration of a single process invocation to avoid repeated walks.
- `.gitignore` template (written on `init`) includes `harness/plugins/`.
- CLI surface:
  - `keystone install` — explicit; also implicit on first cascade resolution when `harness/plugins/` is empty or incomplete.
  - `keystone plugin add <source>@<version>` — appends to `keystone.json`, installs.
  - `keystone plugin update <name> [--to <version>]` — bumps the pinned `version` in `keystone.json`, re-vendors.
  - `keystone plugin remove <name>` — removes from `keystone.json` and `harness/plugins/`.
- `keystone verify` upgraded to walk the plugin tree, report `strict` violations with path context (`acme-org > acme-platform > acme-payments`), and report any drifted plugins that were just reset.
- Examples fixture: a five-deep nested tree with `strict` mid-tree, and a fixture demonstrating a drift-reset triggered by hand-editing a vendored file.

**Exit criteria.**
- A fresh clone of a project + `keystone install` produces a working harness with vendored plugins.
- Manually editing a vendored file and running any cascade-touching command silently resets that plugin, with a log line naming the reset and the offending file(s).
- The Org/Team/Project tier enum is gone from runtime code paths.

**Risks.** Plugins fetched via git over flaky network. Mitigation: content-addressable cache; `--offline` mode that fails fast on cache miss.

---

### Phase 4 — Conventions, generators, doctor (target: 0.18.0)

**Status:** ⏳ Pending.

**Scope.** Lock in Rails-like ergonomics. Document conventions exhaustively. Ship scaffolding generators. Add `keystone doctor` to flag convention violations and plugin drift. Normalize cross-file path references in scaffolded content.

**Deliverables.**
- `docs/conventions.md` — the canonical convention table. Path → port → activation.
- **Path convention for markdown cross-references** — codify and enforce two rules:
  1. **Markdown links to other harness files** are written *relative to the harness root*, not relative to the source file. Example: a guide at `<harness-root>/guides/process/spec.md` linking to its paired corpus references `corpus/process/spec.md` (resolved against the harness root), not `../../corpus/process/spec.md`. Relative paths with `../` or `./` segments are forbidden inside harness markdown.
  2. **Code-relevant content** (sensor commands, file paths the agent reads/writes, the lockfile, the project's source tree) is written *relative to the repo root* — the directory that owns the harness folder and the `keystone.lock.json`.
  - One-time pass to rewrite the ~hundreds of existing relative links in `internal/framework/scaffold/templates/harness/` and `internal/framework/scaffold/templates/optional/starter/` to follow rule 1.
  - `keystone doctor --paths` (or a default check) flags any `../` segment in harness markdown.
  - Generators emit conformant paths by default.
- Generators:
  - `keystone new guide <topic>/<name>` — guide + paired corpus stub.
  - `keystone new sensor <name>` — sensor with frontmatter.
  - `keystone new action <name>` — action.
  - `keystone new playbook <name>` — playbook referencing existing actions.
  - `keystone new adapter <agent>` — per-agent bindings.
  - `keystone new plugin <name>` — scaffold a *plugin repo* (a separate git repo that other projects can vendor).
- `keystone doctor`:
  - Checks every file under `harness/` (excluding `harness/plugins/`) against the conventions; reports misplacements, missing frontmatter, broken forward-links, and any `../`-style paths that violate the rule above.
  - Diffs the user's scaffolded defaults against current templates; reports drift the user may want to refresh.
  - Calls `plugins.verify`; reports any drifted plugins that were reset.
- Dropping a file at a conventional path is sufficient to activate it — no registry edits.

**Exit criteria.**
- Authoring a new sensor / guide / action / playbook / plugin takes one command + filling in the body.
- `doctor` runs in <2s on a fixture project with 200 markdown files.
- `doctor` reports zero `../`-segment paths in a fresh `keystone init` install.
- No code change required to support a new agent — the adapter generator + a default template suffice.

**Risks.** Generator output drift from hand-authored files. Mitigation: a "regenerate" mode that diffs current scaffolds against canonical templates and reports rot.

---

### Phase 5 — Context budgeting as a first-class feature (target: 0.19.0)

**Status:** ⏳ Pending.

**Scope.** Make context budget a framework concern, not an emergent property. Define per-port budgets, measure load, report, enforce.

**Deliverables.**
- `internal/framework/budget/` — token estimator (whitespace-tokenizer, approximate) and budget allocator per port.
- `keystone.json` accepts an optional `budgets` block:
  ```json
  "budgets": {
    "guides":   { "max_tokens": 6000 },
    "corpus":   { "max_tokens_per_load": 4000 },
    "adapters": { "max_tokens": 1000 }
  }
  ```
- A `context-budget` sensor (default-scaffolded) reports the current ambient load and warns when guides approach the budget.
- `keystone doctor --budget` shows per-port utilization and the top contributors.

**Exit criteria.**
- A fresh install reports its ambient load on `init`.
- Going over budget produces a doctor warning (raisable to error by the project).
- Documented in `docs/ports/budget.md`.

**Risks.** Tokenizer drift versus real model tokenizers. Mitigation: document the heuristic; allow an optional `--tokenizer=tiktoken` mode via an opt-in `init` flag (per the [install-time options](../.claude/projects/-Users-tacoda-tacoda-keystone/memory/feedback_install_time_options.md) preference).

---

### Phase 6 — Hardening, upgrade guide, 1.0 (target: 1.0.0)

**Status:** ⏳ Pending.

**Scope.** Freeze the surface. Document compatibility guarantees. Write the one-time 0.x → 1.0 upgrade narrative. Cut the release via tag push.

**Deliverables.**
- `docs/upgrade-0.x-to-1.0.md` — single-page guide for consumers. Canonical command sequence:
  1. Commit and tag the existing 0.x state in the project repo (the user's backup mechanism).
  2. Upgrade the keystone binary.
  3. Run `keystone init --reset --i-understand-this-is-destructive` to scaffold 1.0 defaults.
  4. Run `keystone install` to materialize plugins.
  5. Manually port any non-default content from the prior tag — `git diff <0.x-tag>` is the diff to apply.
- `docs/compatibility.md` — what 1.0 promises:
  - **Stable:** port names, port contracts, `keystone.json` schema, `keystone.lock.json` schema, plugin manifest schema, generator output shape, conventions table.
  - **Free to evolve:** `internal/framework/` packages, scaffold template contents (treated as content, not API), warning/log text.
  - **Deprecation cycle (post-1.0):** one minor with a warning before removal in the next major.
- Framework migration runner is in place but unused at 1.0 — reserved for future `keystone.json` schema bumps.
- Full regression suite: golden-file tests for `init`, `install`, `verify`, `doctor`, generators, plugin drift-reset.
- Release notes summarizing the 0.13 → 1.0 break.

**Exit criteria.**
- All ADRs accepted.
- All ports documented at `docs/ports/<port>.md`.
- Upgrade from a 0.13.0 fixture install verified end-to-end against the upgrade guide.
- `goreleaser` tag push to `v1.0.0`. (Per [release policy](../.claude/projects/-Users-tacoda-tacoda-keystone/memory/feedback_release_via_tag_push.md) — never `gh release create`.)

**Risks.** Users with extensive 0.x customization will lose work if they skip the backup step. Mitigation: prominent warning in the upgrade guide; `--reset` refuses without `--i-understand-this-is-destructive`.

---

## 5. Framework migrations (the only migration tier)

Project content and plugins do not migrate. Only framework-binary concerns ever migrate user files.

### What framework migrations CAN touch
- `keystone.json` — schema upgrades when the binary's expected shape changes.
- `keystone.lock.json` — same.
- `.gitignore` — adding entries the binary requires.

### What framework migrations CANNOT touch
- Anything under `harness/guides/`, `harness/corpus/`, `harness/sensors/`, `harness/playbooks/`, `harness/actions/`, `harness/adapters/` — user content, git-tracked, versioned by the project's own tags.
- Anything under `harness/plugins/` — plugin content, managed by the vendor flow, not migrations.
- Anything under `harness/learning/`, `harness/archive/` — user data.

### Rules
- **Idempotent.** Re-running on an already-migrated install is a no-op.
- **Scoped narrowly.** Config-schema bumps only. No content rewrites.
- **One deprecation cycle (post-1.0).** Schema fields renamed or removed live behind a one-minor warning shim.
- **`migrate --dry-run` is the contract.** Plan output is the source of truth.

### File format

All migration files are JSON. Operations limited to `ensure_section`, `frontmatter_set`, `replace_block` against config files only. `move_file` and `delete_file` for user content are forbidden — those operations belong to the user's git history, not the runtime.

JSON Schema at `docs/schemas/migration.json.schema.json`.

---

## 6. Open questions (resolve in Phase 0)

1. Per-agent menu file generation — does it become a port adapter (so a new agent doesn't require Go changes)? Default position: yes, by Phase 4.
2. Plugin sources beyond `git+tag` — should `keystone.json` support a local-path source for plugin authors testing in-tree? Default position: yes, with a clear warning that local-path plugins disable hash verification.
3. Drift-reset frequency — verify on every cascade resolution (safest) vs only on explicit commands (`verify`, `doctor`, `install`). Default position: every resolution, with the result cached for the duration of a single process invocation.
4. Read-only enforcement on Windows — chmod-based marking is POSIX-only. Default position: best-effort flag for cosmetic UX; the hash check is the real enforcement on all platforms.
5. `keystone init --reset` ergonomics — should it dump a `.keystone-reset.diff` of the user's pre-reset state for easier manual porting? Default position: yes, low cost, helpful for the 0.x → 1.0 upgrade specifically.

---

## 7. What to do next

If this plan is accepted in shape, the next concrete step is Phase 0 — draft the eight ADRs and the first cut of `docs/ports/`. Everything else is gated on those landing.
