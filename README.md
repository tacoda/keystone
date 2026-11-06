# Keystone

<!-- voice-register: informational -->
<!-- voice-english: ste -->

Keystone is the agent charter framework. You write the standards your coding
agents follow in one place, `.agents/`, and keystone delivers them to every
agent you use: Claude Code, Cursor, Copilot, Codex, opencode, Continue, Cline,
Goose, Pi, and Aider. The sources follow the AGENTS.md and Agent Skills
standards, so a host that reads them natively gets them as written. Every
other host gets links or translated files. Where a host can't enforce
something, keystone tells you, and the agent too.

```
brew install tacoda/tap/keystone      # or grab a release binary
keystone init                         # writes .agents/ and .keystone/, projects every host
keystone allocate                     # already have a setup? see where it belongs in .agents/
keystone budget                       # what each host loads into every session
keystone mcp install --agent claude-code
```

Keystone isn't a harness. A harness runs the model: the coding agent, the
orchestrator, the runner. Keystone manages the charter that constrains
whatever harness you run. After install, the tools are optional, because the
charter is plain files that any agent can read.

## Two halves: constraints and capabilities

Every primitive in `.agents/` is one of ten kinds. **Constraints** limit, gate,
or check what the agent does. **Capabilities** add things it can do.

| Category   | Kind         | Folder                 | What it is                                                        |
| ---------- | ------------ | ---------------------- | ----------------------------------------------------------------- |
| constraint | `rule`       | `.agents/rules/`       | A directive. It applies when a touched file matches its `paths:`, or everywhere without them. |
| constraint | `reference`  | `.agents/references/`  | The reasoning behind a rule, opened on demand.                    |
| constraint | `hook`       | `.agents/hooks/`       | A check that runs on an event. A non-zero exit blocks the agent.  |
| constraint | `permission` | `.agents/permissions/` | Allow, ask, and deny rules for tools.                             |
| constraint | `eval`       | `.agents/evals/`       | A test that the charter changes agent behavior.                   |
| capability | `skill`      | `.agents/skills/`      | An Agent Skills `SKILL.md`. It activates when its description matches the task. |
| capability | `command`    | `.agents/commands/`    | A unit of work the user asks for by name.                         |
| capability | `subagent`   | `.agents/agents/`      | A role the agent delegates to.                                    |
| capability | `tool`       | `.agents/tools/`       | A callable you write: over MCP, as a command, or as a host plugin. |
| capability | `document`   | `.agents/documents/`   | A template for governed output, such as a plan or an ADR.         |

Keystone's own state lives in `.keystone/`: the index, the lockfile, the
learning inbox, work in progress, and vendored policies. You edit `.agents/`.
You never edit `.keystone/` by hand.

The 4.x names (`guide`, `corpus`, `sensor`, `posture`, `agent`, `playbook`,
`pattern`) still parse. `keystone lint` warns with the 5.0 spelling, and
`keystone lint --fix` applies it.

## One source, every host

`keystone project` delivers `.agents/` to each host in `keystone.json`
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
`<!-- keystone:end -->`. It holds the orientation, the iron laws, the rules
that apply everywhere, the tools, and one line per host that says what that
host loads by itself. Keystone rewrites only the block. Your text around it
stays. `CLAUDE.md` gets a block that imports `AGENTS.md`.

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

`keystone verify` fails when a cap is exceeded, and names the file and the
overage. The estimate is a word count. Real tokenizers count about 1.35
times more, and the caps use the same estimate.

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
| `keystone init`                | Seed `.agents/` and `.keystone/`, then project every host. `--overwrite` replaces keystone's files. |
| `keystone allocate`            | Report where an existing setup belongs. `--apply` writes, `--overwrite` hands originals over. |
| `keystone budget`              | The context each host loads, against your caps. `--json` for CI.     |
| `keystone new <kind> <id>`     | Scaffold any of the ten kinds, a `policy`, or an `adapter` (a host keystone has no profile for). |
| `keystone index`               | Rebuild `.keystone/INDEX.json` and `INDEX.lite.json`.                |
| `keystone project`             | Deliver `.agents/` to every host in `adapters`.                      |
| `keystone lint`                | Check frontmatter. `--fix` applies the 5.0 spelling.                 |
| `keystone verify`              | Check policy drift, strict items, budget caps, and projection drift. |
| `keystone list`                | List primitives, constraints first. Filter by kind or category.     |
| `keystone show`, `explain`     | One primitive: its fields, its links, how it activates.              |
| `keystone charter coverage`    | Project files that no rule's `paths:` covers.                        |
| `keystone charter conformance` | Does the repo conform to its charter?                                |
| `keystone hook fire <event>`   | Run the hooks on an event. Host bridges call it.                     |
| `keystone signal fire <name>`  | Fire a keystone signal (an event no host sees).                      |
| `keystone tool run <id>`       | Run a tool the way the MCP server does.                              |
| `keystone migrate up`          | Upgrade an older install. See the upgrade guide.                     |
| `keystone watch`               | Re-index and re-project on every save under `.agents/`.              |
| `keystone mcp serve`           | Run the MCP server over stdio.                                       |
| `keystone mcp install`         | Write a host's MCP config.                                           |
| `keystone web serve`           | Open the dashboard at `http://localhost:4773`.                       |

Every command takes `--help`.

## MCP server and dashboard

`keystone mcp install --agent <host>` writes the host's MCP config, and the
host starts the server. The server reads the same files as the CLI. It offers
read tools (`keystone_list_primitives`, `keystone_get_reference`,
`keystone_budget`, …), write tools (`keystone_new_<kind>`), and each
`transport: mcp` tool as `keystone_tool_<id>`.

`keystone web serve` opens a local dashboard. Its nav splits the charter into
constraints and capabilities, and it has pages for coverage, the budget,
signals, evals, and the learning flywheels.

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
