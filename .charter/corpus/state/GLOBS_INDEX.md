---
kind: corpus
id: corpus/state/GLOBS_INDEX
description: 'Reverse-index of every guide that declares globs: in its frontmatter.'
---
# Globs Index

> **Generated.** The **bootstrap** action seeds this from the region map in `CODEBASE_STATE.md`; **synthesize** and **audit** regenerate it whenever a guide's `globs:` frontmatter changes. **Do not edit by hand** — manual edits are overwritten on the next regeneration.

Reverse-index of every guide that declares `globs:` in its frontmatter. Pointer-style adapters (Claude Code, Codex, Aider, Continue, etc.) read this in their action playbooks to gate idiom loading on the touched-files set without re-walking the tree.

Guides without `globs:` are not listed here — they activate ambient per their topic default.

## Index

| Glob pattern | Guides claiming it |
|---|---|
| `.charter/**/*.md` | `.charter/guides/idioms/charter-content/README.md`, `.charter/guides/idioms/charter-content/primitive-shape.md` |
| `.charter/corpus/state/**/*.md` | `.charter/guides/idioms/charter-content/state-files.md` |
| `cmd/**/*.go` | `.charter/guides/computational/gofmt.md`, `.charter/guides/computational/go-vet.md`, `.charter/guides/computational/gopls.md`, `.charter/guides/idioms/go/README.md`, `.charter/guides/idioms/go/stdlib-first.md` |
| `go.mod` | `.charter/guides/idioms/go/stdlib-first.md` |
| `go.sum` | `.charter/guides/idioms/go/stdlib-first.md` |
| `internal/**/*.go` | `.charter/guides/computational/gofmt.md`, `.charter/guides/computational/go-vet.md`, `.charter/guides/computational/gopls.md`, `.charter/guides/idioms/go/README.md`, `.charter/guides/idioms/go/stdlib-first.md` |
| `internal/framework/scaffold/templates/**/*.md` | `.charter/guides/idioms/charter-content/README.md`, `.charter/guides/idioms/charter-content/primitive-shape.md` |
| `internal/framework/scaffold/templates/**/corpus/state/**/*.md` | `.charter/guides/idioms/charter-content/state-files.md` |

## How it's regenerated

1. Walk every guide under `.charter/guides/` and `.charter/policies/*/guides/`.
2. For each guide, read its frontmatter and collect each entry in `globs:` (if present).
3. Invert: for each glob pattern, list the guides that claim it.
4. Sort patterns by path-prefix for stable diffs.
5. Replace the **Index** table above; touch nothing else.

## Consumers

| Adapter | How it uses the index |
|---|---|
| Claude Code | `orient` reads the index to load only the idiom guides whose globs match touched files. |
| Codex | Same as Claude Code, via `.charter/adapters/codex/activation.md`. |
| Aider / Cline / Continue / Goose / Pi | Same pointer-style pattern; per-adapter `activation.md` describes the lookup. |
| Cursor | Does not read this file — it uses native `globs:` on `.cursor/rules/*.mdc`. |
| `_generic` | Does not read this file — falls back to topic defaults. |
