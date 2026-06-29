# Keystone

**Keystone is the agent harness framework.** A small Go toolchain — CLI + MCP
server + localhost dashboard — that drops a structured, primitive-typed
markdown harness into your repo and keeps it healthy. Once installed,
neither tool is required at runtime. The harness is plain files; any agent
that reads `.harness/INDEX.json` and the bodies it points at can operate
the project.

The CLI authors and maintains the harness. The MCP server dispatches the
same harness to host agents (Claude Code, Cursor, Codex, opencode, …) over the
model-context-protocol. The dashboard at `http://localhost:4773` gives a
local operator view — metrics, insights, prune, eval runs.
All three are convenience. The harness alone is enough.

```
brew install tacoda/tap/keystone    # or grab a release binary
keystone init                       # writes .harness/ + agent menu
keystone index                      # emits .harness/INDEX.json
keystone mcp install --agent claude-code   # one-line agent wiring
keystone web serve                  # optional: open the dashboard
```

## What's in the harness

`.harness/` holds **14 primitive kinds**:

| Kind         | What it is                                                            |
| ------------ | --------------------------------------------------------------------- |
| `guide`      | ambient, glob-scoped directive — inferential → `.claude/rules/` shim; computational → a host hook (LSP) |
| `sensor`     | phase-gated check — inferential review (→ agent) or computational (→ hook) |
| `hook`       | the deterministic fire layer: an `event:` → `run:` (shell) or `agent:` dispatch |
| `agent`      | a role spawned as a subagent                                          |
| `command`    | a unit of work / lifecycle step                                       |
| `skill`      | a single composed capability                                          |
| `playbook`   | a composed sequence of commands with `gates:`                         |
| `pattern`    | a reusable documentation pattern (Diátaxis), in prose                 |
| `corpus`     | the reasoning / why, loaded on demand                                 |
| `document`   | a governed output (plan / review / adr / retro / feature)             |
| `concern`    | a composition mixin                                                   |
| `posture`    | tool/permission posture → settings.json                               |
| `tool`       | an author-defined callable (transport: cli \| mcp \| plugin)          |
| `eval`       | the eval harness                                                      |

keystone is the model layer over host primitives: each kind adds
validations, associations, and cross-host projection the bare host file
lacks. It projects to native paths (`.claude/skills/`, `.claude/agents/`,
`.claude/commands/`, `.cursor/rules/`, …) on `keystone project`.

Every primitive carries canonical frontmatter — `kind`, `id`,
`description`, plus per-kind fields (`globs`, `phase`, `triggers`,
`tier`, `mode`, `event`, `run`, `agent`, `returns`, `tools`, `corpus`,
`includes`, …). The walker emits a single `.harness/INDEX.json` listing
every primitive's descriptor; agents read the index first and open
bodies only when their activation conditions match.

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
| `keystone index`         | Walk `.harness/`, emit `INDEX.json`.                                  |
| `keystone lint`          | Validate primitive frontmatter; required fields per kind, deps integrity.      |
| `keystone project`       | Regenerate `.claude/` / `.cursor/` host projections from canonical sources.    |
| `keystone verify`        | Cascade + policy-drift check.                                                  |
| `keystone migrate`       | Version-to-version harness upgrade (… → 3.0). Idempotent.                                    |
| `keystone new <kind>`    | Scaffold any of the 14 primitive kinds + adapter + policy.                     |
| `keystone search <q>`    | Full-text search across every primitive.                                       |
| `keystone graph`         | Render the primitive-relationship graph (Mermaid or DOT).                      |
| `keystone watch`         | fsnotify loop: index + project + lint on change.                               |
| `keystone snapshot`      | Save / list / restore tarballs of `.harness/` for safe experiments.           |
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
  `keystone_harness_bootstrap`, `keystone_target_add`,
  `keystone_index_refresh`, `keystone_project_refresh`

Resource surface:

- `keystone://index` — full `INDEX.json`
- `keystone://primitive/{kind}/{id}` — one body
- `keystone://harness/status` — install audit
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
- **insights** — suggested actions to improve harness performance
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
- **search** — HTMX-live full-text across the harness
- **graph** — Mermaid-rendered dependency graph

SSE push at `/events` swaps fragments when files in `.harness/` change.
REST API under `/api/` is read-only.

## Policies

Vendored harness fragments — declared in `keystone.json`, pinned by
version, hash-verified, drift-reset on `keystone verify`. A policy can
ship any primitive kind; the project layer always wins by default.
Cascade order: project wins; nested policies refine outer; `strict`
declarations lock items absolutely.

```bash
keystone policy add tacoda/tacoda-org@v2.0.0
```

## After install, the tools are optional

The harness is markdown on disk. An agent that reads
`.harness/INDEX.json` plus the primitive bodies it points at can
operate the harness without keystone installed. The CLI helps you
author + maintain; the MCP server gives the agent structured access;
the dashboard gives you a local operator view. None are runtime
dependencies.

## License

MIT. See [`LICENSE`](LICENSE). Contributions welcome — see
[`CONTRIBUTING.md`](CONTRIBUTING.md).

---

**Migration from 1.x:** run `keystone migrate` once. It moves
`harness/` to `.harness/`, lifts `keystone.lock` to
`.harness/lockfile.json`, renames `plugins/` → `policies/` (and
`keystone-plugin.json` → `keystone-policy.json`), rewrites
`keystone.json` to v2 schema, regenerates `INDEX.json` and host
projections. Idempotent — safe to re-run. Backup first via
`keystone snapshot save --label pre-2.0`.
