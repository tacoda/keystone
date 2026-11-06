# Keystone 5.0.0: standard vocabulary plan

**Goal:** Split the charter into two categories, **constraints** and
**capabilities**. Rename every primitive to the name the ecosystem already uses
(rules, hooks, permissions, commands, skills, subagents, tools). Make
`AGENTS.md` + `.agents/` the canonical source, conform skills to the Agent
Skills spec, and reach every other host through symlinks or generated shims.

**Why a major:** the canonical directory moves (`.charter/` to `.agents/`),
kinds are renamed or folded, frontmatter changes shape, and CLI/MCP names
change. A `v5_0` migration moves installed repos forward.

**Builds on:** `keystone-charter-rebrand-plan.md` (4.0.0). The charter/harness
split from `GLOSSARY.md` stays. The charter is now described as two halves:
constraints and capabilities.

---

## The split (load-bearing)

One question sorts every primitive: **does it limit or check what the agent
does, or does it add something the agent can do?**

- **Constraint:** narrows, gates, or verifies behavior. Always on or
  event-driven. The agent cannot opt out.
- **Capability:** extends what the agent can do. Invoked by the user, by the
  model on description match, or by delegation.

A hook that runs `go vet` is a constraint. A skill that knows how to scaffold
a migration is a capability. A tool the hook calls is still a capability; the
hook that calls it is the constraint.

---

## Naming map

### Constraints

| 4.x kind | 5.0 kind | Canonical path | Standard it maps to |
|---|---|---|---|
| `guide` | **rule** (`globs:` becomes `paths:`) | `.agents/rules/<id>.md` | Claude rules, Cursor rules, Copilot instructions, Continue rules, nested `AGENTS.md` |
| `corpus` | **reference** | `.agents/references/<id>.md` | none (keystone extension: the "why" behind a rule) |
| `sensor` (computational) | **hook** `type: command` | `.agents/hooks/<id>.md` | host hook events (`PreToolUse`, `Stop`, ...) |
| `sensor` (inferential) | **hook** `type: agent` or `type: prompt` | `.agents/hooks/<id>.md` | Claude prompt/agent hooks; elsewhere bridged via `keystone hook fire` |
| `posture` | **permission** | `.agents/permissions/<id>.md` | Claude `permissions`, Codex sandbox/approval, opencode `permission`, Cursor CLI permissions |
| `eval` | **eval** | `.agents/evals/<id>.md` | none (keystone extension) |

### Capabilities

| 4.x kind | 5.0 kind | Canonical path | Standard it maps to |
|---|---|---|---|
| `skill` | **skill** | `.agents/skills/<name>/SKILL.md` | Agent Skills spec |
| `playbook` | **skill** (folded) | `.agents/skills/<name>/SKILL.md` | Agent Skills spec; gates become a body checklist |
| `pattern` | **skill** (folded) | `.agents/skills/<name>/SKILL.md` | Agent Skills spec (on-demand recipe) |
| `command` | **command** | `.agents/commands/<id>.md` | Claude/Cursor/opencode commands, Copilot prompt files |
| `agent` | **subagent** | `.agents/agents/<id>.md` | Claude/opencode subagents, Copilot custom agents |
| `tool` | **tool** | `.agents/tools/<id>.md` | MCP tool, served by the keystone MCP server |
| `document` | **document** | `.agents/documents/<id>.md` | none (keystone extension) |
| `concern` | retired | inlined into consumers | `allowed-tools` on the consuming skill/subagent |

13 kinds become 10. Three are keystone extensions (`reference`, `eval`,
`document`) and are labeled as such everywhere they appear.

### Other names

| 4.x | 5.0 |
|---|---|
| `CHARTER.md` (canonical root) | **`AGENTS.md`** (canonical root). `CHARTER.md` removed. |
| `.charter/` | **`.agents/`** for standard primitives, **`.keystone/`** for keystone-private state |
| `.charter/INDEX.json`, `INDEX.lite.json` | `.keystone/INDEX.json`, `.keystone/INDEX.lite.json` |
| `.charter/state/`, `learning/`, `work/`, `archive/`, `lockfile.json`, `sources/` | `.keystone/<same>` |
| `.charter/adapters/<host>/` (per-host docs) | `.keystone/hosts/<host>/` |
| signal | signal (unchanged). Documented as "a custom hook event". |
| `keystone new guide\|sensor\|posture\|agent` | `keystone new rule\|hook\|permission\|subagent` |
| `--charter-root` | `--agents-root` |
| "the charter" | unchanged as the umbrella term: charter = constraints + capabilities |

"Agent" now means the harness's main agent only (as in `AGENTS.md`). A role you
author and delegate to is a **subagent**. The directory stays `agents/`
because that is what every host reads.

---

## Standards conformance

### AGENTS.md

- Root `AGENTS.md` is the canonical, keystone-rendered entrypoint. It carries
  the iron laws, always-on rules, the constraint/capability table, and a
  pointer to `.keystone/INDEX.lite.json`.
- Plain Markdown, no frontmatter, per the spec.
- Optional: rules whose `paths:` resolve to exactly one directory may also emit
  a nested `<dir>/AGENTS.md` (opt-in, `keystone.json "nestedAgentsMd": true`).
  Off by default because glob-to-directory mapping is lossy.

### Agent Skills spec

- Canonical `SKILL.md` frontmatter uses only spec fields: `name` (must equal the
  directory name, lowercase-hyphen, max 64), `description` (max 1024),
  `license`, `compatibility`, `metadata` (string to string), `allowed-tools`.
