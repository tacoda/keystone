# Keystone — the agent harness framework

## Stack

- **Go 1.22+** (1.25 in mise). Single binary, single module.
- **CLI:** Cobra (`cmd/keystone/`). Subcommands wired in `root.go`.
- **MCP server:** mark3labs/mcp-go (`internal/framework/mcp/`).
- **YAML:** gopkg.in/yaml.v3. **One YAML parser, period** (see `guides/idioms/go/stdlib-first`).
- **Templates:** go:embed under `internal/framework/scaffold/templates/`.
- **Release:** GoReleaser via tag push (no `gh release create` — see memory).

## Layout

```
cmd/keystone/         — CLI entrypoint + every subcommand (one file each)
internal/framework/   — agent-agnostic core
  primitive/          — descriptor parse, walk, index, projection
  config/             — keystone.json (ProjectConfig + HookSpec)
  loader/             — policy cascade + verify
  policies/           — vendored policy resolver
  lockfile/           — install state
  manifest/           — policy manifest schema
  migrations/         — version-to-version upgrade steps
  scaffold/           — keystone init / bootstrap templates
  sensors/            — sensor runner registry (keystone-owned sensors only)
  adapters/           — per-host projection (claudecode, cursor, …)
  mcp/                — MCP server
  web/                — local dashboard
.keystone/harness/    — this project's own harness (dogfood)
.claude/              — projected from .keystone/harness/ — do not hand-edit
```

## Common commands

```
go build ./cmd/keystone        # build the CLI
go install ./cmd/keystone      # install on PATH (refresh after every Frontmatter or CLI change — hooks call the installed binary)
go test ./...                  # all tests
go vet ./...                   # static check (must pass before commit)
keystone index                 # regenerate .keystone/INDEX*.json
keystone project               # regenerate .claude/* + merge hooks into settings
keystone watch                 # long-running: re-project on every guide/sensor/persona save (300ms debounce)
keystone verify                # policy + cascade check
keystone verify --sensor <id>  # run one keystone-owned sensor (hook entry point)
```

> **Instant projection.** Run `keystone watch` in a side terminal while
> editing primitives. Any save under `.keystone/harness/` regenerates
> `INDEX.json` + `INDEX.lite.json` + `.claude/*` + `AGENTS.md` +
> `.cursor/rules/*` + `.aider.conf.yml` + `.continue/rules/*` within
> the debounce window. The next agent invocation sees the updated
> surface immediately — no manual `keystone project` needed.

> Hooks under `.claude/settings.json` invoke `keystone verify --sensor X`.
> If you see `unknown flag --sensor`, the installed binary is stale —
> `go install ./cmd/keystone` to refresh.

## Where things go

- **New rule / convention** → `.keystone/harness/guides/idioms/<topic>.md` + paired corpus
- **New automated check (computational)** → `.keystone/harness/sensors/<id>.md` + add a `hooks:` entry in `keystone.json`
- **New workflow** → `.keystone/harness/playbooks/<id>.md`
- **New review persona** → `.keystone/harness/personas/<id>.md`
- **New keystone subcommand** → `cmd/keystone/<name>.go` + wire in `root.go`

After any primitive add / move / delete: run `keystone index && keystone project`.

<!-- keystone:start -->
## Keystone harness

This project uses a **keystone harness**. The framework's primitives —
guides, corpus, sensors, actions, playbooks — plus host-native ones
(skills, subagents, commands, rules) all live under
[`.keystone/harness/`](.keystone/harness/). Discover what's available
through the index; open primitive bodies on demand.

**Read first:**
[`.keystone/INDEX.lite.json`](.keystone/INDEX.lite.json) — the cheap
discovery surface (kind + id + description only). Browse this to pick
which primitive you need. Then open the full
[`.keystone/INDEX.json`](.keystone/INDEX.json) for that primitive's
`path` / `globs` / `triggers`, and only open the body itself when you
decide to activate it.

**Activate by:**

| Kind         | When to open                                                            |
| ------------ | ----------------------------------------------------------------------- |
| **guide**    | Touched files match the entry's `globs:` (or no globs declared).        |
| **rule**     | Same as guide — host-native flavor (Cursor-style, plain directive).      |
| **corpus**   | A guide's `traces:` (or a prose forward-link) points at it.             |
| **action**   | User's intent matches `description` + `phase`. Open the playbook body. |
| **playbook** | Same as action — composed sequence of actions.                          |
| **sensor**   | Inside an action, per-phase, narrowed by `globs:`.                      |
| **skill**    | Claude Code auto-activates by `triggers:` match.                        |
| **subagent** | Spawn via the Task tool by `id`. The system prompt is the body.         |
| **command**  | Host slash mechanism: user types `/<id>`.                                |

**Lifecycle** — to kick off a unit of work, say "**run task on
`<ticket-id>`**" (runs the **task** playbook). For any single action,
ask in natural language ("run verify", "do a review pass") — the
action's body lives at its INDEX `path`.

**Iron laws** — non-negotiable across every phase. Distilled from
`guides/process/`; open the linked guide only when the rule is contested
or ambiguous.

- No proceeding without explicit acceptance criteria. ([`spec`](.keystone/harness/guides/process/spec.md))
- No completion claims without fresh verification — sensors must have
  run this turn, against the post-edit state, with cited tool output.
  ([`verification`](.keystone/harness/guides/process/verification.md),
  [`self-validation`](.keystone/harness/guides/process/self-validation.md))
- No commits with failing sensors. Never `--no-verify`.
- No AI attribution in commits, PRs, or tracker comments.
- No silent overwrites of state files.
- No reading or writing **sensitive files** — `.env*`, `*.pem`, `*.key`,
  `id_rsa`, `credentials.json`, `secrets/`, anything matching
  `*secret*`/`*credential*`/`*password*`. Ask the user out-of-band.
  ([`sensitive-files`](.keystone/harness/guides/process/sensitive-files.md))
- No **dangerous action** without explicit in-turn confirmation —
  `rm -rf`, `git push --force`, `git reset --hard`, prod DB writes,
  external comms (Slack/email), system installs, secret rotation. One
  confirmation, one action. Mode never loosens this.
  ([`dangerous-actions`](.keystone/harness/guides/process/dangerous-actions.md))
- No invented imports, methods, config keys, or CLI flags. Grep / read
  the manifest / check `--help` before referencing.
  ([`grounding`](.keystone/harness/guides/process/grounding.md))
- No "while I'm here" cleanups. Every changed line traces to the
  request. Style, formatting, renames, file moves get their own commit.
  ([`surgical-edits`](.keystone/harness/guides/process/surgical-edits.md))
- No accepting a subagent's "done" report as evidence. Read the diff;
  re-run sensors in the parent's turn.
  ([`subagent-trust`](.keystone/harness/guides/process/subagent-trust.md))

**Override** — your project files at `.keystone/harness/<kind>/<id>.md`
always win by default. Among installed policies, policies nested deeper
in `keystone.json` refine outer policies. A policy can mark an item
`strict` to make it absolute — nothing else can override a strict
item.
<!-- keystone:end -->
