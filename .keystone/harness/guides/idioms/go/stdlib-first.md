---
kind: guide
id: guides/idioms/go/stdlib-first
description: 'Prefer Go stdlib over new third-party deps.'
globs:
  - "cmd/**/*.go"
  - "internal/**/*.go"
  - "go.mod"
  - "go.sum"
tags:
  - go
  - dependencies
---
# Stdlib first — rules

The rules from [`corpus/idioms/go/stdlib-first.md`](../../../corpus/idioms/go/stdlib-first.md).

## IRON LAW

- **No new direct dep added without naming the stdlib (or already-vendored lib) it replaces and why that option is insufficient.** Recorded in the commit message body.

## GOLDEN RULE

- Filesystem / path → `io/fs`, `path/filepath`, `os`.
- Strings / bytes / parsing → `strings`, `bytes`, `bufio`, `regexp`.
- Serialization → `encoding/json`, `gopkg.in/yaml.v3` (already vendored). One YAML parser, period.
- CLI commands → `github.com/spf13/cobra` (already vendored).
- MCP server → `github.com/mark3labs/mcp-go` (already vendored).
- Filesystem watching → `github.com/fsnotify/fsnotify` (already vendored).
- Terminal I/O → `golang.org/x/term` (already vendored).
- Concurrency → `sync`, `context`, channels.

## Anti-patterns

- Two libraries doing the same job (e.g., a second YAML parser, a second CLI framework).
- Pulling a dep in for one helper. Copy 5 lines instead and attribute the source.
- Wrapping stdlib in a thin abstraction "for testability" before there is a second implementation.