- Keystone fields move: `kind` is inferred from path. `id` becomes `name`.
  `triggers:` folds into `description` (the spec's activation contract).
  `tags:` becomes `metadata.tags` (comma-joined string).
- Optional `scripts/`, `references/`, `assets/` subdirectories are preserved
  and projected verbatim.
- `keystone lint` runs a spec validator on every skill. A skill that fails the
  spec fails lint.

### `.agents/`

- `.agents/` holds only standard primitives in standard formats, so a host
  that reads `.agents/skills/` natively needs no shim.
- Keystone-private files never land in `.agents/`. They live in `.keystone/`.
- The other `.agents/` subdirectories (`rules/`, `commands/`, `agents/`,
  `hooks/`, `permissions/`) have no cross-host spec today. Keystone defines
  their frontmatter as a superset of the Claude Code format, which is the
  closest thing to a de facto standard. That choice is recorded in ADR 0010.

---

## Host reach: native, symlink, or shim

Every (host, kind) pair resolves to one of four strategies. `HostProfile`
grows from four booleans into this per-kind matrix.

| Strategy | When | Example |
|---|---|---|
| **native** | host reads `.agents/<kind>` in the canonical format | a host that reads `.agents/skills/` |
| **symlink** | host reads the identical format at another path | `.claude/skills/<name>` to `../../.agents/skills/<name>` |
| **shim** | host reads a different format or file | Cursor `.mdc`, Copilot `.instructions.md`, `.claude/settings.json` hooks + permissions, `CLAUDE.md` with `@AGENTS.md` |
| **unsupported** | host has no equivalent | subagents on Aider. `AGENTS.md` lists the gap in a per-host "on this host" block. |

Rules for links and shims:

- **Per-entry symlinks, never whole-directory links.** `.claude/skills/` may
  hold user skills keystone does not own. Linking each entry leaves those alone.
- **Ownership is tracked** in `.keystone/lockfile.json`. Prune removes only
  entries keystone created.
- **Copy fallback.** `keystone.json "linkMode": "symlink" | "copy"`. Default
  `symlink` on macOS/Linux, `copy` on Windows (symlinks need Developer Mode
  there, and `install.ps1` ships today). `copy` writes a generated header and
  is drift-checked by `keystone verify`.
- **Security-bearing shims never degrade silently.** If a host cannot express a
  `deny` permission or a blocking hook, `keystone project` prints it and
  `AGENTS.md` lists it as "not enforced on <host>". It never drops it quietly.

### Host matrix

Verified in Phase 0. The sourced matrix, the strategy per cell, and the
findings that shaped the design live in [`docs/hosts.md`](docs/hosts.md).
Strategies there add two to the four above: **index** (listed in `AGENTS.md`,
opened on demand) and **report** (a constraint the host cannot enforce).

Out of scope unless you add it: new hosts (Gemini CLI, Windsurf, Amp, Zed).
The matrix makes adding one a profile entry plus golden files.

---

## Phases

Each phase ends green: `go build ./... && go vet ./... && go test ./...`,
`keystone verify`, and `pre_commit_code_health_safeguard` with no regression.
Work on `feat/5.0-standard-vocabulary`, one PR per phase or pair of phases.

### Phase 0: Research and decision lock

- Verify the host matrix against each host's current docs (Context7 or the
  official docs). Record source URL and date per cell in `docs/hosts.md`.
- Pin the Agent Skills spec version and the `AGENTS.md` spec version we
  conform to. Record which hosts read `.agents/skills/` today.
- Write **ADR 0010: constraints and capabilities**. It records the split, the
  naming map, the fold decisions (playbook, pattern, concern), `.agents/` vs
  `.keystone/`, and the Claude-superset frontmatter choice for unspecified kinds.
- Exit: ADR accepted, every matrix cell sourced or marked unsupported.

### Phase 1: Domain model

- `internal/framework/primitive`: new `Kind` set (10 kinds) plus a `Category`
  type (`constraint`, `capability`) on every kind. `AllKinds`, the dir-to-kind
  map, and `primitive.go` comments updated.
- Legacy alias table (old kind to new kind) used **only** by the migration and
  by an error message: `kind "guide" was renamed to "rule"; run keystone migrate`.
  The loader does not read old kinds at runtime.
- Sensor split: `mode: computational` becomes `type: command`;
  `mode: inferential` becomes `type: agent` (with `subagent:`) or
  `type: prompt`. `on:` becomes `event:`. Matchers stay.
- INDEX schema v5: adds `category`, drops `kind` aliases. `INDEX.lite.json`
  groups by category.
- Verify: unit tests for kind parse, category assignment, alias error text.

### Phase 2: Canonical layout and standard frontmatter

- Config: `DefaultAgentsRoot = ".agents"`, `DefaultStateRoot = ".keystone"`.
  `--agents-root` flag. Walk reads `.agents/`; INDEX, lockfile, state,
  learning, work, archive, sources, hosts read and write `.keystone/`.
- Skill parser and writer speak the Agent Skills spec (name, description,
  metadata, allowed-tools). Spec validator in `keystone lint`.
- Rule, command, subagent, hook, permission frontmatter: Claude-superset shapes
  from ADR 0010, with a JSON schema each under `docs/schemas/`.
- Verify: parse/round-trip tests per kind; spec validator rejects a bad `name`,
  a name/dir mismatch, and an over-long description.

### Phase 3: Projection engine

- Replace the hard-coded `.claude` paths in `primitive/project.go`
  (`ProjectionRelPath`) with a per-host, per-kind strategy table driven by
  `HostProfile`.
- Implement the four strategies. Symlinks are relative, per entry, and
  recorded in the lockfile. `linkMode: copy` writes headed copies.
- Prune: remove lockfile-owned entries that no longer have a source.
  Never touch unowned files.
- `keystone verify` gains a projection drift check: broken symlinks, edited
  copies, missing shims.
- Verify: golden-file tests per host on a fixture repo (one of each kind);
  a test that a user-authored `.claude/skills/mine/` survives project and prune.

### Phase 4: Host adapters

- Port each existing adapter (claudecode, cursor, opencode, aider,
  continueide, agnostic) onto the strategy table. Add profiles for codex,
  github-copilot, cline, goose, pi from the Phase 0 matrix. Hosts without
  native project config get `AGENTS.md` plus the "on this host" block only.
- `AGENTS.md` renderer replaces `CHARTER.md` + `RenderPointer`. Thin host files
  (`CLAUDE.md`, `.aider.conf.yml`, Cursor and Continue pointer rules) point at
  `AGENTS.md`.
- Scaffold templates under `internal/framework/scaffold/templates/targets/*`
  regenerate from the profiles instead of being hand-kept.
- Verify: golden files per host; `keystone project` on this repo is idempotent
  (second run writes nothing).

### Phase 5: Hooks and permissions (security-sensitive)

