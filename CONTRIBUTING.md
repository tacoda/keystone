# Contributing to keystone

Thanks for working on keystone. Brief, opinionated guide — read it
once, then file issues / open PRs.

## What keystone is

Keystone is an **agent harness framework**. The harness — a tree of
typed primitives at `.keystone/harness/` — is the product. The CLI
and the MCP server are tooling that helps you author and serve that
harness, but neither is required at runtime. Once a project has a
harness, any agent that reads `.keystone/INDEX.json` and the
primitive bodies it points at can operate on it.

That framing affects the work:

- **Harness shape changes are big deals.** The on-disk layout +
  frontmatter schema are the public contract. Breaking changes ship
  in a major release (e.g. 2.0) with a migrator.
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
  primitive/                   # canonical primitive types + walker + indexer + linter + projection
  scaffold/                    # embedded templates (init writes from here)
  mcp/                         # Go MCP server (tools, prompts, resources, adapters)
  web/                         # localhost dashboard + REST API + SSE push
  budget/                      # token estimates per port
docs/
  ports/                       # one .md per port / kind contract
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
  (`keystone.json`, `.keystone/lockfile.json`, `.keystone/context.json`,
  `.keystone/INDEX.json`, patches/*.json). YAML inside markdown
  frontmatter is fine (markdown convention; the parser accepts JSON
  there too).

## Adding a primitive kind

The bar is higher than adding a tool or a command. Before opening a
PR:

1. Update `docs/ports/primitive.md` with the new `kind` row.
2. Add the constant + entry in `internal/framework/primitive/primitive.go`.
3. Extend `walk.go` with the scan root.
4. Decide projection target (if any) and add to `project.go`.
5. Write a generator in `cmd/keystone/new_<kind>.go`.
6. Add a `keystone:new-<kind>` skill.
7. Update the menu templates so the dashboard + agent know about the
   new kind.

Two-layer taxonomy still applies — framework abstraction (guide,
corpus, sensor, action, playbook) vs. agent abstraction (rule, skill,
subagent, command). Pick the right layer.

## Filing issues

- **Bug.** Reproduction steps, expected vs. actual, what you ran on
  (OS, Go version, keystone version).
- **Feature.** Describe the user-visible behavior change first. If
  the harness contract changes, say so up front — it changes the
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

Tag-driven via goreleaser. Maintainers run `git tag v2.x.y && git
push --tags`; CI does the rest. No `gh release create`.

## License

MIT. By contributing, you agree your changes ship under the same
license. See [`LICENSE`](LICENSE).
