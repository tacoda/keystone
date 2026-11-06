---
kind: rule
id: idioms/go/README
description: 'Go idiom rules, entry point.'
paths:
  - "cmd/**/*.go"
  - "internal/**/*.go"
---
# Go idiom rules

Rules for Go code in this repo. Each pairs with a reference in [`references/idioms/go/`](../../../references/idioms/go/), which holds the reasoning and examples.

## Activation

`paths:` above scopes these rules. They apply when a touched file matches `cmd/**/*.go` or `internal/**/*.go`.

## Rule files

- [`stdlib-first.md`](stdlib-first.md): prefer the stdlib over new deps.

More files accumulate through the **learn** and **synthesize** flywheel.
