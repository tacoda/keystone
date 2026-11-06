---
kind: guide
mode: computational
event: PostToolUse
run: 'command -v gofmt >/dev/null 2>&1 && test -z "$(gofmt -l cmd internal)" || true'
id: guides/computational/gofmt
description: 'gofmt — canonical Go formatting.'
globs:
  - "cmd/**/*.go"
  - "internal/**/*.go"
---
# gofmt

**What it covers** — canonical Go source layout: indentation, brace placement, import grouping/ordering, trailing whitespace.
**Activation** — `go fmt ./...` on demand; editor on-save (`gopls` runs it).
**Authority** — blocking. Diffs that contain unformatted Go must not commit.
**Configured by** — none (gofmt is opinionless and unconfigured).