- Hooks: one bridge per host event (`keystone hook fire <event>`), as today.
  Where a host supports `prompt`/`agent` hook types natively, project them
  natively; elsewhere the bridge runs them. Signals stay the custom-event
  namespace.
- Permissions: one canonical `allow` / `ask` / `deny` model in Claude's rule
  syntax (`Bash(go test:*)`, `Read(./.env*)`). Translate per host where the
  host can express it. Report every rule a host cannot enforce.
- `deny` wins on conflict across files and cascade layers, the same as today.
- Run the `keystone-security-reviewer` subagent and the security-review skill
  on this phase's diff before merge.
- Verify: table tests per host translation; a test that an unexpressible
  `deny` is reported and never dropped; a test that a cascade `deny` beats a
  project `allow`.

### Phase 6: CLI, MCP, dashboard

- `keystone new rule|reference|hook|permission|eval|command|skill|subagent|tool|document`.
  Old `new <kind>` names exit non-zero with the new name.
- `keystone list`, `show`, `explain`, `charter coverage`, `charter conformance`
  group output by category.
- MCP tool names and descriptions renamed; MCP instructions describe the two
  categories. Dashboard nav splits into Constraints and Capabilities.
- Verify: `mcp_test.go`, CLI tests, dashboard route tests; no 404 on any nav link.

### Phase 6b: `keystone allocate` (convert existing setups; context de-bloat)

User-requested. One command for every "what's already here" case. Two
entry points into a charter, and nothing overwritten without `--overwrite`:

- **Seed a clean project:** `keystone init` (skips existing files by
  default). Its destructive `--reset` is replaced by `--overwrite`, which
  replaces conflicting template files and never deletes anything else.
- **Convert an existing setup:** `keystone allocate`.

`keystone allocate` reads, by default, every known host-specific setup and
root file:

- `CLAUDE.md` / `AGENTS.md` / `GEMINI.md` / `.goosehints` sections;
- `.claude/skills|agents|commands|rules/`, `.claude/settings.json` hooks and
  permissions;
- `.cursorrules`, `.cursor/rules/*.mdc`, `.cursor/skills|agents|commands/`;
- `.github/copilot-instructions.md`, `.github/instructions|prompts|agents/`;
- `.clinerules/`, `.windsurfrules`, `.continue/rules|prompts/`, `.opencode/`.

Each item is classified into a 5.0 kind (root-file sections by heuristic:
paths/globs → rule; "always run X before commit" → hook; allow/deny lists
→ permission; numbered user procedure → command, agent-triggered → skill;
long rationale → reference; iron laws stay in the root file) and mapped to
its `.agents/` path, converting frontmatter (`globs:`/`applyTo:` → `paths:`,
Cursor `alwaysApply` → no paths, Claude `name:` → spec skill, …).

- Default is a report: per item, source → target kind and path, and the
  always-loaded tokens it would save (pairs with Phase 6c budgets).
- `--apply` writes into `.agents/` only. It never overwrites an existing
  `.agents/` file (conflicts are reported) and leaves every original in place.
- `--overwrite` additionally records the originals as keystone-owned, so the
  next `keystone project` replaces them with links and shims. This is the
  only path that replaces user files.
- A `keystone-allocate` skill wraps the report so an agent can refine the
  borderline classifications.
- Verify: fixture repos per source (Claude-only, Cursor-only, Copilot-only,
  mixed) → report golden files; `--apply` round-trip leaves `keystone lint`
  clean and every original byte-identical; `--overwrite` then `project`
  leaves no unowned conflicts.

### Phase 6c: Context budget, reportable and enforced

User-requested. Today `internal/framework/budget` estimates tokens per
charter directory ("port") and `keystone doctor` prints it against
`keystone.json` `budgets` caps. That measures files on disk, not what an
agent loads. 5.0 measures the load.

- **Two tiers, per host.** *Always loaded:* root file (`AGENTS.md` /
  `CLAUDE.md` shim), rules without `paths:`, every skill and subagent
  `description` (hosts list them up front). *On demand:* path-scoped rules,
  skill/command/subagent bodies, references. Computed from the projection
  table, so each host's number reflects what that host reads.
- **`keystone budget`** prints both tiers per host, then per category and
  kind, with the top files. `--json` for CI and the dashboard.
- **Caps.** `keystone.json` `budgets` re-keyed to 5.0: `always_loaded`
  (the one that matters most), per-kind `max_tokens`, and `max_tokens_per_load`
  for any single file. 4.x port keys keep working (mapped to kinds).
- **Managed, not just reported.** `keystone verify` fails when a cap is
  exceeded, naming the file and the overage. Each overage suggests
  `keystone allocate` (Phase 6b), and `allocate` reports savings against the
  `always_loaded` cap.
- Surfaces: CLI, MCP `keystone_budget` tool, a dashboard Budget page.
- Out of scope: exact tokenizer counts. The estimate stays bytes/4 (it
  under-counts BPE by 25 to 35%, already documented); caps are set against
  the same estimate, so the comparison holds.
- Verify: golden test on a fixture charter for both tiers per host; a
  verify test that fails on an over-cap root file and passes after trimming.

### Phase 7: Migration and dogfood

- The layout half of `v5_0` landed in Phase 2 (moves only, merge into an
  existing `.agents/`, refuse on conflicting files or unknown entries; Down
  restores `.charter/`). This repo was migrated then.
- Correction (Phase 2): the migrations package's iron laws forbid rewriting
  user frontmatter (law 2) and require runtime back-compat (law 6). So 5.0
  does **not** rewrite frontmatter or turn 4.x names into a hard error.
  4.x kinds, dirs, `globs:`, and `on:` stay readable permanently; `keystone
  lint` warns with the 5.0 spelling, and an opt-in `keystone lint --fix`
  (user-invoked, shows the diff) applies it. `triggers:` is reported, not
  folded: the user moves it into `description`.
- `CHARTER.md` content merges into the new `AGENTS.md`; old projected files
  under `.claude/` etc. are replaced by links and shims (Phase 4 renderer).
- Installed policies: 4.x-shaped policy trees keep loading through the
  legacy dir and kind mapping (law 6). Republish keystone's own policies in
  5.0 shape; no resolver refusal.
