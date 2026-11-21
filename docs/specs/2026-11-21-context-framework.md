---
kind: spec
id: 2026-11-21-context-framework
status: approved
owner: tacoda
approved: 2026-11-21
---

# Keystone 6.0 — the context framework

## Intent (restated)

Keystone becomes the **agent context framework**. The thing it manages
keeps its name — the **charter** — with a new definition: **the charter is
the codified context in the repository**, all of it, not only the authored
standards and constraints. Keystone manages that context as one
standardized, lazy-loaded, agent-agnostic layer — the primitives under
`.agents/` **and** the repository's adopted documents — queried through a
**CLI** and an **MCP server** by any coding agent, and browsable in a **web
dashboard**: an easy, visual, interactive way to work with the charter.
Keystone manages the charter; it does not run the agent.

## The four pillars

1. **The model gets only what it needs.** Nothing loads eagerly except the
   index; every body opens on demand, scoped by paths, task, or query;
   budgets cap the always-loaded tier and `verify` enforces them.
2. **Context is managed, including drift.** The charter is living context:
   keystone detects when it no longer matches the code (diff- and
   region-level drift), names the stale primitives, and helps refresh them
   — so the layer stays trustworthy, not just compact.
3. **One place, compatible with any coding agent.** A single charter under
   `.agents/` plus adopted repo docs; every coding agent reaches it through
   CLI, MCP, the AGENTS.md pointer, or the opt-in compat projection.
4. **Primitives distribute easily via policies.** A policy is a versioned
   bundle of primitives; teams vendor it with `keystone policy add`, update
   it in place, and its rules, references, corpus, skills, and hooks
   cascade into every adopting repo — project wins by default, `strict`
   wins absolutely.

**Why use it:** a team running multiple coding agents gets one place where
context lives, delivery of that context to every coding agent, and the
tooling to manage it effectively — instead of context scattered across
per-host folders with no index and no budget. Policies make the same
primitives easy to distribute across projects and teams: vendor once, and
the cascade carries rules, references, skills, and hooks into every repo
that adopts it.

The layer **enables and gets out of the way**. It never impedes the agent
except at the initial `keystone init` step.

## Codified vs. ambient

Keystone manages **codified context**: authored, versioned files in the
repository — rules, references, corpus entries, skills, and the rest of the
primitive kinds, plus adopted repository documents. **Ambient context is out
of scope**: MCP servers, CLIs, APIs, external services, and anything live.

Keystone is **not a harness** and never grows one. The layer is aware of
the **inferential/deterministic boundary**: what an agent *interprets*
(rules, references, corpus, skill and agent prompts) is codified here;
what must *execute deterministically* (builds, tests, gates) belongs to
hosts and CI, which the charter can name but never replace.

## Decisions (locked with the owner, 2026-11-21)

