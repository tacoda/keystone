# Hooks

Each file here is a hook: a check that runs on an event and can block the agent. A computational hook runs a shell command, and a non-zero exit blocks. An inferential hook asks an agent for a verdict.

## Shape

```markdown
---
id: build
description: The project's build step.
mode: computational    # or inferential
event: pre-verify      # a host event or a signal
run: go build ./...    # computational only
---

# build

What a failure means and how to fix it.
```

- `event` is a host event (`PreToolUse`, `PostToolUse`, `Stop`, and others) or a signal. Without it, the hook never fires.
- `run` is required for a computational hook. `returns` is required for an inferential one.
- `keystone verify` fires the `pre-verify` signal before it checks, and a failing hook there blocks. Fire any signal with `keystone signal fire <name>`.

## The shipped hooks

The computational hooks (`lint`, `type-check`, `test`, `build`, `coverage`, `secret-scan`, `vuln-scan`, `sast`, `stack-drift`) fire on `pre-verify`. `commit-message` fires on `PreToolUse`. Each one ships with a TODO `run:` until you put in the project's command. The `bootstrap` command records those commands in `.agents/references/state/CODEBASE_STATE.md`.

The inferential hooks (the `review-*` set, `drift`, `spec-adherence`, and others) declare no `event:`. The `verify`, `review`, and `audit` commands run them.

## Add one

Run `keystone new hook <id>`, then run `keystone index && keystone project`. Each host bridge calls the `keystone` binary on your `PATH`, so every machine that runs agents needs it installed.