- Dogfood: add `opencode` (and `codex`, `github-copilot`)
  to `keystone.json` `adapters`, `keystone index && keystone project`, commit.
- Verify: up/down round-trip test on a 4.x fixture; idempotent re-run test;
  `keystone verify` clean; `keystone charter conformance` CONFORMANT.

### Phase 8: Docs

- `README.md`, `GLOSSARY.md` (constraints vs capabilities, kind table, the
  "agent vs subagent" note), `CONVENTIONS.md`, `CONTRIBUTING.md`, `docs/**`,
  `docs/hosts.md`, and every rule under `.agents/rules/charter-content/`.
- `docs/upgrade-4.x-to-5.0.md`: what moves, what folds, what `lint` warns
  about, how to choose `linkMode`. Known conflict: a repo with both
  `guides/README.md` and a 4.x `rules/README.md` stops the migration (both
  map to `.agents/rules/README.md`); delete the `rules/` one, it described
  the retired guide/rule split. After upgrading, refresh the installed binary
  (`go install` or the release) before hooks fire: an old binary reads
  `.charter/` and finds nothing.
- `CHANGELOG.md` 5.0.0 entry with every breaking change listed.
- Run the voice chain from `~/.claude/VOICE.md` section 15 on README and the
  upgrade guide.

### Phase 9: Bug hunt and full verify

- Adversarial pass for rename fallout: stale `.charter` literals, broken
  `go:embed`, hard-coded `.claude` paths, old kind strings in templates and
  MCP text.
- Cross-platform: run the projection golden tests on Windows in CI (`copy`
  mode) and on macOS/Linux (`symlink` mode).
- Smoke test on one real external repo (a Parento repo or a scratch clone):
  `keystone migrate`, then open it in Claude Code and one other host, and check
  that rules, a skill, a command, a subagent, a hook, and a deny rule each work.
- `analyze_change_set` against `main` with no regression.

### Phase 10: Release candidate, then 5.0.0

- Tag `v5.0.0-rc.1` (tag push, GoReleaser). Dogfood for at least one week on
  real work.
- Fix what the RC surfaces, then tag `v5.0.0`. No `gh release create`.

### Phase 11: gh-pages site

- Update site copy for constraints and capabilities and the new layout, after
  the release so version references match. Confirm before pushing.

---

## Risks

- **Host facts change fast.** The matrix is the riskiest input. Phase 0 sources
  every cell, and the golden tests make a host change a one-file fix.
- **Lossy folds.** `triggers:` into `description`, playbook gates into a body
  checklist. Mitigated by the migration report and the RC window.
- **Symlinks on Windows and in git.** Mitigated by `linkMode: copy` and CI on
  Windows. Some tools follow symlinks badly; the copy mode covers them too.
- **Silent security downgrade.** A permission or blocking hook that a host
  cannot express. Mitigated by the loud report and the "not enforced" block.
- **Installed policies break.** Every 4.x policy needs republishing. The
  resolver error names the fix.
- **Two breaking majors in a row.** 4.0 moved `.harness/` to `.charter/`; 5.0
  moves it again. The upgrade guide should say why plainly: the standards
  settled on `.agents/`, and following them removes the per-host projection
  most users never wanted to maintain.

## Open decisions (recommendation first)

1. **Canonical root.** Recommend `.agents/` canonical with `.keystone/` for
   state. Alternative: keep `.charter/` canonical and project into `.agents/`
   (smaller migration, but the charter stays a keystone-only format).
2. **Folds.** Recommend playbook and pattern into skill, concern retired,
   corpus renamed `reference`. Alternative: keep playbook as its own
   capability kind if the human `gates:` need machine enforcement.
3. **agent vs subagent.** Recommend kind `subagent`, directory `agents/`.
4. **Default link mode.** Recommend symlink on macOS/Linux, copy on Windows.
5. **Nested `AGENTS.md` for scoped rules.** Recommend opt-in, off by default.

## Status

- Phase 0: done. ADR 0010 accepted; matrix in `docs/hosts.md`.
- Phase 1: done. 10 kinds + `Category`; INDEX v5.0 (`by_category`, lite
  `category`, constraints first); `paths:` / `event:` canonical with
  `globs:` / `on:` aliases; `primitive.IsKind` for user-typed kind filters.
  Transition choice: 4.x kind names normalize at parse time (not only in
  Walk), so 4.x trees and `keystone new <old-kind>` keep working; `concern`
  lints as a warning naming `keystone migrate`. Permanent, per migrations
  iron law 6 (see the Phase 7 correction). Skill `triggers:` is no longer
  required (description is the activation contract). Deferred: hook `type:`
  (command/agent/prompt) to Phase 5, where projection first needs it.
- Phase 2a: done. `.agents/` (primitives) + `.keystone/` (INDEX, lockfile,
  learning, work, archive, state, sources, hosts, policies, snapshots);
  scaffold templates restructured to `agents/` + `keystone/`; `InferKind`
  reads the first segment under the walk root (works for policy trees);
  dead `charterRoot` params removed from the policies and lockfile APIs;
  `v5_0` layout migration + tests; this repo migrated (no `.charter/`).
  CodeScene: `v5_0.go` String-Heavy / Primitive Obsession accepted (path-moving
  code; same acceptance as the 4.0 CLI path code); every other file stable or
  improved.
- Phase 2b: done. Spec fields parse (`name` → ID, `allowed-tools` → Tools,
  `metadata.model/tags` → Model/Tags), so spec-shaped and 4.x skills yield
  the same descriptor. Lint: spec violations are errors (name charset/length,
  name ≠ dir, description > 1024, compatibility > 500); missing `name` and
  non-spec keys warn. `keystone lint --fix` rewrites project skills to spec
  shape (triggers folded into description, or metadata when over 1024) and
  inlines retired concerns into any `includes:` consumer; bodies untouched,
  every rewrite printed. Dogfood: 20 skills + 4 subagents fixed, both
  concerns deleted, lint at 0 warnings; scaffold templates and `keystone new
  skill` emit spec shape. Found for Phase 4: the Claude skill projection
  writes `tools:`, but Claude Code's skill key is `allowed-tools`.
