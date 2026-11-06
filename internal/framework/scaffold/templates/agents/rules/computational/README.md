# Computational rules

A computational rule records a deterministic tool the stack already runs, such as a formatter, a linter, or a language server. It runs a command on an event instead of being read.

## Shape

```markdown
---
id: computational/gofmt
description: gofmt, canonical Go formatting.
mode: computational
event: PostToolUse
run: 'test -z "$(gofmt -l cmd internal)"'
---

# gofmt

What the tool covers, where its config lives, and how to treat a failure.
```

`keystone hook fire` runs any primitive with both `event:` and `run:`, so a computational rule fires the same way a hook does. A non-zero exit blocks. Without `event:`, the rule never runs.

A computational rule gets no host shim and no `AGENTS.md` line. Its `paths:` are documentation only, because the hook layer doesn't filter on them.

## Ships empty

The `bootstrap` command inventories the tools the stack supports and proposes one file per tool. A tool may need a config file, a binary, or an agent setting at install time. Offer that as a `keystone init` option, not as a file here.

## Anti-patterns

- A file for a tool the project doesn't run. List only what runs.
- A rule that duplicates a hook, such as an editor lint and a CI lint with different configs. Point both at one config.
- A tool signal the agent ignores. Honor it, or lower its authority on purpose.
