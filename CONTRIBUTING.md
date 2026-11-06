# Contributing to keystone

Thanks for working on keystone. Brief, opinionated guide — read it
once, then file issues / open PRs.

## What keystone is

Keystone is the **agent charter framework**. The charter, a tree
of typed primitives at `.agents/`, is the product. A **harness** is
the engine that runs the model (Claude Code, the orchestrator, the
runner); Keystone is not a harness, it manages the charter that
constrains one. The CLI and the MCP server are tooling that helps you
author and serve that charter, but neither is required at runtime. Once
a project has a charter, any agent that reads `.keystone/INDEX.json` and
the primitive bodies it points at can operate on it. (See
[`GLOSSARY.md`](GLOSSARY.md) for the charter/harness distinction.)

That framing affects the work:

- **Charter shape changes are big deals.** The on-disk layout +
  frontmatter schema are the public contract. Breaking changes ship
  in a major release (e.g. 4.0) with a migrator.
- **CLI + MCP server are interfaces.** Adding tools / commands is
  low-friction; changing what a primitive *is* requires more care.

## Setup

```
go install ./cmd/keystone
keystone --help
```

Go 1.25+. No other prerequisites for the core. The web subcommand
embeds htmx and a small CSS file — no JS toolchain.

## Local dev loop

```
go build ./...
go test ./...
go run ./cmd/keystone --help
```

Smoke an end-to-end:

```
WORK=$(mktemp -d)
go run ./cmd/keystone init "$WORK" --agent codex
go run ./cmd/keystone index --dir "$WORK"
go run ./cmd/keystone lint --dir "$WORK" --verbose
go run ./cmd/keystone web serve --dir "$WORK" --port 4773
```

## Project layout

```
cmd/keystone/                  # Cobra CLI entry points; one verb per file
internal/framework/
  config/                      # keystone.json schema + paths
  lockfile/                    # .keystone/lockfile.json
  loader/                      # cascade verifier + dependency graph
  manifest/                    # keystone-policy.json (per-policy)
  patch/                       # patch op model + apply
  policies/                    # vendored-policy install/verify/reset
  primitive/                   # canonical primitive types + walker + indexer + linter
  projection/                  # per-host profiles; writes host files from .agents/
  scaffold/                    # embedded templates (init writes from here)
  mcp/                         # Go MCP server (tools, prompts, resources, adapters)
  web/                         # localhost dashboard + REST API + SSE push
  budget/                      # token estimates per host and per kind
docs/
  kinds/                       # one .md per primitive kind contract
  internals/                   # budget, patch, state ledger, flywheel sink, host docs
  conventions.md               # the cross-cutting layout + rules
```

## Style + conventions

- **Comments.** Comments explain *why*, not *what*. Don't narrate
  obvious code. Don't reference PRs or tickets in code comments.
- **Naming.** Domain vocabulary. No "Manager" / "Helper" /
  "Processor" classes. Functions are verbs.
- **Errors.** Wrap with context: `fmt.Errorf("read %s: %w", path,
  err)`. Don't `panic` outside of `init()` / `main()` boot paths.
- **Tests.** Table-driven where natural. Tests live next to the code
  (`*_test.go`). New behavior comes with a test.
- **Frontmatter.** Every shipped template primitive declares
  canonical primitive frontmatter (`kind`, `id`, `description`, plus
  per-kind required fields). `keystone lint` is the gate.
- **No YAML files** keystone owns. Standalone config is JSON
  (`keystone.json`, `.keystone/lockfile.json`,
  `.keystone/INDEX.json`, patches/*.json). YAML inside markdown
  frontmatter is fine (markdown convention; the parser accepts JSON
  there too).

## Adding a primitive kind

The bar is higher than adding a tool or a command. Before opening a
PR:

1. Update `docs/kinds/README.md` with the new `kind` row.
2. Add the constant, its category, and its folder in
   `internal/framework/primitive/primitive.go`.
3. Decide how each host gets it (if at all) in
   `internal/framework/projection/hosts.go`.
4. Write a generator and wire it into `cmd/keystone/new.go`.
5. Add the kind to the `keystone-new` skill.
6. Add the kind to the dashboard's kind lists in
   `internal/framework/web/`.

There are 10 kinds in two categories. Constraints are `rule`, `reference`,
`hook`, `permission`, and `eval`. Capabilities are `skill`, `command`,
`subagent`, `tool`, and `document`. A hook, tool, or subagent names what
fires it with `event:`. Pick the kind by what the primitive *is*, not
how it fires.

## Filing issues

- **Bug.** Reproduction steps, expected vs. actual, what you ran on
  (OS, Go version, keystone version).
- **Feature.** Describe the user-visible behavior change first. If
  the charter contract changes, say so up front — it changes the
  review bar.

## Pull requests

- Small + focused. One concern per PR. Multiple unrelated changes
  invite review fatigue + churn.
- Tests included where they apply. `go test ./...` clean.
- Don't bump the major version in a feature PR. Major bumps ship
  with a migrator.
- Don't add AI-attribution lines to commits. Don't reference the
  agent that wrote a change.

## Releases

Tag-driven via goreleaser. Maintainers run `git tag v5.x.y && git
push --tags`; CI does the rest. No `gh release create`.

## License

MIT. By contributing, you agree your changes ship under the same
license. See [`LICENSE`](LICENSE).
