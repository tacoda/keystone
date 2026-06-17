# Keystone

**Keystone is the agent harness framework.** A small Go toolchain — CLI + MCP
server + localhost dashboard — that drops a structured, primitive-typed
markdown harness into your repo and keeps it healthy. Once installed,
neither tool is required at runtime. The harness is plain files; any agent
that reads `.keystone/INDEX.json` and the bodies it points at can operate
the project.

The CLI authors and maintains the harness. The MCP server dispatches the
same harness to host agents (Claude Code, Cursor, Codex, …) over the
model-context-protocol. The dashboard at `http://localhost:4773` gives a
local operator view — metrics, insights, prune, eval runs, source query.
All three are convenience. The harness alone is enough.

```
brew install tacoda/tap/keystone    # or grab a release binary
keystone init                       # writes .keystone/ + agent menu
keystone index                      # emits .keystone/INDEX.json
keystone mcp install --agent claude-code   # one-line agent wiring
keystone web serve                  # optional: open the dashboard
```

## What's in the harness

`.keystone/harness/` holds **11 primitive kinds** in two layers:

| Layer        | Kinds                                                                 |
| ------------ | --------------------------------------------------------------------- |
| **Framework**| guide, corpus, sensor, action, playbook, eval, source                 |
| **Agent**    | rule, skill, subagent, command, persona                               |

Framework kinds are keystone-canonical (the agent doesn't natively know
about them — the harness teaches it). Agent kinds align with what host
agents already understand; keystone projects them to native paths
(`.claude/skills/`, `.claude/agents/`, `.claude/commands/`,
`.cursor/rules/`, …) on `keystone project`.

Every primitive carries canonical frontmatter — `kind`, `id`,
`description`, plus per-kind required fields (`globs`, `phase`,
`triggers`, `severity`, `tools`, `args`, `traces`, `deps`). The walker
emits a single `.keystone/INDEX.json` listing every primitive's
descriptor; agents read the index first and open bodies only when
their activation conditions match.

## The contract: rules → corpus → external → ask

Five-stage runtime resolution flow — encoded in the MCP server's
`instructions` block and in `guides/process/runtime-resolution.md`.

1. **Rules.** Find applicable guides + sensors via `INDEX.json`,
   filtered by touched-files globs and phase. Project wins by default;
   policies refine via nesting; `strict` items lock absolutely.
2. **Corpus.** When a rule's body isn't enough, follow its `traces:`
   to the linked corpus reasoning.
3. **External.** Insufficient still? Query configured external sources
   (Linear, Confluence, folder, URL) — only when the MCP server is
   running and `.keystone/context.json` declares them.
4. **Apply.** External results are never applied silently. The agent
   asks the user: project, team policy, org policy, or session?
5. **Contradictions.** Conflicts between rules, corpus, and external
   answers trigger the agent to ask the user to resolve.

## CLI

| Verb                     | What it does                                                                  |
| ------------------------ | ----------------------------------------------------------------------------- |
| `keystone init`          | Minimum-friction scaffold. One question (agent target) or zero with `--agent`. |
| `keystone index`         | Walk `.keystone/harness/`, emit `INDEX.json`.                                  |
| `keystone lint`          | Validate primitive frontmatter; required fields per kind, deps integrity.      |
| `keystone project`       | Regenerate `.claude/` / `.cursor/` host projections from canonical sources.    |
| `keystone verify`        | Cascade + policy-drift check.                                                  |
| `keystone migrate`       | One-shot 1.x → 2.0 layout + schema upgrade.                                    |
| `keystone new <kind>`    | Scaffold any of the 11 primitive kinds + adapter + policy.                     |
| `keystone search <q>`    | Full-text search across every primitive.                                       |
| `keystone graph`         | Render the primitive-relationship graph (Mermaid or DOT).                      |
| `keystone watch`         | fsnotify loop: index + project + lint on change.                               |
| `keystone snapshot`      | Save / list / restore tarballs of `.keystone/` for safe experiments.           |
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

Tool surface:

- **Read** — `keystone_list_primitives`, `keystone_get_primitive`,
  `keystone_get_corpus`
- **Search** — `keystone_search`
- **Eval** — `keystone_eval_list`, `keystone_eval_run`, `keystone_eval_report`,
  `keystone_eval_baseline`
- **Sources** — `keystone_source_list`, `keystone_source_query`,
  `keystone_source_health`
- **Write** — `keystone_new_<kind>` for every kind, plus
  `keystone_harness_bootstrap`, `keystone_target_add`,
  `keystone_index_refresh`, `keystone_project_refresh`

Resource surface:

- `keystone://index` — full `INDEX.json`
- `keystone://primitive/{kind}/{id}` — one body
- `keystone://harness/status` — install audit
- `keystone://source/list`, `keystone://source/{name}/health`
- `skill://list`, `skill://{name}/SKILL.md` — auto-discovery

Prompt surface: `keystone_bootstrap`, `keystone_task`, `keystone_audit`,
`keystone_learn`.

## Dashboard

`keystone web serve --port 4773` (the default — KEYS on a phone keypad).
Localhost-only; same binary; reuses the same data model as the CLI and
MCP server.

Pages:

- **home** — counts by kind, source healths
- **metrics** — primitive counts, lint stats, freshness, index health
- **insights** — suggested actions to improve harness performance
- **primitives** — table w/ kind + glob filter; detail page w/
  cross-references
- **policies** — list + add/remove
- **investigator** — primitives grouped by cascade layer, w/ search
- **sources** — list + add/remove + per-source query + health probe
- **verify** — one-click `keystone verify` + result pane
- **prune** — lint findings + heuristics (orphan corpus, empty
  bodies, duplicate descriptions) w/ per-row prune
- **inbox** — walk learning candidates, accept/reject
- **flywheels** — copy-paste invocations for learn / synthesize / audit
- **evals** — declared evals + run button + report fragment
- **search** — HTMX-live full-text across the harness
- **graph** — Mermaid-rendered dependency graph

SSE push at `/events` swaps fragments when files in `.keystone/` change.
REST API under `/api/` is read-only.

## Policies

Vendored harness fragments — declared in `keystone.json`, pinned by
version, hash-verified, drift-reset on `keystone verify`. Policies can
ship every framework abstraction (guide, corpus, sensor, action,
playbook, eval, source) but **never** agent abstractions (rule, skill,
subagent, command, persona) — those stay project-owned. Cascade order:
project wins by default; nested policies refine outer; `strict`
declarations lock items absolutely.

```bash
keystone policy add tacoda/tacoda-org@v2.0.0
```

## After install, the tools are optional

The harness is markdown on disk. An agent that reads
`.keystone/INDEX.json` plus the primitive bodies it points at can
operate the harness without keystone installed. The CLI helps you
author + maintain; the MCP server gives the agent structured access;
the dashboard gives you a local operator view. None are runtime
dependencies.

## License

MIT. See [`LICENSE`](LICENSE). Contributions welcome — see
[`CONTRIBUTING.md`](CONTRIBUTING.md).

---

**Migration from 1.x:** run `keystone migrate` once. It moves
`harness/` to `.keystone/harness/`, lifts `keystone.lock` to
`.keystone/lockfile.json`, renames `plugins/` → `policies/` (and
`keystone-plugin.json` → `keystone-policy.json`), rewrites
`keystone.json` to v2 schema, regenerates `INDEX.json` and host
projections. Idempotent — safe to re-run. Backup first via
`keystone snapshot save --label pre-2.0`.
