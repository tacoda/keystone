---
status: approved
captured: 2026-11-21
approved: 2026-11-21
topic: keystone 6.0 — context framework pivot
spec: 2026-11-21-context-framework
---

# Plan — Keystone 6.0: the context framework

Turns spec `2026-11-21-context-framework` into implementation. Keystone
becomes the agent context framework; the **charter** keeps its name and is
redefined as the codified context in the repository — standardized under
`.agents/`, lazy-loaded, queried via CLI and MCP by any coding agent.
Projection stays (owner decision) and gets a minimality pass.

No charter→context renames (decision 11). The Go package, CLI group, and
MCP tool names that say "charter" are already correct. Execution order is
fixed for phases 1–3 (redefinition and kinds before adoption); 4–5 may
interleave; 6–7 land last. Each phase is one implementation pass with its
own verification gate — `go test ./...`, `keystone verify`, and the
phase's success criteria — before the next starts. Nothing here removes a
kept feature (decision 10).

## Phase 1 — Redefine the charter; drop the split; `agent` kind

Same word, new definition — plus the category removal, `subagent`→`agent`,
and retirement of the 4.x aliases. One pass, all surfaces, with migration
`v6_0`.

1. Drop the category system: `Category`, `kindCategory`, `CategoryOf`,
   `CategoryTitle`, `categoryRank`, category fields in `INDEX.json` /
   `INDEX.lite.json`, category sorting, `keystone list` category filters,
   the MCP `listFilter` category param, web filter/nav/templates, and the
   constraint/capability sections of the AGENTS.md template. Consumers
   sort by kind (`SortByCategory` → `SortByKind`).
2. Rename kind `subagent` → `agent` (folder inference from `agents/`
   unchanged; `KindSubagent` → `KindAgent`). Update docs/kinds, `keystone
   show agent <id>`, host-matrix column heading.
3. Retire 4.x legacy aliases (`guide`, `sensor`, `posture`, `playbook`,
   `pattern`, folder `corpus`→ now the corpus kind). `lint --fix` rewrites
   old spellings to 6.0.
4. Rewrite the MCP server `instructions` and CLI help text: the charter is
   the codified context in the repository (the query flow itself lands in
   Phase 4).
5. Migration `v6_0`: rename `subagent` frontmatter, rewrite 4.x spellings,
   strip categories from state files, and **retire `.agents/tools/` trees**
   (archive in place with a README pointer; lint explains the removal).
   Extend the 5.x fixture trees.
6. Remove the `tool` kind entirely (decision 12 — harness machinery, not
   context): `KindTool`, `.agents/tools/` handling, `keystone tool run`
   (`toolrun`), the MCP server's dynamic tool serving, per-host tool
   projections (opencode plugins, `.mcp.json` tool entries), the AGENTS.md
   template's Tools section, `keystone new tool`, the schema `$defs` entry,
   and `docs/kinds/tool.md`. The kind set drops to nine.
7. Docs sweep: README, GLOSSARY, CONTRIBUTING, docs/kinds/README, hosts.md
   header, AGENTS.md/CLAUDE.md templates, install messages — every
   definition of "charter" becomes the codified context; constraints /
   capabilities prose comes out; the inferential/deterministic boundary
   gets stated where kinds are described.

**Success criteria:** `rg -i 'constraint|capability'` over the tree
returns only CHANGELOG history and migration compat notes; `INDEX.lite.json`
has kind/id/description only; `keystone show agent <id>` resolves a renamed
subagent; a 5.x fixture migrates to a green 6.0 tree via `keystone migrate`
+ `lint --fix`; docs define the charter as codified context.

## Phase 2 — The corpus kind

Standalone information: the repo's knowledge, not tied to any rule.

1. Add `KindCorpus` + `.agents/corpus/` to kind tables, folder inference,
   lint rules, and `canonicalDirKind`.
2. Frontmatter: `id`, `description`, `tags`, optional `paths:` (scope to
   regions). Nothing else — decision 9 defers richer fields.
