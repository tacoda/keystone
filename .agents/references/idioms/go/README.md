---
kind: reference
id: idioms/go/README
description: 'Go idiom set for the keystone project.'
---
# Go idioms, reasoning

Stack-specific patterns for Go code in this repo. Each idiom here pairs with a rule file in [`rules/idioms/go/`](../../../rules/idioms/go/).

## Region

Go source lives in:

- `cmd/keystone/`: the CLI entrypoint and cobra wiring.
- `internal/framework/`: the framework implementation, everything outside the CLI shell.

## Stack character

- A single-binary CLI built on `spf13/cobra`.
- It reads and writes the filesystem (`.agents/`, `.keystone/`). No DB and no HTTP server. The MCP server uses stdio.
- Library deps stay small. Stdlib first, then single-purpose third-party additions only.

## Layout

- `stdlib-first.md`: prefer the stdlib over third-party packages for anything it covers.
- More idioms accumulate through the **learn** and **synthesize** flywheel.
