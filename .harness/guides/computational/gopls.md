---
kind: guide
mode: computational
event: PostToolUse
run: 'command -v gopls >/dev/null 2>&1 && gopls version >/dev/null 2>&1 || true'
id: guides/computational/gopls
description: 'gopls — Go LSP for in-editor signal.'
globs:
  - "cmd/**/*.go"
  - "internal/**/*.go"
---
# gopls

**What it covers** — type errors, unused imports / variables, refactor hints, completions, hover docs, jump-to-definition, on-save formatting (delegates to gofmt).
**Activation** — LSP in the editor; the Serena integration in this repo (`.serena/`) drives semantic symbol queries on top of gopls.
**Authority** — advisory in real time, blocking once the same error surfaces through `go build ./...` (type-check sensor).
**Configured by** — none in this repo (defaults). Per-editor settings live outside the repo.
