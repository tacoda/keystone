---
kind: guide
mode: computational
event: PostToolUse
run: 'command -v go >/dev/null 2>&1 && go vet ./... || true'
id: guides/computational/go-vet
description: 'go vet — suspicious-construct check.'
globs:
  - "cmd/**/*.go"
  - "internal/**/*.go"
---
# go vet

**What it covers** — suspicious constructs: printf-argument mismatches, copylocks, shadowed variables, unreachable code, struct-tag typos, common goroutine bugs.
**Activation** — `go vet ./...` from the **lint** sensor. Runs in `verify`.
**Authority** — blocking when invoked by the lint sensor; advisory otherwise.
**Configured by** — none (vet has no project config; behavior is per Go version).
