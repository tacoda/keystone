---
kind: reference
id: corpus/state/GLOBS_INDEX
description: 'Reverse-index of every rule that declares paths: in its frontmatter.'
---
# Globs Index

> **Generated.** The **bootstrap** command seeds this from the region map in `CODEBASE_STATE.md`; **synthesize** and **audit** regenerate it whenever a rule's `paths:` frontmatter changes. **Do not edit by hand**: manual edits are overwritten on the next regeneration.

Reverse-index of every rule that declares `paths:` in its frontmatter. The **orient** command reads it to load only the rules that match the touched files, without walking the tree again.

Rules without `paths:` are not listed here. They apply everywhere.

## Index

| Path pattern | Rules claiming it |
|---|---|
| `.agents/**/*.md` | `.agents/rules/idioms/charter-content/README.md`, `.agents/rules/idioms/charter-content/primitive-shape.md` |
| `.agents/references/state/**/*.md` | `.agents/rules/idioms/charter-content/state-files.md` |
| `cmd/**/*.go` | `.agents/rules/computational/gofmt.md`, `.agents/rules/computational/go-vet.md`, `.agents/rules/computational/gopls.md`, `.agents/rules/idioms/go/README.md`, `.agents/rules/idioms/go/stdlib-first.md` |
| `go.mod` | `.agents/rules/idioms/go/stdlib-first.md` |
| `go.sum` | `.agents/rules/idioms/go/stdlib-first.md` |
| `internal/**/*.go` | `.agents/rules/computational/gofmt.md`, `.agents/rules/computational/go-vet.md`, `.agents/rules/computational/gopls.md`, `.agents/rules/idioms/go/README.md`, `.agents/rules/idioms/go/stdlib-first.md` |
| `internal/framework/scaffold/templates/**/*.md` | `.agents/rules/idioms/charter-content/README.md`, `.agents/rules/idioms/charter-content/primitive-shape.md` |
| `internal/framework/scaffold/templates/**/references/state/**/*.md` | `.agents/rules/idioms/charter-content/state-files.md` |

## How it's regenerated

1. Walk every rule under `.agents/rules/` and `.keystone/policies/*/rules/`.
2. For each rule, read its frontmatter and collect each entry in `paths:` (if present).
3. Invert: for each path pattern, list the rules that claim it.
4. Sort patterns by path-prefix for stable diffs.
5. Replace the **Index** table above; touch nothing else.

## Consumers

| Host | How it uses the index |
|---|---|
| Any host | The **orient** command reads the index to load only the rules whose `paths:` match touched files. |
| Claude Code, Cursor, Copilot, Continue, Cline | Also load each rule with `paths:` natively, through the shim `keystone project` writes. |
