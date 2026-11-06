---
kind: corpus
id: corpus/idioms/go/README
description: 'Go idiom set for the keystone project.'
---
# Go idioms — reasoning

Stack-specific patterns for Go code in this repo. Each idiom in this folder is paired with a rules file at [`../../../guides/idioms/go/`](../../../guides/idioms/go/).

## Region

Go source lives in:

- `cmd/keystone/` — CLI entrypoint, cobra wiring.
- `internal/framework/` — the framework implementation. Everything that isn't the CLI shell.

## Stack character

- Single-binary CLI built around `spf13/cobra`.
- Reads / writes filesystem (`.charter/`, project state). No DB, no HTTP server. (MCP server uses stdio.)
- Library deps kept small. Stdlib first, single-purpose third-party additions only.

## Layout

- `stdlib-first.md` — prefer stdlib over third-party for primitives that exist there.
- (more idioms accumulate via the **learn** → **synthesize** flywheel.)
