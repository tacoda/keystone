# Keystone

<!-- voice-register: informational -->
<!-- voice-english: ste -->

Keystone is the **charter framework** — the framework for managing context.
The **charter** is the codified context in your repository, and it lives in
one place: `.agents/` (plus the repo documents you adopt via `charter.roots`),
indexed at `.keystone/INDEX.json`. It is the first level of the **habitat**
your agent works in.

The charter is agent-agnostic: any coding agent queries the same context
through the CLI or the MCP server, and people browse it in the dashboard —
Claude Code, Cursor, Copilot, Codex, opencode, Continue, Cline, Goose, Pi,
and Aider all read the same charter. That makes it good for teams: the
context stays in one manageable place, regardless of which host each
developer prefers. Managing it is easy: one indexed place, queried on demand,
never loaded eagerly. And the charter is dynamic:
`learn` / `synthesize` / `audit` / `drift` and the relocation ladder
(`charter move` / `promote` / `demote`) keep it improving — learning, pruning,
and relocating context so the agent's output gets better over time.

```
brew install tacoda/tap/keystone      # or grab a release binary
keystone init                         # writes .agents/, .keystone/, keystone.json, AGENTS.md
keystone charter query cmd/...       # what context applies here? (CLI)
keystone mcp install --agent claude-code   # same reads over MCP
keystone web serve                    # dashboard: primitives, drift, flywheels
keystone allocate                     # already have a setup? see where it belongs in .agents/
keystone budget                       # what each host loads into every session
```

Keystone isn't a harness. A harness runs the model: the coding agent, the
orchestrator, the runner. Keystone manages the charter — the codified context
— that whatever harness you run reads. After install, the tools are optional,
because the charter is plain files that any agent can read.

For legacy projects, keystone is the way in: drop it into a codebase no one
has chartered yet, characterize what is true, and start tackling longstanding
technical debt at scale — large-scale refactorings behind migration
documents, test expansion per region, and the other code concerns the
rule set names. The transition kit makes the codebase agent-ready: agents
check the empirical map and the migration documents before they touch a
region, instead of re-deriving the shape of the work from scratch.

In a way this is an agent-readiness kit: it gives you an easy place to add
and edit agent primitives, assess and adjust them, and a way to onboard a
legacy project. The easy part is what keystone scaffolds — adding things is
easy. The hard part is the code changes, and they must be able to happen
while still delivering product work. And it does that without giving up on
shipping new features: the kits constrain the regions they name —
characterize-before-touch, tests-first-per-region, no-rewrite-without-migration —
everything else keeps moving; new feature work and legacy stabilization run
in parallel, one charter, no freeze.

## The kinds

The charter is the codified context of the repository — everything a coding
agent should be able to look up about how this codebase works and how to
work in it. Every primitive is one of ten kinds, with frontmatter that the
index holds and the agent opens on demand. The triad: **rule** = how,
**reference** = why, **corpus** = what.

| Kind         | Folder                 | What it is                                                        |
| ------------ | ---------------------- | ----------------------------------------------------------------- |
| `rule`       | `.agents/rules/`       | A directive. It applies when a touched file matches its `paths:`, or everywhere without them. |
| `reference`  | `.agents/references/`  | The reasoning behind a rule, opened on demand.                    |
| `hook`       | `.agents/hooks/`       | A check that runs on an event. A non-zero exit blocks the agent.  |
| `permission` | `.agents/permissions/` | Allow, ask, and deny rules for tools.                             |
| `eval`       | `.agents/evals/`       | A test that the charter changes agent behavior.                   |
| `skill`      | `.agents/skills/`      | An Agent Skills `SKILL.md`. It activates when its description matches the task. |
| `command`    | `.agents/commands/`    | A unit of work the user asks for by name.                         |
| `agent`      | `.agents/agents/`      | A role the agent delegates to (a subagent).                       |
| `document`   | `.agents/documents/`   | A template for governed output, such as a plan or an ADR.         |

Keystone manages context, not callables: it is not a harness, and the charter
carries no `tool` kind — what must execute deterministically belongs to hosts
and CI, which the charter can name but never replace.

Keystone's own state lives in `.keystone/`: the index, the lockfile, the
learning inbox, work in progress, and vendored policies. You edit `.agents/`.
You never edit `.keystone/` by hand.

The 4.x names (`guide`, `sensor`, `posture`, `playbook`, `pattern`) are
retired in 6.0. `keystone lint` warns with the 6.0 spelling, and
`keystone lint --fix` rewrites them; `keystone migrate up` upgrades a whole
older tree.