- Phase 2c: done. `docs/schemas/primitive.schema.json` — one `$defs`
  entry per kind (picked by directory; frontmatter alone cannot tell kinds
  apart). All 199 primitives in `.agents/` and the scaffold templates
  validate. Found while validating: YAML 1.1 parsers read the 4.x key `on:`
  as boolean `true`, and `run: true` as a boolean — so lint now warns on
  every 4.x spelling (`on:`, `globs:`, 4.x kind names) and `lint --fix`
  rewrites them in place (key order and quoting kept). Dogfood + templates
  fixed (73 + 64 files, at most 2 lines each); 8 placeholder hooks quoted
  `run: "true"`.
- **Phase 2 complete.**
- Phase 3: done. `internal/framework/projection`: pure `Plan` over host
  `Profile`s (per-kind `Rule` with strategy + dest + link target, and the
  dirs each host already `Reads`), `Apply`, `Check`. A host that already
  reads a primitive's location gets no output; a host that sees one
  primitive at two paths gets a warning. Ownership recorded in lockfile
  `projections`; never overwrites an unowned path; prunes only owned paths
  (guarded: local, outside `.agents/` and `.keystone/`). First 5.0 run
  adopts 4.x `keystone-*` entries so stale ones prune. `link_mode`
  (`symlink` | `copy`, copy default on Windows). `keystone verify` fails on
  projection drift. Claude Code profile: skills symlinked per entry, rules /
  commands / subagents / inferential hooks shimmed (hosts.md corrected).
  Dogfood: `.claude/skills/*` are now links; stale `keystone-task` pruned.
  User skills now project under their own name (`/deploy`, not
  `/keystone-deploy`) — CHANGELOG item. The Phase 2 `tools:` vs
  `allowed-tools` finding is moot for skills (links carry spec frontmatter).
  Known gap, not fixed here: `docs/schemas/keystone.json.schema.json` sets
  `additionalProperties: false` but lacks `adapters` and `signals`.
  CodeScene: String-Heavy / Primitive Obsession in `projection` accepted
  (path code, as in Phase 2); `verify`/`lint`/`project` CLI got simpler via
  a shared `parseFlags`. Claude Code picked up the linked skills live in the
  session that built this (symlinked skill dirs confirmed working).
- Phase 4a: done (dogfooded with `project --overwrite`, user-approved). Host profiles for all 10 hosts
  in `internal/framework/projection/hosts.go`, selected by keystone.json
  `adapters` (claude-code always). Per-host shim renderers (yaml.v3
  frontmatter; Codex TOML). New `copy` strategy for Continue and Cline
  skills. Old cursor/continue/opencode adapters retired from live projection
  (cursor/continueide kept FROZEN for the v2.x migrations; opencode
  deleted). Fixed: Claude Code rule shims emitted `globs:`, which Claude
  Code ignores, so path-scoped rules loaded every turn — now `paths:`.
  User rule (2026-11-05): never overwrite without `--overwrite`. So Apply no
  longer auto-adopts 4.x files; `keystone project --overwrite` is the only
  way to replace an unowned path, and the `v5_0` migration records 4.x
  `keystone-*` projections as owned (explicit `keystone migrate`).
- Path sweep (user-requested, 2026-11-05): every live reference to the 4.x
  `.charter/` layout and the 1.x `charter/` prefix now points at `.agents/`
  or `.keystone/` (old dir names mapped: actions→commands, personas→agents,
  guides→rules, corpus→references, sensors→hooks, playbooks→skills,
  adapters→hosts). 234 files: charter content, scaffold templates, host
  docs, MCP prompts, README/CLAUDE/CONTRIBUTING/GLOSSARY/GOVERNANCE,
  docs/conventions, docs/ports. 278 relative links repaired in installed
  coordinates. Left on purpose: CHANGELOG, ADRs, old plans (history), frozen
  migrations. Still broken (pre-existing, not folder moves — Phase 8):
  links to missing `principles/`, `policies/universal`, `policy-audit.md`,
  `rules/process/sensors.md`, and keystone's own `docs/ports` from templates.
  Also fixed: `Apply` preflight (a refused plan writes nothing), duplicate
  warnings aggregated per host + kind, empty per-skill dirs pruned,
  INSTALL_PROFILE writes `kind: reference` + a working lockfile link.
  Open: Cursor sees each skill at `.agents/skills` and `.claude/skills`
  (dedupe unverified, Phase 9 smoke); command `audit` and skill
  `keystone:audit` share the projected name `keystone-audit` (pre-existing
  in 4.x; Claude Code lets the skill win) — Phase 6 renaming.
- Phase 4b: done. `AGENTS.md` is the rendered entrypoint, written as a
  managed block (`<!-- keystone:start/end -->`): replaced in place, or
  inserted after the first H1 with every other byte kept, so a user's own
  AGENTS.md is never overwritten. It carries constraints / capabilities,
  rules that apply everywhere (one line each; bodies on demand), iron laws,
  lifecycle, override, and an "On each host" line per enabled host (what it
  loads itself vs. what to open from `.agents/`). `CLAUDE.md` gets a block
  with `@AGENTS.md`; `.aider.conf.yml` (owned whole file) reads AGENTS.md +
  INDEX.lite. `CHARTER.md` and Aider `CONVENTIONS.md` are no longer written;
  the `v5_0` migration removes them (and a pointer-only AGENTS.md) only when
  their content has the exact 4.x generated prefix. `verify` reports a stale
  or edited managed block. 809 words always-loaded for this repo.
- Phase 4c: done. All `targets/` pointer templates and the menu installer
  deleted (every host reads AGENTS.md itself; CLAUDE.md and .aider.conf.yml
  come from projection). Fixed a 4.x gap: `init --agent X` never wrote
  keystone.json `adapters`, so only Claude Code was ever projected; now init
  writes them and `keystone target add` appends and re-projects. Frozen:
  `adapters/agnostic`, `aider` (v2.x migrations only).
  Phase 8 note: README.md still describes CHARTER.md.
