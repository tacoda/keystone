---
kind: guide
id: guides/idioms/go/README
description: 'Go idiom rules — entry point.'
globs:
  - "cmd/**/*.go"
  - "internal/**/*.go"
---
# Go idiom rules

Rules for Go code in this repo. Paired with [`../../../corpus/idioms/go/`](../../../corpus/idioms/go/) (reasoning, examples).

## Activation

`globs:` above narrows ambient stack activation. These rules fire when a touched file matches `cmd/**/*.go` or `internal/**/*.go`.

## Rule files

- [`stdlib-first.md`](stdlib-first.md) — prefer stdlib over new deps.

(More files accumulate via the **learn** → **synthesize** flywheel.)
