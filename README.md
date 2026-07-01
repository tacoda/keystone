# Keystone

**Keystone is the coding-agent charter manager** — constraint engineering at
the repository level. A small Go toolchain — CLI + MCP server + localhost
dashboard — that drops a structured, primitive-typed markdown **charter** into
your repo and keeps it healthy. Once installed, neither tool is required at
runtime. The charter is plain files; any agent that reads
`.charter/INDEX.json` and the bodies it points at can operate the project.

A **harness** is the engine that runs the model — the coding agent (Claude
Code, Cursor, Codex, opencode, …), the orchestrator, the runner. Keystone is
not a harness. It manages the **charter**: the standards you *author* to
constrain whatever harness runs, so each unique repo gets reliable, quality
output. (Author the spec → charter. Be the engine → harness. See
[`GLOSSARY.md`](GLOSSARY.md).)

The CLI authors and maintains the charter. The MCP server dispatches the
same charter to host agents (Claude Code, Cursor, Codex, opencode, …) over the
model-context-protocol. The dashboard at `http://localhost:4773` gives a
local operator view — metrics, insights, prune, eval runs.
All three are convenience. The charter alone is enough.

```
brew install tacoda/tap/keystone    # or grab a release binary
keystone init                       # writes .charter/ + agent menu
keystone index                      # emits .charter/INDEX.json
keystone mcp install --agent claude-code   # one-line agent wiring
keystone web serve                  # optional: open the dashboard
```

## What's in the charter

`.charter/` holds **13 primitive kinds**:

| Kind         | What it is                                                            |
| ------------ | --------------------------------------------------------------------- |
| `guide`      | ambient, glob-scoped directive — inferential → `.claude/rules/` shim; computational → a host hook (LSP) |
| `sensor`     | a check that reacts to a signal or host phase (`on:`) — computational (`run:` → exit/HTTP status verdict) or inferential (agent review → schema); gates |
| `agent`      | a role spawned as a subagent                                          |
| `command`    | a unit of work / lifecycle step                                       |
| `skill`      | a single composed capability                                          |
| `playbook`   | a composed sequence of commands with `gates:`                         |
| `pattern`    | a reusable documentation pattern (Diátaxis), in prose                 |
| `corpus`     | the reasoning / why, loaded on demand                                 |
| `document`   | a governed output (plan / review / adr / retro / feature)             |
| `concern`    | a composition mixin                                                   |
| `posture`    | tool/permission posture → settings.json                               |
| `tool`       | an author-defined external callable (transport: cli \| http \| mcp \| plugin); on-demand, or a side-effect when it declares `on:` |
| `eval`       | the eval harness                                                      |

keystone is the model layer over host primitives: each kind adds
validations, associations, and cross-host projection the bare host file
lacks. It projects to native paths (`.claude/skills/`, `.claude/agents/`,
`.claude/commands/`, `.cursor/rules/`, …) on `keystone project`.

Every primitive carries canonical frontmatter — `kind`, `id`,
`description`, plus per-kind fields (`globs`, `phase`, `triggers`,
`tier`, `mode`, `event`, `run`, `agent`, `returns`, `tools`, `corpus`,
`includes`, …). The walker emits a single `.charter/INDEX.json` listing
every primitive's descriptor; agents read the index first and open
bodies only when their activation conditions match.

## Signals

A **signal** is a keystone framework event the host can't see — the
extensible, higher-level counterpart to a host hook phase. A primitive
subscribes to one via `on:` (like a skill declares `triggers:`): a
`sensor` runs a check, a `tool` fires a side-effect, an `agent` runs a
review. Host phases (`PreToolUse`, `Stop`, …) are a closed set bridged
into the host; **any other `on:` value is a signal**, so projects define
their own — declare them in `keystone.json` `signals:` and fire with
`keystone signal fire <name>` (`keystone signal list` to discover). The
`hook` kind is retired — reactions self-subscribe.

## The contract: rules → corpus → ask

Runtime resolution flow — encoded in the MCP server's `instructions`
block and in `guides/process/runtime-resolution.md`.

1. **Rules.** Find applicable guides via `INDEX.json`, filtered by
   touched-files globs and phase. Project wins by default; policies
   refine via nesting; `strict` items lock absolutely.
2. **Corpus.** When a guide's body isn't enough, follow its `corpus:`
   to the linked reasoning.
3. **Contradictions.** Conflicts between rules and corpus trigger the
   agent to ask the user to resolve. (External-system access is a
   `tool`, not a resolution stage.)

## CLI

