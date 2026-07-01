# Keystone Charter Rebrand — Plan (v4.0.0)

**Positioning:** Keystone is **the coding-agent charter manager** — *constraint
engineering at the repository level*. It is **not a harness.** A harness is the
engine (Claude Code, the orchestrator, the runner). Keystone manages the
**charter**: the authored standards that constrain whatever harness runs, so
each unique repo gets reliable, quality agent output.

**Scope (locked): C — full rename.** Positioning + user-facing surface +
internal Go identifiers. Breaking (`.harness/` dir moves) → ships **4.0.0** with
a migration.

**Context:** Extends `~/tacoda/all-the-things/keystone-rebrand.md` (the AAIF
submission play — rebrand as the charter manager, ship the lexicon as the
standard, then submit). This doc is the *execution* plan for the rebrand itself;
AAIF submission + governance scaffolding are downstream (see Appendix).

Supersedes: `parento-harness-gaps.md`, `keystone-loading-gaps-plan.md` (neither
was a rebrand plan).

---

## The definition (load-bearing)

Split everything around the agent by **one question: did you _author_ it to
specify behavior, or is it the _engine_ that applies what's specified?**

- **Charter** — everything you author to specify how the agent behaves: CLAUDE.md,
  guides/rules, corpus, sensors, hooks, skills, playbooks, personas, policy.
  Mostly prose, sometimes code. *You write it.*
- **Harness** — the engine that interprets and applies the charter against the
  model: the coding agent (Claude Code), frameworks (LangChain, CrewAI), the
  orchestrator, the loop, the runner that fires your hooks.

**The test:** authorship, not executability. A pre-commit hook *executes* and is
still charter, because you authored it to state a standard; the runner that
fires it is the harness. **Author the spec → charter. Be the engine → harness.**

This resolves the rename cleanly: `.harness/` holds *authored primitives* →
it's charter → `.charter/`. The keystone binary/MCP/dashboard are machinery that
*applies* the charter — but Keystone's job is to **manage** the charter, hence
"the charter manager."

---

## Naming map

Rule: **"harness" survives only where it means the agent/engine.** Everywhere it
means *Keystone's authored artifact*, it becomes **charter**.

| Old | New |
| --- | --- |
| `.harness/` directory | `.charter/` |
| "the harness" (authored primitive set) | "the charter" |
| `HARNESS.md` | `CHARTER.md` |
| `keystone harness bootstrap` (CLI) | `keystone charter bootstrap` |
| `keystone_harness_bootstrap` (MCP tool) | `keystone_charter_bootstrap` |
| `--harness-root` flag | `--charter-root` |
| `DefaultHarnessRoot`, `harnessRoot`, `HarnessRoot` (Go) | `DefaultCharterRoot`, `charterRoot`, `CharterRoot` |
| `harness-content` guides/idioms | `charter-content` |
| `harness-debt` sensor | `charter-debt` |
| MCP tagline "the harness, served" | "the charter, served" |
| README tagline "the agent harness framework" | "the coding-agent charter manager" |
| **STAYS "harness":** "coding agents and other harnesses", "the harness runs/applies", host-engine references | unchanged |

Unchanged (product, not artifact): **Keystone**, `keystone` binary,
`keystone.json`, `KeystoneDir`. Deprecated `harness_root` config field stays
readable for back-compat.

---

## Deliverables

