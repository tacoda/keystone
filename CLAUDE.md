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
keystone signal fire <name>    # dispatch primitives subscribed (on:) to a signal (hook fire = alias)
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
- **New computational check** → `.charter/sensors/<id>.md` (`mode: computational`, `on:` + `run:`)
- **New review** → `.charter/sensors/<id>.md` (`mode: inferential`, `on:` + `returns:`) or a reviewer `.charter/agents/<id>.md`
- **New side-effect** → `.charter/tools/<id>.md` (`transport:` + `on:`)
- **New workflow** → `.charter/playbooks/<id>.md`
- **New keystone subcommand** → `cmd/keystone/<name>.go` + wire in `root.go`

After any primitive add / move / delete: run `keystone index && keystone project`.

<!-- keystone:start -->
@CHARTER.md

You **must** read [`CHARTER.md`](CHARTER.md) before doing anything in this repo — it carries the iron laws and the ambient rules that govern the charter. The import above loads it; do not proceed without it.

## On this host — Claude Code

- **Subagents** — spawn charter agents (`.charter/agents/`) as subagents via the Task tool for review/scout work.
- **Slash commands** — charter commands and playbooks surface as `/keystone-<id>`.
- **Skills** — auto-activate by their `triggers:`.
- **Hooks** — charter hooks fire automatically on Claude Code lifecycle events.
<!-- keystone:end -->
