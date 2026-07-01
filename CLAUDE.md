# Keystone — the coding-agent charter manager

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
.charter/    — this project's own charter (dogfood)
.claude/              — projected from .charter/ — do not hand-edit
```

## Common commands

```
go build ./cmd/keystone        # build the CLI
go install ./cmd/keystone      # install on PATH (refresh after every Frontmatter or CLI change — hooks call the installed binary)
go test ./...                  # all tests
go vet ./...                   # static check (must pass before commit)
keystone index                 # regenerate .charter/INDEX*.json
keystone project               # regenerate .claude/* + merge hooks into settings
keystone watch                 # long-running: re-project on every guide/sensor/agent save (300ms debounce)
keystone verify                # policy + cascade check (auto-fires pre/post-verify hooks)
keystone hook fire <event>     # dispatch framework hooks bound to an event
```

> **Instant projection.** Run `keystone watch` in a side terminal while
> editing primitives. Any save under `.charter/` regenerates
> `INDEX.json` + `INDEX.lite.json` + `.claude/*` + `AGENTS.md` +
> `.cursor/rules/*` + `.aider.conf.yml` + `.continue/rules/*` within
> the debounce window. The next agent invocation sees the updated
> surface immediately — no manual `keystone project` needed.

> Hooks under `.claude/settings.json` are a single bridge per host phase
> (`keystone hook fire <phase>`) that dispatches the matching framework
> hooks. Refresh the installed binary (`go install ./cmd/keystone`) after
> any frontmatter/CLI change — the hooks call it.

## Where things go

- **New rule / convention** → `.charter/guides/idioms/<topic>.md` + paired corpus
- **New computational check** → `.charter/hooks/<id>.md` (`event:` + `run:`)
- **New review** → `.charter/sensors/<id>.md` (`mode: inferential`, `returns:`) or a reviewer `.charter/agents/<id>.md`
- **New workflow** → `.charter/playbooks/<id>.md`
- **New keystone subcommand** → `cmd/keystone/<name>.go` + wire in `root.go`

After any primitive add / move / delete: run `keystone index && keystone project`.

<!-- keystone:start -->
## Keystone harness

This project uses a **keystone harness**. Its primitives — guides,
sensors, hooks, agents, commands, skills, playbooks, patterns, corpus,
documents, concerns, posture, tools — all live under
[`.harness/`](.harness/). Discover what's available through the index;
open primitive bodies on demand.

**Read first:**
[`.harness/INDEX.lite.json`](.harness/INDEX.lite.json) — the cheap
discovery surface (kind + id + description only). Browse this to pick
which primitive you need. Then open the full
[`.harness/INDEX.json`](.harness/INDEX.json) for that primitive's
`path` / `globs` / `triggers`, and only open the body itself when you
decide to activate it.

**Activate by:**

| Kind         | When to open                                                            |
| ------------ | ----------------------------------------------------------------------- |
| **guide**    | Touched files match the entry's `globs:`. Inferential → a directive; computational → a host hook (LSP). |
| **corpus**   | A guide's `corpus:` (or a prose forward-link) points at it — the *why*. |
| **command**  | User's intent matches `description` + `phase`; a unit of work. Host slash: `/keystone-<id>`. |
| **playbook** | A composed sequence of commands with human `gates:`.                    |
| **sensor**   | An inferential review at a gate — dispatched as an agent, returns a `returns:` verdict. |
| **hook**     | Fires deterministically on an `event:` (host phase or framework event) → `run:` shell or `agent:` dispatch. |
| **skill**    | Claude Code auto-activates by `triggers:` match.                        |
| **agent**    | Spawn via the Task tool by `id`. The system prompt is the body.         |
| **pattern**  | A reusable documentation pattern (Diátaxis) — apply when writing docs.  |

**Lifecycle** — to kick off a unit of work, say "**run task on
`<ticket-id>`**" (runs the **task** playbook). For any single command,
ask in natural language ("run verify", "do a review pass") — the
command's body lives at its INDEX `path`.

**Iron laws** — non-negotiable across every phase. Distilled from
`guides/process/`; open the linked guide only when the rule is contested
or ambiguous.

- No proceeding without explicit acceptance criteria. ([`spec`](.harness/guides/process/spec.md))
- No completion claims without fresh verification — checks must have
  run this turn, against the post-edit state, with cited tool output.
  ([`verification`](.harness/guides/process/verification.md),
  [`self-validation`](.harness/guides/process/self-validation.md))
- No commits with failing checks. Never `--no-verify`.
- No AI attribution in commits, PRs, or tracker comments.
- No silent overwrites of state files.
- No reading or writing **sensitive files** — `.env*`, `*.pem`, `*.key`,
  `id_rsa`, `credentials.json`, `secrets/`, anything matching
  `*secret*`/`*credential*`/`*password*`. Ask the user out-of-band.
  ([`sensitive-files`](.harness/guides/process/sensitive-files.md))
- No **dangerous action** without explicit in-turn confirmation —
  `rm -rf`, `git push --force`, `git reset --hard`, prod DB writes,
  external comms (Slack/email), system installs, secret rotation. One
  confirmation, one action. Mode never loosens this.
  ([`dangerous-actions`](.harness/guides/process/dangerous-actions.md))
- No invented imports, methods, config keys, or CLI flags. Grep / read
  the manifest / check `--help` before referencing.
  ([`grounding`](.harness/guides/process/grounding.md))
- No "while I'm here" cleanups. Every changed line traces to the
  request. Style, formatting, renames, file moves get their own commit.
  ([`surgical-edits`](.harness/guides/process/surgical-edits.md))
- No accepting a subagent's "done" report as evidence. Read the diff;
  re-run checks in the parent's turn.
  ([`subagent-trust`](.harness/guides/process/subagent-trust.md))

**Override** — your project files at `.harness/<kind>/<id>.md`
always win by default. Among installed policies, policies nested deeper
in `keystone.json` refine outer policies. A policy can mark an item
`strict` to make it absolute — nothing else can override a strict
item.
<!-- keystone:end -->
