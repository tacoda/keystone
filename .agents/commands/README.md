# Commands

Each file here is a command: a unit of work the user asks for by name, such as `/spec` or "run verify". A command does nothing until someone calls it. If the work should start from the task itself, write a skill instead.

## Shape

```markdown
---
id: spec
description: Capture intent and acceptance criteria for a unit of work.
args:
  - name: card
    type: string
    required: false
    description: The tracker card id.
---

# spec

The prompt the host gives the agent when the user calls the command.
```

- `id` is the name the user calls.
- `args` lists each parameter: `name`, `type`, `required`, and `description`.
- Commands are flat: one file per command, no topic folders.

## Add one

Run `keystone new command <id>`, fill in the body, then run `keystone index && keystone project`. Where the host has slash commands, the project step writes one, such as `.claude/commands/<id>.md`. Elsewhere, the command is listed in `AGENTS.md`. Edit this file, never the copy.

## The commands

| Command | File | When |
|---|---|---|
| **bootstrap** | [`bootstrap.md`](bootstrap.md) | Once per project: detect the stack, seed state, classify hooks. |
| **spec** | [`spec.md`](spec.md) | First on any task: capture intent and acceptance criteria. |
| **orient** | [`orient.md`](orient.md) | After spec: read `CODEBASE_STATE.md` and the idioms for the touched region, then sketch a plan. |
| **check-drift** | [`check-drift.md`](check-drift.md) | Between implementation and verify: a fast drift check on the diff. |
| **verify** | [`verify.md`](verify.md) | Before commit: lint, type-check, test, build, drift, and commit-message. |
| **review** | [`review.md`](review.md) | After verify: functional, security, risk, and deployment review, plus spec adherence. |
| **learn** | [`learn.md`](learn.md) | When something surprising happens: capture it to `.keystone/learning/inbox/`. |
| **audit** | [`audit.md`](audit.md) | Periodically: the learning and pruning flywheels together. Pruning writes to `.agents/references/state/charter-debt.md`. |
| **debt-review** | [`debt-review.md`](debt-review.md) | Periodically: triage `.agents/references/state/code-debt.md`. |
| **synthesize** | [`synthesize.md`](synthesize.md) | When the inbox has filled: promote candidates into rules and references. |
| **mode** | [`mode.md`](mode.md) | When pacing needs to change: paired, solo, or autopilot. |

The [`task`](../skills/task/SKILL.md) skill chains spec through review for one unit of work.
