# Commands

Keystone ships no commands: its lifecycle steps (spec, orient, verify, review, …) are `keystone-*` skills. Each file here is a command: a unit of work the user asks for by name, such as `/release-notes`. A command does nothing until someone calls it. If the work should start from the task itself, write a skill instead.

## Shape

```markdown
---
id: release-notes
description: Draft release notes from the commits since the last tag.
args:
  - name: card
    type: string
    required: false
    description: The tracker card id.
---

# release-notes

The prompt the host gives the agent when the user calls the command.
```

- `id` is the name the user calls.
- `args` lists each parameter: `name`, `type`, `required`, and `description`.
- Commands are flat: one file per command, no topic folders.

## Add one

Run `keystone new command <id>`, fill in the body, then run `keystone index && keystone project`. Where the host has slash commands, the project step writes one, such as `.claude/commands/<id>.md`. Elsewhere, the command is listed in `AGENTS.md`. Edit this file, never the copy.

## The lifecycle skills

These used to be commands. Each is now a `keystone-*` skill that activates on its own or by name.

| Skill | File | When |
|---|---|---|
| **keystone-bootstrap** | [`SKILL.md`](../skills/keystone-bootstrap/SKILL.md) | Once per project: detect the stack, seed state, classify hooks. |
| **keystone-spec** | [`SKILL.md`](../skills/keystone-spec/SKILL.md) | First on any task: capture intent and acceptance criteria. |
| **keystone-orient** | [`SKILL.md`](../skills/keystone-orient/SKILL.md) | After spec: read `CODEBASE_STATE.md` and the idioms for the touched region, then sketch a plan. |
| **keystone-check-drift** | [`SKILL.md`](../skills/keystone-check-drift/SKILL.md) | Between implementation and verify: a fast drift check on the diff. |
| **keystone-verify** | [`SKILL.md`](../skills/keystone-verify/SKILL.md) | Before commit: lint, type-check, test, build, drift, and commit-message. |
| **keystone-review** | [`SKILL.md`](../skills/keystone-review/SKILL.md) | After verify: functional, security, risk, and deployment review, plus spec adherence. |
| **keystone-learn** | [`SKILL.md`](../skills/keystone-learn/SKILL.md) | When something surprising happens: capture it to `.keystone/learning/inbox/`. |
| **keystone-audit** | [`SKILL.md`](../skills/keystone-audit/SKILL.md) | Periodically: the learning and pruning flywheels together. Pruning writes to `.agents/references/state/charter-debt.md`. |
| **keystone-debt-review** | [`SKILL.md`](../skills/keystone-debt-review/SKILL.md) | Periodically: triage `.agents/references/state/code-debt.md`. |
| **keystone-synthesize** | [`SKILL.md`](../skills/keystone-synthesize/SKILL.md) | When the inbox has filled: promote candidates into rules and references. |
| **keystone-mode** | [`SKILL.md`](../skills/keystone-mode/SKILL.md) | When pacing needs to change: paired, solo, or autopilot. |

The [`task`](../skills/task/SKILL.md) skill chains spec through review for one unit of work.