3. Wire into index (`by_kind`), search, `show`, budget (on-demand tier),
   and `docs/schemas/primitive.schema.json` (`$defs` entry).
4. Write `docs/kinds/corpus.md` with the one-line distinction: **corpus is
   standalone information; reference is the why behind a rule.**

**Success criteria:** `keystone new corpus <id>` → index → `show` →
`search` hit round-trips; budget counts it on demand; schema validates;
lint rejects a corpus entry without `description`.

## Phase 3 — Repository adoption (`charter.roots`)

The boundary is the repository (decision 7). Repo docs become queryable
corpus entries in place.

1. `keystone.json` gains `charter.roots`: an array of globs (default
   `["docs/**/*.md", "*.md"]` on fresh init, excluding `.agents/`,
   `.keystone/`, `CHANGELOG.md`? — decide in pass; exclusions configurable).
2. Indexer walks the globs, indexes each file as a corpus primitive with
   `provenance: repo` (authored entries stay `provenance: agents`),
   `path` pointing at the original file.
3. Description auto-derived (first heading or first paragraph); never
   modifies the adopted file.
4. `get`/`show` resolve adopted entries to the original path; search and
   the query surface (Phase 4) include them.

**Success criteria:** with this repo configured, `keystone search glossary`
returns `GLOSSARY.md` and `show` prints it, file unmoved and unedited;
INDEX entries carry `provenance: repo`; lint does not demand frontmatter in
adopted files; `keystone index` is idempotent over adopted sets.

## Phase 4 — Invocation surfaces (the front door)

Availability over adoption (decision 15): everything keystone can do is
invocable by users, scripts, rules, skills, and hosts; nothing is imposed.

1. CLI `keystone charter query [files…|<task>]`: composes rules matched by
   `paths:`, corpus entries matched by `paths:`, adopted docs matched by
   region, and references cited by those rules — descriptors first, bodies
   on `--full`.
2. MCP `keystone_charter_query` with the same composition; results carry
   `path` so the agent opens bodies only when needed.
3. Every capability gets an invocation on each surface — CLI verb and MCP
   tool, not only a planted skill. The learning and pruning flywheels are
   the templates: `keystone learn` / `keystone_learn`, `keystone
   synthesize` / `keystone_synthesize`, `keystone audit`, `keystone
   check-drift`, `keystone mode`. Skills remain wrappers for teams that
   want them planted, never the only path.
4. Server instructions lead with the query and invocation flow;
   `keystone_explain` stays for single-primitive why.
5. init stops planting the lifecycle skills and phase rules by default —
   a fresh project gets the minimal charter and full invocability; the
   lifecycle moves behind an opt-in `--starter workflow` pack.

**Success criteria:** on this repo, querying `internal/framework/projection/`
returns the path-scoped rules plus any scoped corpus, in descriptor form;
an MCP-only client (existing `server_test.go` harness) completes a
cold-start question/answer with no host files present; the flywheel
capabilities run through CLI and MCP with no lifecycle skill planted
(delete `task` + the keystone-* lifecycle skills from a scratch init and
exercise the verbs).

## Phase 5 — Context drift management (pillar 2)

The charter is living context; keystone notices when it rots. Today drift
is a hook in this repo's own charter (`drift` — diff vs loaded rules) plus
`risk-fingerprint`. The framework itself offers nothing. Design details get
a mini-spec when this phase opens; the shape:

1. `keystone drift [files…|--diff]` — report which relevant primitives
   (rules by `paths:`, corpus by region) no longer match the code they
   scope: symbols/paths they name are gone, guidance contradicts the
   diff, body references dead files.
2. Staleness as data: a `verified:` date on primitives (author-set or
   stamped by drift runs), surfaced in `show`, INDEX, and the dashboard.
3. `keystone verify` reports stale-but-in-scope primitives; the dashboard
   gets a drift pane fed by the same data.