| # | Question | Decision |
|---|---|---|
| 1 | `reference` vs `corpus` | Keep `reference` (the why behind a rule). Add `corpus` as a **new** kind: standalone information with no rule attached. |
| 2 | `subagent` → `agent` | Yes. Kind name becomes `agent`; folder stays `.agents/agents/`. |
| 3 | Keep `primitive`? | Yes. Every kind is an agent primitive; a rule is a primitive, and so is a corpus entry. |
| 4 | Constraint/capability split | **Drop it.** Kinds are kinds; the index stops sorting and labeling by category. |
| 5 | `eval`, `document`, `signal`, `policy` | Keep as-is. |
| 6 | Folder | `.agents/` stays the root, with a fixed set of canonical subfolders. |
| 7 | Boundary | The repository. Corpus covers docs that live outside `.agents/` today (`GLOSSARY.md`, `docs/adr/`, `docs/plans/`, …), adopted in place — not moved, not edited. |
| 8 | Projection | Keep the compatibility layer (owner confirmed 2026-11-21: removal is an unacceptable tradeoff — native activation and host-side enforcement stay). Demote to opt-in compat; the context layer (CLI + MCP) is the front door. Phase 6 includes a minimality pass: the projection design must be minimal and clean. |
| 9 | Frontmatter features | Deferred. After 6.0 core: catalogue what exists, add what earns its keep. Bar: easy to use, easy to understand, framework-like extensibility, enable-and-get-out-of-the-way. |
| 10 | Governance features & interfaces | Hooks, permissions, evals, signals, policies, learning: keep as-is. The web dashboard stays a first-class interface (owner, 2026-11-21): the easy visual interactive way to work with the charter. |
| 11 | Terminology | **Keep "charter"** (owner, 2026-11-21): the charter *is* the codified context in the repository. No charter→context rename; existing surface names stay (`internal/framework/charter`, `keystone charter`, `keystone_charter_*`). New surfaces use charter-consistent names: `charter.roots`, `keystone_charter_query`. Positioning shifts from "agent charter framework" to "agent context framework". |
| 12 | No custom harness | **Remove the `tool` kind** (owner, 2026-11-21): keystone manages the context (charter) only — it never runs a loop, serves author-defined callables, or acts as an engine. Deterministic execution stays with hosts and CI. The framework is **aware of the inferential/deterministic boundary**: it codifies what agents interpret (rules, references, corpus, prompts) and declares checks, and it says so — anything that must *execute* belongs outside this layer. Entirely agent-agnostic. |
| 13 | Managed-kind scope | **In:** skills and commands (prompts and procedures the agent interprets) and **permissions** (declared allow/ask/deny — hosts enforce). **Out:** tools, MCP servers, and plugins as charter kinds — keystone authors and serves none of them (keystone's own MCP server is the query interface for the charter, not a managed primitive). Host hook-bridging is projection mechanics, not a plugin kind. |
| 14 | The kind triad (owner, 2026-11-21) | **rules** = how and specifics (directives, `paths:`-scoped). **references** = why (rationale, explanations, deeper reasoning). **corpus** = what (facts, general patterns, product information; standalone, no rule required). Keep the kind name `reference` (not `resource`); the triad is framing, not a rename. The `corpus:` frontmatter field (cited reference ids) keeps its name; 4.x `kind: corpus` stays a legacy alias for `reference` while the `.agents/corpus/` folder means the new corpus kind. |
| 15 | Availability over adoption (owner, 2026-11-21) | **No default phased workflow.** init stops planting the lifecycle skills (spec, orient, task, review, verify, mode, learn, synthesize, audit, …) and the phase rules into every project. Everything keystone can do is **available and invocable** — by users, scripts, rules, skills, and hosts — through CLI verbs, MCP tools, and opt-in starter packs; the learning and pruning flywheels are the model. Users and teams decide what to adopt in practice. |

## Naming map

| Surface | 5.x | 6.0 |
|---|---|---|
| Concept "charter" | the authored standards that constrain agents | **the codified context in the repository** (same word, expanded definition) |
| Positioning | the agent charter framework | the agent context framework |
| Category split | constraint / capability | *(removed)* |
| Kind | `rule` | `rule` (unchanged) |
| Kind | `reference` | `reference` (unchanged) |
| Kind | — | `corpus` (new: standalone information) |
| Kind | `hook`, `permission`, `eval`, `skill`, `command`, `document` | unchanged |
| Kind | `tool` | **removed** (harness machinery, not context; `toolrun` and MCP tool-serving go with it) |
| Kind | `subagent` | `agent` |
| Unit | primitive | primitive (unchanged) |
| Go / CLI / MCP names | `internal/framework/charter`, `keystone charter …`, `keystone_charter_*` | unchanged |
| Config | — | `charter.roots` (globs adopting repo docs as corpus) |
| 4.x aliases | `guide`, `corpus`(folder), `sensor`, `posture`, `agent`, `playbook`, `pattern` parse silently | retired; `lint --fix` rewrites; `corpus` folder now means the corpus kind |

`.agents/` stays the root; `keystone.json` stays; `INDEX.json` /
`INDEX.lite.json` stay (category fields dropped; `INDEX.lite.json` becomes
kind, id, description). No code-level renames beyond the kind rename and
the category removal — "charter" keeps its name everywhere.

## Acceptance criteria

A reviewer can verify each objectively.

1. **Query front door.** On a fresh `keystone init` repo, an MCP-capable
   agent discovers and loads every piece of context through MCP alone —
   index, body, search, and region/task query — with **zero host-specific
   files** on disk. The CLI exposes the same reads from the same index.
2. **Nothing eager but the index.** The only always-loaded artifacts are the
   AGENTS.md pointer block and `INDEX.lite.json`. `keystone budget` reports
   the always-loaded tier of this repo under its cap after conversion.
3. **Corpus works both ways.** An authored entry in `.agents/corpus/`
   round-trips (new → index → show → search). A repo document outside
   `.agents/` (this repo's `GLOSSARY.md`, `docs/adr/**`) is indexed and
   queryable **in place** — unmoved, unedited, provenance-marked.
4. **Redefinition lands cleanly.** The word "charter" remains everywhere
   it appears today, defined as the codified context in the repository —
   README, GLOSSARY, MCP instructions, and the AGENTS.md template all say
   so; no constraint/capability labels remain in index, CLI, or generated
   AGENTS.md; the `subagent` kind is `agent`. `keystone migrate` plus
   `keystone lint --fix` bring a 5.x tree (this one) to 6.0 green.
5. **Kept features still work.** Hooks, permissions, evals, signals,
   policies, learning, and the dashboard pass their suites unchanged in
   behavior.
6. **Projection is opt-in.** Fresh `keystone init` writes nothing outside
   `.agents/`, `.keystone/`, `keystone.json`, and the AGENTS.md block.
   Explicit opt-in `keystone project` still projects with a green lockfile.
7. **Release.** 6.0.0 ships per the release checklist; README, GLOSSARY,
   docs, and the site lead with the context framing.

## Out of scope

- Ambient context: MCP servers, CLIs, APIs, external services.
- Frontmatter feature additions (decision 9 — separate pass after 6.0).
- Removing projection outright (decision 8 — revisit after 6.0).
- Semantic/embedding search.
