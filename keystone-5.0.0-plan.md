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

### Phase 7: Migration and dogfood

- `internal/framework/migrations/v5_0.go`:
  - move `.charter/<kind>/` to `.agents/<new-kind>/` and keystone state to
    `.keystone/`;
  - rewrite frontmatter (frontmatter only, never bodies; the 4.0 lesson);
  - fold playbooks and patterns into skills, inline the two concerns, split
    sensors into hooks;
  - fold `triggers:` into `description`. This is lossy, so the migration writes
    `.keystone/migration-5.0-report.md` listing every rewritten description for
    human review;
  - `CHARTER.md` content merges into the new `AGENTS.md`; old projected files
    under `.claude/` etc. are replaced by links and shims.
  - `Down` reverses the moves; documented as best-effort for the lossy folds.
- Installed policies: bump the policy manifest schema. The resolver refuses a
  4.x-shaped policy with a clear message. Republish keystone's own policies in
  5.0 shape.
- Dogfood: migrate this repo, add `opencode` (and `codex`, `github-copilot`)
  to `keystone.json` `adapters`, `keystone index && keystone project`, commit.
- Verify: up/down round-trip test on a 4.x fixture; idempotent re-run test;
  `keystone verify` clean; `keystone charter conformance` CONFORMANT.

### Phase 8: Docs

- `README.md`, `GLOSSARY.md` (constraints vs capabilities, kind table, the
  "agent vs subagent" note), `CONVENTIONS.md`, `CONTRIBUTING.md`, `docs/**`,
  `docs/hosts.md`, and every rule under `.agents/rules/charter-content/`.
- `docs/upgrade-4.x-to-5.0.md`: what moves, what folds, what to review in the
  migration report, how to choose `linkMode`.
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
  Walk), so 4.x trees and `keystone new <old-kind>` keep working until the
  Phase 7 migration; `concern` lints as a warning naming `keystone migrate`.
  Phase 7 turns legacy names into a hard error. Skill `triggers:` is no longer
  required (description is the activation contract). Deferred: hook `type:`
  (command/agent/prompt) to Phase 5, where projection first needs it.