4. The repo's own drift hook keeps working; the framework feature
   subsumes its mechanics, not its judgment.

**Success criteria:** on this repo, `keystone drift` names at least one
genuinely stale primitive against a synthetic diff (fixture-proven), and
`verify` surfaces it; a `verified:` date round-trips through index/show.

## Phase 6 — Lazy-load hardening

1. AGENTS.md managed-block template shrinks — keeping the `# Charter`
   heading, redefined — to: what the charter is (the codified context in
   this repository), how to query it (CLI one-liners + MCP tool names),
   the index pointer, and the per-host line. No inlined rule bodies.
2. Regenerate this repo's AGENTS.md from the new template; move any
   inlined always-on content into `.agents/` primitives (or corpus) so
   nothing is lost.
3. Budget: default `always_loaded` cap applies; `verify` fails overages
   with file + overage named.

**Success criteria:** this repo's always-loaded tier ≤ its cap; the
generated block contains no primitive body text; `keystone verify` green.

## Phase 7 — Projection demotion (kept, minimal and clean)

Owner decision (2026-11-21): the compatibility layer stays — native
activation and host-side enforcement are required. It becomes opt-in and
gets a minimality pass.

1. `keystone init` writes an empty `adapters` list; `project` requires
   opt-in (`keystone target add …` or an explicit flag) and says so.
2. README/docs re-order: query layer first, projection under "compat";
   `hosts.md` header gains a compat notice and keeps its re-verify dates.
3. `doctor` reports configured adapters as opt-in state, not default.
4. Minimality pass over the projection design: collapse redundant
   strategies, remove dead options, simplify the adapter surface — same
   behavior, less machinery. Individually reviewable steps, green after
   each.

**Success criteria:** fresh `keystone init` in a temp dir writes nothing
outside `.agents/`, `.keystone/`, `keystone.json`, and the AGENTS.md block
(verified by filesystem diff); opt-in `project` for claude-code still
passes with a green lockfile.

## Phase 8 — Dogfood, docs, release

1. Convert this repo's own tree: `agent` kind renames, corpus entries,
   `charter.roots` adopting `GLOSSARY.md` and `docs/**`. This repo's own
   phased lifecycle (task, spec, orient, …) becomes the **worked example**
   of an opted-in workflow — init no longer plants it by default; a
   `--starter workflow` pack (alongside `universal-principles`) carries
   the lifecycle skills for teams who want them.
2. Rewrite README/GLOSSARY/CONTRIBUTING to the context framing; CHANGELOG.
3. Release 6.0.0 per the release checklist; update the gh-pages site.

**Success criteria:** release checklist green; this repo migrates itself
and `keystone verify` passes on its own tree; site leads with the context
framing.

## Risks

- **Rename regressions.** Mitigation: fixture trees under migrations, full
  suite per phase, `rg` sweep in Phase 1 criteria.
- **corpus/reference confusion.** Mitigation: the one-line distinction in
  both kind docs and the index template (Phase 2.4).
- **Redefinition drift.** Old "constraints and capabilities" definitions of
  charter linger in prose. Mitigation: Phase 1 docs sweep + success
  criterion (`rg -i 'constraint|capability'`).
- **Host-matrix rot after demotion.** Mitigation: compat label + existing
  re-verify discipline in `hosts.md` (Phase 7.2).
- **Adoption glob over-reach** (indexing build junk). Mitigation:
  exclusions default-on for `.keystone/`, `.git/`, vendor dirs; idempotent
  index over adopted sets (Phase 3).
- **Drift over-design.** Drift detection can balloon into a static analyzer.
  Mitigation: Phase 5 opens with a mini-spec and a hard scope line —
  keystone names stale context; it does not refactor code.

## Deferred (post-6.0)

- Frontmatter feature catalogue and additions (decision 9): a dedicated
  spec; bar = easy to use and understand, framework-like extensibility,
  enable-and-get-out-of-the-way.
- Projection removal decision (decision 8).
- Semantic search.
