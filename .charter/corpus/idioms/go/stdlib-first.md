---
kind: corpus
id: corpus/idioms/go/stdlib-first
description: 'Prefer Go stdlib for primitives that exist there.'
---
# Stdlib first

One-paragraph statement: Reach for Go's standard library before adding a dependency. The stdlib covers file I/O, path manipulation, JSON/YAML (via `gopkg.in/yaml.v3` — already vendored), HTTP, templating, regexp, and goroutine primitives. New direct deps are added only when the stdlib lacks the capability or the alternative is materially smaller / safer for the surface this CLI needs.

> **Rules extracted:** [`guides/idioms/go/stdlib-first.md`](../../../guides/idioms/go/stdlib-first.md).

## How to apply

- Filesystem walks → `io/fs`, `path/filepath`, `os`. Not a third-party walker.
- JSON / YAML → `encoding/json` and the already-vendored `gopkg.in/yaml.v3`. Do not introduce a second YAML lib.
- CLI flags / commands → `spf13/cobra` (already vendored); plain stdlib `flag` for tiny single-binary tools is fine inside `cmd/keystone/` if a subcommand needs nothing more.
- Terminal I/O → `golang.org/x/term` (already vendored). No `tview` / `bubbletea` unless a feature genuinely needs full TUI.
- Concurrency → `sync`, `context`, channels. No external worker pools.

## Anti-patterns

- Adding a third-party library where 10 lines of stdlib code (`filepath.Walk`, `regexp.MustCompile`, `bufio.Scanner`) cover the case.
- Two libraries that do the same thing (two YAML parsers, two CLI frameworks, two loggers).
- Pulling in a dep for a single helper function — copy 5 lines instead, attribute the source.

## Review checklist

- Every new direct dep in `go.mod` justified by a capability the stdlib lacks?
- Existing vendored libs (`cobra`, `mcp-go`, `fsnotify`, `yaml.v3`, `x/term`) used where they fit, not duplicated?
- Indirect-dep growth proportional to the new direct dep's surface?

**Traces to:** the project's "simplicity first" / "manage complexity ruthlessly" principle in `/Users/tacoda/.claude/CLAUDE.md` and the lazy-senior-dev posture in the user's CLAUDE.md.