- **Phase 4 complete.**
- Phase 5a: done. Security fix (own commit): a failing hook on a host event
  exited 1, which Claude Code / Codex / Cursor treat as non-blocking, so
  PreToolUse checks never blocked; it now exits 2, and the host's event JSON
  reaches each hook on stdin (`KEYSTONE_HOST` names the payload shape).
  Hook bridges on the engine: Claude Code (merge into settings.json), Cursor
  (merge into .cursor/hooks.json, `failClosed: true`), Copilot (owned
  .github/hooks/keystone.json), opencode (owned plugin; Bun `$` throws on
  non-zero, which blocks). New `merge` strategy for shared config files:
  keystone entries found by their `keystone hook fire` command, user entries
  kept, never owned or pruned; drift checked by re-merging. Fixed: the
  Claude Code merge replaced a whole event's hook list, deleting user hooks
  on the same event. Codex / Goose / Pi (bridge formats unverified) and
  Cline / Continue / Aider (no hook system) are reported by `project` and in
  AGENTS.md ("does not run hooks — run the checks yourself"). Warnings: an
  event a host has no equivalent for; Copilot may double-fire hooks it also
  reads from .claude/settings.json. Removing the last host-event hook strips
  keystone's entries from shared files. CodeScene: String-Heavy in
  projection/hooks.go accepted (event and path names).
  Open: Codex bridge once its hooks.json shape is verified; payload
  passing to the opencode plugin.
- Phase 5b: done. Canonical permission rules (Claude
  syntax) translated per host. Claude Code: identical, merged into
  settings.json. Cursor: `.cursor/cli.json` allow/deny — `Bash` only for a
  single command word (`Shell(cmd)`), `Read`, `Edit|Write → Write`,
  `WebFetch(domain:x)`; no `ask` in Cursor, so ask rules are reported.
  Codex: owned `.codex/rules/keystone.rules` prefix rules; an exact-command
  allow is not translated (a prefix would broaden it), an exact deny/ask
  may be (stricter). opencode: reported for now — `permission.bash` is
  last-match-wins, so a correct merge must preserve JSON key order; pending
  an order-preserving merge. Copilot / Goose / Pi / Cline / Continue / Aider:
  reported (no committed permission config). Every untranslated rule is
  listed by `keystone project` and counted in AGENTS.md per host — never
  dropped silently. Shared-file merges record keystone's own entries in the
  lockfile (`managed_entries`): a rule dropped from the charter is removed,
  a user's identical entry is kept, and a no-op merge never rewrites the
  file. The old Claude Code posture projection was additive only (a removed
  allow lingered); it is replaced. Caught by tests: Claude Code and Cursor
  permission bridges were not wired (a gofmt-alignment miss), which would
  have reported every rule instead of enforcing it — now guarded by a test.
- Phase 5c: done. A security review of the Phase 5 diff found 12 issues;
  each one was checked against the code. Fixed:
  - `hook fire` failed open on any error before the hooks ran. A bad flag,
    an older binary that rejects `--host`, and a walk error all exited 1,
    which hosts treat as non-blocking. Every error on a host phase now
    exits 2. The Claude Code and Copilot bridges add `|| exit 2`, so a
    missing binary also blocks.
  - `hook fire` read the charter relative to cwd, so a bridge run from a
    subdirectory found no hooks and passed. It now searches upward for
    `.agents/`. A host phase with no charter exits 2.
  - Host-phase hooks had no timeout. A hung check reached the host's
    timeout, which the host treats as non-blocking. They now stop at 50s
    and count as a failure.
  - Cursor: an exact allow `Bash(ls)` became `Shell(ls)`, which allows any
    arguments. Only a `cmd:*` allow translates now.
  - A wildcard in the middle of a spec (`Bash(curl *)`) became a literal
    `*` token, so the Codex deny matched nothing. Such rules are now
    reported on every host.
  - `Read` and `Write` paths with Claude's `/`, `//`, and `~/` anchors
    were passed to Cursor unchanged. They are now reported.
  - Managed entries were lost when apply failed after a merge, and a host
    removed from `adapters` kept keystone's allows forever. The record is
    now saved on the error path, and a disabled host's recorded entries
    are stripped.
  - The opencode plugin now passes the event input on stdin.
  - Prune refuses a path whose parent resolves outside the project
    through a symlink.
  - Lint warns when an inferential hook binds a host event, because it
    cannot block there.
  Accepted: a user's own entry that starts with `keystone hook fire ` in
  `.cursor/hooks.json` is treated as keystone's. A host removed from
  `adapters` keeps keystone's hook-bridge entries in its shared file;
  that keeps checks running, so it fails safe. Unverified, listed in
  docs/hosts.md: Cursor `failClosed`, exit-2, and stdout semantics; Cursor
  path globs; the Copilot deny-on-non-zero claim.
- Phase 6.1: done. `keystone new` takes the 10 kinds (constraints, then
  capabilities) plus `adapter` and `policy`; the 4.x names (`guide`,
  `corpus`, `sensor`, `posture`, `agent`, `playbook`, `pattern`) exit
  non-zero and name the new one. New `keystone new hook`. Scaffolds write
  5.0 kinds, and a rule's `corpus:` entry resolves to its paired reference.
  Fixed a Phase 1 miss: `corpus:` refs resolved only under the 4.x
  `corpus/` kind, so they never resolved against `reference`. `init
  --reset` (which ran `RemoveAll` on `.agents/`) is replaced by
  `init --overwrite`: keystone's scaffold and host files are replaced in
  place, and nothing is deleted. The 6 stale `keystone-new-*` scaffold
  skills (guide, corpus, sensor, playbook, command, action) are one
  `keystone-new` skill.
  For Phase 7: existing installs keep the removed `keystone-new-*` skill
  folders (the migration should remove them when unmodified). Wire the
  `cli` and `plugin` tool transports (today only `mcp` is served).
  For Phase 8: `docs/conventions.md` still names `new guide|sensor|playbook`.
- Phase 6.2: done. `keystone list` and `keystone charter show` print
  constraints first, under a heading per category, kinds in canonical
  order; `keystone list constraints|capabilities` filters by category.
  `show` and `explain` print the category; `show --json` gains `category`
  and renames `globs` to `paths`. Coverage and conformance say rule,
  reference, and `paths:` instead of guide, corpus, and globs.