- **Rebranded docs** — README/CHANGELOG/CONTRIBUTING/CONVENTIONS/`docs/**`/AGENTS.md/CLAUDE.md lead with "charter manager" + the authorship test.
- **`GLOSSARY.md` / lexicon** — ships the harness↔charter standard *inside the repo*: the two boxes + the authorship test. This is the disambiguation, hosted (AAIF ask #2).
- **Code rename** — full `.harness`→`.charter` across Go, flags, MCP, adapters, templates.
- **`v4_0` migration** — moves installed bases forward.
- **4.0.0 release** — tag push → GoReleaser.
- **gh-pages site** — rebranded, post-release.

---

## Status
- **Phase 1 — DONE.** README, GLOSSARY.md (new), CLAUDE.md, CONTRIBUTING.md, docs/** swept.
- **Phase 2 — DONE.** Full code rename; `go build/vet/test` all green; gofmt clean. 266 files renamed, 40 modified. Fixed along the way: frozen migrations to literal historical paths, restored migrate.go legacy detection, topics.go classifier ordering bug (lockfile under `.charter/`), stale `.keystone` runtime refs (mcp INDEX path, audit dir, snapshot target, watcher), `context.json` removed, `.charter-snapshots` rename.
- **Phase 2b — DONE.** Canonical root `CHARTER.md` (agnostic.CharterBody/ProjectCharterMD) + `HostProfile` + `RenderPointer` (per-host capability delta). project.go/watch.go/init.go emit CHARTER.md + thin AGENTS.md/aider pointers; cursor+continue emit always-apply charter-pointer rules; all 11 `targets/*` seed templates converted to thin pointers (claude-code uses `@CHARTER.md` import). Tests added; build/vet/test green. Decision: skills/commands/agents still project natively (unchanged); vocab rename dropped (kept community names).
- **Phase 3 / F / 4 / 5 / 6 — pending.**

## Phases

### Phase 1 — Docs + lexicon + positioning (user's "start with README")
Prose only. README, CHANGELOG, CONTRIBUTING, CONVENTIONS, `docs/**`, AGENTS.md,
CLAUDE.md. Lead positioning with "the coding-agent charter manager / constraint
engineering." Write **`GLOSSARY.md`** defining charter, harness, the authorship
test. Apply naming map to prose; keep "harness" where it means the engine.
- Verify: remaining `rg -i harness` hits are all engine-meaning (manual read).

### Phase 2 — Code rename + consistency (user's "pass through code")
Go source, comments, string literals, CLI flags, MCP tool/prompt names,
adapters, scaffold templates (`universal-principles/harness/`→`.../charter/`),
config consts (`DefaultCharterRoot`, `.charter`). Also rename in lockstep:
`--harness-root`→`--charter-root`, `harness-debt`→`charter-debt` (sensor +
ledger + all refs), and the CLAUDE.md/AGENTS.md/CONVENTIONS.md generated
templates + adapters. Note `AGENTS.md` template carries a **stale
`.keystone/harness/`** path — fix to `.charter/`.
- **Also: remove `context.json`** (vestigial — no read/write path; `primitive.go`
  already flags it for removal). Drop refs; `v4_0` deletes it. Confirmed with user.
- Verify each step: `go build ./... && go vet ./... && go test ./...` green.

### Phase 2b — CHARTER.md entrypoint + thin adapters (new; user-requested)
Collapse the orientation content currently **duplicated** across the CLAUDE.md
managed block, `agnostic/agents_md.go` (AGENTS.md), `aider/conventions.go`,
cursor rules, etc. into a **single canonical `CHARTER.md` at the repo root**,
rendered by `keystone project`.
- New **`HostProfile`** per adapter: capability flags `subagents`,
  `slashCommands`, `skillsAutoActivate`, `hooks`.
- Each adapter emits a **thin, fully-pointer host file** (decision 3b): no
  duplicated orientation. Claude Code uses native `@CHARTER.md` import where
  supported (decision 2); others emit an **imperative** load instruction —
  "You MUST read `./CHARTER.md` now; it carries the iron laws + ambient rules
  governing this repo" — so ambient rules still apply on non-import hosts.
- Host file also carries the **capability delta** from its `HostProfile`
  (e.g. Claude Code: subagents via Task tool + `/keystone-*`; Cursor/Aider/
  Continue: no subagents).
- **Unchanged: primitive projection.** Skills → `.claude/skills/`, commands →
  `.claude/commands/`, agents/personas → `.claude/agents/` (and each host's
  equivalents) still project as separate native files the host auto-discovers.
  Going thin touches ONLY the orientation file. CHARTER.md surfaces these
  (activation table + INDEX pointer); a host's delta only *scopes* which kinds
  it supports — it never drops the projected files.
- Verify: `keystone project` on this repo produces root `CHARTER.md` + thin
  host files; `go test ./...` green.

### Phase 3 — Migration + dogfood
- New `internal/framework/migrations/v4_0.go`: `.harness/`→`.charter/`,
  `HARNESS.md`→`CHARTER.md`, rewrite `keystone.json` refs, re-emit INDEX.
  `Down` reverses.
- Migrate this repo's own charter; regen INDEX + `.claude/` via
  `keystone index && keystone project`.
- Verify: `keystone verify` clean; migration up/down round-trips in test.

### Phase 4 — Bug-hunt + full verify
Adversarial pass for rename fallout: stale path strings, broken `go:embed`, dead
`.harness` literals, adapter output paths, doc links.
- Verify: `go test ./...`, `keystone verify`, `keystone index` diff sane,
  `pre_commit_code_health_safeguard` / `analyze_change_set` no regression.

### Phase F — 4.0 charter features (all four, user-approved)
Additive, each its own slice. Land after the rebrand core (2b/3), before release.
- **F1 — Charter coverage** (`keystone charter coverage`): report files/dirs no guide governs ("uncharted territory"). Reuses glob machinery. + dashboard page.
- **F2 — Effective charter view** (`keystone charter show --effective`): materialize the fully-resolved post-cascade charter (project + nested policies + strict). Surfaces existing loader output. + dashboard page.
- **F3 — Charter amendments + provenance**: first-class human-readable record of charter evolution + who authored/ratified each fragment, hash-pinned. Ties to learning flywheel + AAIF GOVERNANCE story.
- **F4 — Charter conformance score**: fold evals + quality-radar into one charter-adherence score. Define criteria sharply first (avoid vanity metric).
- Each: tests + `keystone verify` clean + code-health no regression.

### Phase 5 — Release 4.0.0
Tag push → GoReleaser. **No `gh release create`.** CHANGELOG 4.0.0 entry: breaking
`.harness`→`.charter` move + migration note.

### Phase 6 — gh-pages site
Separate `gh-pages` branch. Rebrand site copy to charter manager / constraint
engineering + the two boxes. After the release so version refs match. Live
publish — confirm before push.

---

## Risks
- **Breaking for installed base** — mitigated by `v4_0` migration; loud CHANGELOG note.
- **Semantic over-replace** — blind sed wrongly renames engine-meaning "harness". Each phase reads context, not just count. The authorship test is the arbiter.
- **Embed/path drift** — `go:embed` + hardcoded `.harness` literals must all move together, or the binary breaks at *runtime*, not compile.
- **Lexicon** — "habitat" is fully retired; the vocabulary is **charter / harness**, period. Any stray "habitat" in older docs is a bug to sweep, not a term to preserve.

---

## Appendix — AAIF motivation (downstream, not this repo's code work)
From `all-the-things/keystone-rebrand.md`: the rebrand is the on-ramp to an AAIF
project submission. Beyond this rename, the submission needs: 2-org production
adoption, committers from 2+ orgs, `GOVERNANCE.md` + roadmap. Those are separate
work — noted here so the rebrand copy stays consistent with the submission pitch
("Keystone manages your coding-agent charter").