## One source, queried by every agent

The front door is the query layer, not any host folder: `keystone charter
query <files…|--task "…">` on the CLI, `keystone_charter_query` over MCP,
and the dashboard. Everything keystone can do is invocable there — including
the hygiene verbs `keystone learn | synthesize | audit | drift | mode` (and
their `keystone_*` MCP counterparts), so scripts and agents drive the same
charter without any planted skills.

## Compat: one source, every host folder (opt-in)

Host-native projection is the compatibility layer. A fresh `keystone init`
writes no host files; `keystone target add <host>` opts one in, and
`keystone project` then delivers `.agents/` to each host in `keystone.json`
`adapters`, by the cheapest method that host supports:

- **Native.** The host already reads the path. Codex and opencode read
  `.agents/skills/`.
- **Link.** The host reads the same format at another path. Claude Code gets
  a symlink per skill in `.claude/skills/`. Set `"link_mode": "copy"` where
  symlinks are a problem.
- **Shim.** The host needs a different format. Cursor gets `.mdc` rules,
  Copilot gets `.instructions.md` files, and Codex gets TOML subagents.
- **Index.** The host has no slot for the kind. `AGENTS.md` lists it, and the
  agent opens it from `.agents/` when it applies.

`AGENTS.md` carries a managed block between `<!-- keystone:start -->` and
`<!-- keystone:end -->`. It holds only the orientation: what the charter is,
the index pointer, the query surface (CLI + MCP), and one line per projected
host. No primitive bodies load eagerly. Keystone rewrites only the block.
Your text around it stays. `CLAUDE.md` gets a block that imports `AGENTS.md`.

Keystone records every file it writes in `.keystone/lockfile.json`. It never
replaces a file it didn't write. Pass `--overwrite` when you want it to.

### Hooks and permissions

A hook runs through a bridge in each host's own hook config: Claude Code,
Cursor, Copilot, Codex, and opencode. A failing check blocks the tool call,
and so does an error before the checks run. I've run that on Claude Code;
for the other four it's what each host's hook docs specify.

Permission rules use Claude Code syntax, `Bash(go test:*)`. Keystone writes
them to Claude Code unchanged, translates them for Cursor and Codex, and
never makes an allow broader than you wrote it.

For an agent that isn't on the list, `keystone new adapter <name>` defines
it in `.agents/hosts/<name>.json`: per kind, it says whether the host reads
`.agents/` natively, gets links or copies in its own folder, or opens the
kind from `AGENTS.md`.