- Phase 6.3: done. MCP: `keystone_get_corpus` is `keystone_get_reference`
  (default source kind `rule`); `keystone_new_*` tools are the 5.0 kinds
  (the 4.x verbs failed once the CLI refused them); `keystone_list_primitives`
  takes `category` and `paths` (`glob` still accepted); the server
  instructions describe the two categories; server version 3.0.0. Fixed:
  MCP resolved a `corpus:` entry only under a `corpus/` id, so a 5.0
  reference id never resolved. Dashboard: the nav splits the charter into
  constraints and capabilities; the primitives page filters by category
  and `paths`; the new-primitive form offers the 10 kinds (it offered
  guide/corpus/sensor/action/playbook, which all failed) and its redirect
  went to a 404 route. A test GETs every nav link. The runtime-resolution
  rule names `keystone_get_reference`; this repo's copy also lost a stale
  stage-3 pointer to the removed `keystone_source_query`.
- Phase 6b.1: done. `keystone allocate [--apply]` (package
  `internal/framework/allocate`) reads every host-specific setup — `.claude/`
  skills, agents, commands, rules, and `settings.json` permissions and
  matcher-free command hooks; `.cursor/rules/*.mdc` (agent-requested →
  skill), `.cursor/skills|agents|commands`, `.cursorrules`; Copilot
  instructions, prompts, agents; `.clinerules` (file or folder),
  `.windsurfrules`; `.continue/rules|prompts`; `.opencode/agent(s)|command(s)`.
  Keystone's own projections (lockfile `projections`, symlinks) and its
  managed settings entries are skipped. Frontmatter: `globs:`/`applyTo:` →
  `paths:` (comma strings → lists), `alwaysApply` dropped (true → no paths),
  `tools:` strings → lists, `id:` and `description:` added when missing
  (each change named in the report). `--apply` writes new files under
  `.agents/` only; it never overwrites one and never changes an original.
  Known gaps, reported per item: a subagent with no `tools:` (it had every
  tool on its host) needs a list; a Claude hook with a matcher stays in
  settings.json; a moved Claude hook now blocks on any non-zero exit.
  Fixture test covers every source family and the round trip.
- Phase 6b.2: done. `keystone allocate` also reads the root files
  (`CLAUDE.md`, `AGENTS.md`, `GEMINI.md`, `.goosehints`), splits them at
  `## `, and proposes moving each section that need not load every session:
  a backticked file-extension pattern (`*.go`, `src/**/*.tsx`) → path-scoped
  rule; 3+ numbered steps → command; a line pairing "before you commit /
  always run" with a backticked command → hook on Stop; a "why/background/
  architecture" heading or 300+ words → reference. Iron-law, never, and
  security headings stay, as do the preamble and keystone's managed block.
  The report gives each root file's always-loaded size and the tokens each
  move saves. Two false positives from this repo's CLAUDE.md shaped the
  heuristics: the cue and command must share a line, and folder patterns
  (`.claude/*`) do not scope a rule.
- Phase 6b.3: done. `keystone allocate --overwrite` applies, then hands
  over the originals whose converted file was written: host files become
  keystone-owned (lockfile `projections`; for a skill, its whole folder),
  moved sections leave `CLAUDE.md` / `AGENTS.md`, moved hooks leave
  `settings.json`, and moved permission rules become keystone-managed
  entries. It then re-indexes and re-projects. Guards: an original whose
  `.agents/` target already existed is never handed over (its content
  would be lost), and neither is one whose host is not in `adapters`
  (an owned path no projection covers is pruned); both are reported with
  the fix. `GEMINI.md` and `.goosehints` sections are never removed (no
  host projection replaces them). A `keystone-allocate` skill has an agent
  review the classifications and run `--overwrite` only on the user's
  explicit request. Tests assert mappings per source rather than golden
  report files (the report wording is not a contract).
  Open decision: Claude Code projections carry a `keystone-` file and name
  prefix, so a converted `reviewer` subagent becomes `keystone-reviewer`
  and `/release` becomes `/keystone-release`. Ownership now lives in the
  lockfile, so the prefix is no longer needed to mark keystone's files.
- Prefix decision (user-approved): the projection engine drops the 4.x
  `keystone-` prefix. A primitive projects under its own id, kebab-cased
  (`reviewer`, `/release`, `go-stdlib-first.md`); an id that already starts
  with `keystone` keeps it (`keystone-verify`). Ownership lives in the
  lockfile, so the prefix no longer marks keystone's files. The frozen
  `primitive.Project` (2.x migrations) keeps the prefix, so the v5.0
  migration still recognizes the 4.x files it adopts. This repo re-projected
  cleanly: 62 written, 62 old paths pruned, then idempotent.
  For Phase 9: the shipped `review` command now projects to `/review`, which
  I believe collides with Claude Code's built-in `/review`. The shipped
  commands also duplicate the shipped `keystone-*` skills one for one
  (`audit` / `keystone-audit`, …); decide whether to keep both.