| Verb                     | What it does                                                                  |
| ------------------------ | ----------------------------------------------------------------------------- |
| `keystone init`          | Minimum-friction scaffold. One question (agent target) or zero with `--agent`. |
| `keystone index`         | Walk `.charter/`, emit `INDEX.json`.                                  |
| `keystone lint`          | Validate primitive frontmatter; required fields per kind, deps integrity.      |
| `keystone project`       | Regenerate `.claude/` / `.cursor/` host projections from canonical sources.    |
| `keystone verify`        | Cascade + policy-drift check.                                                  |
| `keystone charter coverage` | Which project files no guide governs ("uncharted territory").               |
| `keystone charter show`  | Charter roster; `--effective` resolves the post-cascade winning set.           |
| `keystone signal fire\|list` | Fire or list signals (extensible framework events).                        |
| `keystone migrate`       | Version-to-version charter upgrade (… → 4.0). Idempotent.                                    |
| `keystone new <kind>`    | Scaffold any of the 13 primitive kinds + adapter + policy.                     |
| `keystone search <q>`    | Full-text search across every primitive.                                       |
| `keystone graph`         | Render the primitive-relationship graph (Mermaid or DOT).                      |
| `keystone watch`         | fsnotify loop: index + project + lint on change.                               |
| `keystone snapshot`      | Save / list / restore tarballs of `.charter/` for safe experiments.           |
| `keystone eval run`      | Run static + sensor evals. `--baseline <ref>` for regression diffs.            |
| `keystone policy`        | Add / update / remove vendored policies.                                       |
| `keystone mcp serve`     | Launch the MCP server over stdio (host-launched).                              |
| `keystone mcp install`   | Write the host agent's MCP config (`.mcp.json`, `.cursor/mcp.json`, …).        |
| `keystone web serve`     | Open the localhost dashboard.                                                  |

Every command has `--help`. `keystone completion bash|zsh|fish` for shell
autocomplete.

## MCP server

Same binary, separate verb. `keystone mcp install --agent claude-code`
writes `.mcp.json`; the host launches the server on session start.
`keystone mcp install --agent opencode` writes the server into
`opencode.json`'s `mcp` key as a `local` (stdio) server.

Tool surface:

- **Read** — `keystone_list_primitives`, `keystone_get_primitive`,
  `keystone_get_corpus`
- **Search** — `keystone_search`
- **Eval** — `keystone_eval_list`, `keystone_eval_run`, `keystone_eval_report`,
  `keystone_eval_baseline`
- **Write** — `keystone_new_<kind>` for every kind, plus
  `keystone_charter_bootstrap`, `keystone_target_add`,
  `keystone_index_refresh`, `keystone_project_refresh`

Resource surface:

- `keystone://index` — full `INDEX.json`
- `keystone://primitive/{kind}/{id}` — one body
- `keystone://charter/status` — install audit
- `skill://list`, `skill://{name}/SKILL.md` — auto-discovery

Prompt surface: `keystone_bootstrap`, `keystone_task`, `keystone_audit`,
`keystone_learn`.

## Dashboard

`keystone web serve --port 4773` (the default — KEYS on a phone keypad).
Localhost-only; same binary; reuses the same data model as the CLI and
MCP server.

Pages:

- **home** — counts by kind
- **metrics** — primitive counts, lint stats, freshness, index health
- **insights** — suggested actions to improve charter performance
- **primitives** — table w/ kind + glob filter; detail page w/
  cross-references
- **policies** — list + add/remove
- **investigator** — primitives grouped by cascade layer, w/ search
- **verify** — one-click `keystone verify` + result pane
- **prune** — lint findings + heuristics (orphan corpus, empty
  bodies, duplicate descriptions) w/ per-row prune
- **inbox** — walk learning candidates, accept/reject
- **flywheels** — copy-paste invocations for learn / synthesize / audit
- **evals** — declared evals + run button + report fragment
- **search** — HTMX-live full-text across the charter
- **graph** — Mermaid-rendered dependency graph

SSE push at `/events` swaps fragments when files in `.charter/` change.
REST API under `/api/` is read-only.

## Policies

Vendored charter fragments — declared in `keystone.json`, pinned by
version, hash-verified, drift-reset on `keystone verify`. A policy can
ship any primitive kind; the project layer always wins by default.
Cascade order: project wins; nested policies refine outer; `strict`
declarations lock items absolutely.

```bash
keystone policy add tacoda/tacoda-org@v2.0.0
```

## After install, the tools are optional

The charter is markdown on disk. An agent that reads
`.charter/INDEX.json` plus the primitive bodies it points at can
operate the charter without keystone installed. The CLI helps you
author + maintain; the MCP server gives the agent structured access;
the dashboard gives you a local operator view. None are runtime
dependencies.

## License

MIT. See [`LICENSE`](LICENSE). Contributions welcome — see
[`CONTRIBUTING.md`](CONTRIBUTING.md).

---

**Migration to 4.0:** run `keystone migrate` once. It brings any prior
layout up to current — most notably 4.0 renames `.harness/` to
`.charter/` (and `HARNESS.md` → `CHARTER.md`), rewrites `keystone.json`
references, and regenerates `INDEX.json` and host projections.
Idempotent — safe to re-run. Backup first via
`keystone snapshot save --label pre-4.0`.