Some hosts can't run hooks or enforce some permission rules. `keystone
project` names each gap, and `AGENTS.md` tells the agent on that host to run
the checks itself. [`docs/hosts.md`](docs/hosts.md) has the matrix and its
sources.

## Bring what you already have

`keystone init` seeds a project and skips every file that already exists.
To convert an existing setup, run `keystone allocate`. It reads `.claude/`,
`.cursor/rules`, `.cursorrules`, Copilot instructions, `.clinerules`,
`.windsurfrules`, `.continue/`, and `.opencode/`. For each item, it reports
the kind and the `.agents/` path where the item belongs.

It also splits `CLAUDE.md`, `AGENTS.md`, `GEMINI.md`, and `.goosehints` into
sections, and proposes moving each one that doesn't need to load every
session. A section that names file patterns becomes a path-scoped rule, a
numbered procedure becomes a command, and long reasoning becomes a
reference. Iron laws stay where they are.

- `keystone allocate` writes nothing.
- `keystone allocate --apply` writes new files under `.agents/` and changes
  no original.
- `keystone allocate --overwrite` hands the originals to keystone, so the
  next projection replaces them.

## Context budget

`keystone budget` measures two tiers for each host. The **always loaded**
tier is what the host reads at the start of every session: root files, the
name and description of every skill, command, and subagent it lists, and
rules without `paths:`. The **on demand** tier is every body the agent opens
when it applies.

Set caps in `keystone.json`:

```json
"budgets": {
  "always_loaded": { "max_tokens": 4000 },
  "reference":     { "max_tokens_per_load": 3000 }
}
```

With no `budgets` at all, a default `always_loaded` cap of 8000 (estimate
words) applies. `keystone verify` fails when a cap is exceeded, and names
the file and the overage. The estimate is a word count. Real tokenizers
count about 1.35 times more, and the caps use the same estimate.

## How an agent resolves a scenario

The MCP server's instructions and the `runtime-resolution` rule state the
same order:

1. **Rules.** Find the rules for the touched files and the current phase.
   The project wins by default. Policies refine it, and a `strict` item is
   absolute.
2. **References.** When a rule's body isn't enough, open the references its
   `corpus:` field links.
3. **Contradictions.** When a rule and its reference disagree, ask the user.

## CLI

| Command                        | What it does                                                         |
| ------------------------------ | -------------------------------------------------------------------- |
| `keystone init`                | Seed `.agents/`, `.keystone/`, `keystone.json`, and the AGENTS.md block. Host projection is opt-in: `keystone target add <host>`. `--overwrite` replaces keystone's files. |
| `keystone allocate`            | Report where an existing setup belongs. `--apply` writes, `--overwrite` hands originals over. |
| `keystone budget`              | The context each host loads, against the caps (default always_loaded: 8000). `--json` for CI. |
| `keystone new <kind> <id>`     | Scaffold any of the ten kinds, a `policy`, or an `adapter` (a host keystone has no profile for). |
| `keystone index`               | Rebuild `.keystone/INDEX.json` and `INDEX.lite.json` (incl. adopted `charter.roots` docs). |
| `keystone project`             | Deliver `.agents/` to every host in `adapters` (opt-in compat layer). |
| `keystone lint`                | Check frontmatter. `--fix` applies the 6.0 spelling.                 |
| `keystone verify`              | Check policy drift, strict items, budget caps, projection drift; warns on context drift. |
| `keystone list`                | List primitives, grouped by kind. Filter by kind or tag.            |
| `keystone show`, `explain`     | One primitive: its fields, its links, how it activates.              |
| `keystone charter query`       | What context applies to these files / this task? `--full` for bodies. The front door. |
| `keystone charter move`        | Relocate a primitive to another kind (e.g. rule → hook); `promote` / `demote` step the corpus → reference → rule ladder. Rewrites citations, re-indexes. |
| `keystone charter coverage`    | Project files that no rule's `paths:` covers.                        |
| `keystone charter conformance` | Does the repo conform to its charter?                                |
| `keystone search <query>`      | Full-text search over descriptors and bodies.                       |
| `keystone learn` / `learning list` / `synthesize` | The Learning flywheel: capture candidates, then promote accepted ones to their proposed layer. |
| `keystone audit`               | Dual-flywheel report: inbox + pruning candidates (uncited/unscoped/untriggered). |
| `keystone drift`               | Context-drift warnings: dead links, vanished `paths:`, stale `verified:`. `[files…|--diff]` scopes it. |
| `keystone mode [paired|solo|autopilot]` | Read or set the pacing mode.                                  |
| `keystone hook fire <event>`   | Run the hooks on an event. Host bridges call it.                     |
| `keystone signal fire <name>`  | Fire a keystone signal (an event no host sees).                      |
| `keystone migrate up`          | Upgrade an older install. See the upgrade guide.                     |
| `keystone watch`               | Re-index and re-project on every save under `.agents/`.              |
| `keystone mcp serve`           | Run the MCP server over stdio.                                       |
| `keystone mcp install`         | Write a host's MCP config.                                           |
| `keystone web serve`           | Open the dashboard at `http://localhost:4773`.                       |

Every command takes `--help`.

## MCP server and dashboard

`keystone mcp install --agent <host>` writes the host's MCP config, and the
host starts the server. The server reads the same files as the CLI. It offers
read tools (`keystone_charter_query`, `keystone_list_primitives`,
`keystone_get_primitive`, `keystone_search`, `keystone_drift_report`,
`keystone_budget`, …), write tools (`keystone_new_<kind>`,
`keystone_charter_move`, …), and the hygiene tools (`keystone_learn`,
`keystone_synthesize`, `keystone_audit`, `keystone_mode`).

`keystone web serve` opens a local dashboard over the same files: primitives,
coverage, the budget, signals, evals, the drift pane, and the learning
flywheels.

## Policies

A policy is a charter fragment you vendor from another repo. Keystone pins it
by version, checks its hash, and resets drift on `keystone verify`. The
project layer wins by default, and a policy's `strict` items lock.

```bash
keystone policy add tacoda/tacoda-org@v2.0.0
```

## Upgrading from 4.x

Run `keystone migrate up`, then install the 5.0 binary before your hooks run
again. [`docs/upgrade-4.x-to-5.0.md`](docs/upgrade-4.x-to-5.0.md) covers what
moves, what each 4.x name became, and the one conflict that stops the
migration.

## License

MIT. See [`LICENSE`](LICENSE). Contributions are welcome: see
[`CONTRIBUTING.md`](CONTRIBUTING.md). [`GOVERNANCE.md`](GOVERNANCE.md) covers
how the project and the charter change.