- Phase 6c.1: done. `keystone budget [--json]` measures what each host
  loads: always (root files — AGENTS.md everywhere, CLAUDE.md on Claude
  Code, the lite index on Aider — plus each skill, command, and subagent
  listing and each pathless rule the host loads by itself, decided from the
  host profile) and on demand (rule, reference, skill, command, subagent,
  tool, document bodies, per kind and largest first). Caps in
  `keystone.json` `budgets`: `always_loaded`, a kind (`max_tokens` for the
  kind's total, `max_tokens_per_load` per file), or a 4.x key mapped to its
  kind; two keys for one kind keep the stricter. `keystone verify` and
  `keystone budget` fail on an exceeded cap, naming the file and overage
  and pointing at `keystone allocate`. `init` and `doctor` print the
  per-host summary. The 4.x per-port Allocator is removed: on 5.0 it
  counted nothing, because its port names were the 4.x folders. Plan text
  corrected: the estimate is a word count, not bytes/4.
- Phase 6c.2: done. MCP `keystone_budget` returns the report as JSON;
  the dashboard has a Budget page (`/charter/budget`, in the nav) with caps
  exceeded, the always-loaded tier per host, the on-demand tier per kind,
  and the largest files; `keystone allocate` prints each host's
  always-loaded tier against the `always_loaded` cap. The measurement lives
  in `budget.MeasureProject`, shared by the CLI, MCP, and dashboard.
- Phase 7: done.
  - Migration: v5.0 Up moves each unedited 4.x scaffold skill
    (`keystone-new-guide|corpus|sensor|playbook|command|action`; unedited
    = only a SKILL.md whose sha256 matches a shipped version, 22 versions
    from git history) to `.keystone/archive/5.0/skills/`; an edited one
    stays (iron law 2). Down returns them first. Their 4.x `.claude/`
    projections are already recorded as owned, so `keystone project`
    prunes them. Up/down round trip and idempotent re-run are covered.
  - Tools: `keystone tool run|list` and package `toolrun` (shared with the
    MCP server, which now also rejects a missing required argument).
    `cli` and `plugin` tools are listed in AGENTS.md with their
    `keystone tool run` command; a `plugin` tool is also an opencode
    custom tool (`.opencode/tools/<id>.ts`, format from the opencode docs).
    Unverified: whether opencode installs `@opencode-ai/plugin` for
    project tools by itself.
  - Dogfood: `adapters` adds opencode, codex, github-copilot; `project`
    wrote 88 files with no refusal and re-runs unchanged; `verify` clean,
    `lint` clean, `charter conformance` CONFORMANT.
  - Not done here: republishing keystone's own policy repos in 5.0 shape
    (separate repos; 4.x policy trees keep loading through the legacy
    folder and kind mapping).
- Phase 8 (in progress): 8.1 README, GLOSSARY, upgrade guide, CHANGELOG
  done; 8.2 `docs/kinds/` (one page per kind), `docs/internals/`,
  conventions, compatibility, CONTRIBUTING, charter-content rules, hosts
  matrix done. Scaffold READMEs in progress. Open decision: the per-host
  templates (`templates/keystone/hosts/*`, ~16,500 words of 4.x content
  installed into every repo) and `keystone new adapter`.
- Phase 9 bug list (found during Phase 8, each checked in code):
  1. Policy `strict` keys and prefixes in `loader/cascade.go` and
     `manifest/manifest.go` are the 4.x names (`sensors`, `guides`, …): a
     5.0-shaped policy's strict items may not lock. Security-relevant.
  2. Shim bodies keep the source's relative links, so links break in the
     projected copy (`.cursor/skills/*/SKILL.md`, …): ~300 broken links.
  3. `keystone doctor`'s no-`../` link rule predates the two-root layout;
     links between `.agents/` and `.keystone/` need `../` (124 hits).
  4. Sensor-level evals run a hook's first fenced code block, not its
     `run:`, so a hook from `keystone new hook` fails its own eval.
  5. Lint accepts `transport: http`; nothing serves it.
  6. Lint treats a hook with no `mode:` as inferential (needs `returns:`).
  7. `keystone hook fire` reads the lockfile from cwd for its pending-
     migration warning (wrong from a subdirectory).
  8. Shipped `review` command projects to `/review` (likely shadowed by
     Claude Code's built-in); commands duplicate the `keystone-*` skills.
  9. Codex hook bridge; Cursor `failClosed`/exit-2/stdout semantics;
     Copilot deny-on-non-zero; opencode `@opencode-ai/plugin` install.
- Hosts decision (user): per-host docs removed from the scaffold;
  `keystone new adapter <name>` now defines a custom host in
  `.agents/hosts/<name>.json` (native | link | copy | index per kind for
  skill, rule, command, subagent; `dir` local and outside `.agents/` and
  `.keystone/`; built-in names refused). `keystone target add` accepts it.
  No format conversion, hook bridge, or permission translation: reported.
  A copied rule keeps `paths:` frontmatter; whether a custom host honors it
  is the host's business.
- Phase 8 remaining (8.4): the bodies of the shipped commands, hooks, and
  skills (templates and this repo) still use 4.x terms (`corpus/state/`,
  "sensor", "guide"). Scaffold READMEs, docs, and the fresh-install link
  check are clean.
- Phase 8.4: done. The bodies of every shipped command, hook, skill,
  subagent, rule, reference, and optional starter (templates and this repo,
  284 files) use 5.0 terms and paths; ids, `name:`, and `corpus:` entries
  are unchanged (identity). Pointers to the removed per-host docs now point
  to the `AGENTS.md` host lines. 19 links that were broken in 4.x starter
  packs are fixed. A fresh install has 0 broken links; lint and verify are
  clean. 83 remaining hits are ids, book titles, ordinary English, eval
  level values, and historical ledger rows.
- Phase 9 additions from 8.4:
  10. `keystone verify --sensor <id>` keeps the 4.x flag name (hooks run
      `keystone verify --sensor drift`); add `--hook`, keep `--sensor` as an
      alias.
  11. `keystone new policy` writes a `corpus/` folder and a README listing
      `guides/`, `sensors/`, `playbooks/` (pairs with item 1).
  12. `spec-author` writes `.keystone/learning/specs/`; the `spec` command
      says `docs/specs/`. Pick one.
  13. Stale comments: `cmd/keystone/conditional.go:22`,
      `internal/framework/scaffold/embed.go` (`targets/`, `charter/`).
- Phase 9: done. Bug list items 1-13 fixed or decided:
  strict/required locks map to 5.0 folders (a "guides" lock had stopped
  firing); new policies are 5.0-shaped; shims rebase their links (0 broken
  projected links, from ~300); doctor checks that links resolve; hook evals
  run `run:`; `http` tools warn; hooks with `run:` and no `mode:` lint as
  computational; the migration warning finds the project root; `verify
  --hook` (alias `--sensor`); spec path is `docs/specs/`; stale comments.
  User decisions: shipped commands folded into the `keystone-*` skills;
  the fail-closed bridge stays (recovery documented).
  Found during the phase: prune could delete a charter source through a
  symlinked host folder (fixed, regression test); Cursor blocked every
  `preToolUse` on the bridge's plain-text stdout (fixed from Cursor's
  docs); Codex hooks exist (bridge added); a fresh init reported 8 pending
  migrations (fixed). Host behaviors settled from docs, recorded in
  docs/hosts.md; only Claude Code is verified by a real run.
  Smoke test (fresh project, all 10 hosts plus a custom host): init,
  target add, new for every kind, project (idempotent), lint, verify,
  doctor, budget, allocate, hook fire per host, tool list: all pass.
