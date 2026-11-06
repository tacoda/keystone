# Charter

This folder holds the project's charter: the standards that shape how a coding agent works in this repo. You own every file here. Version it with the code and edit it freely.

Keystone keeps its own state in `.keystone/`. Hosts get generated copies (`.claude/`, `.cursor/`, the managed block in `AGENTS.md`). Edit the source here, never a generated file.

## Two categories, ten kinds

Each file is a primitive. Keystone reads its kind from the first folder under `.agents/`, so `kind:` is optional.

A **constraint** limits, gates, or checks what the agent does:

| Kind | Folder | What it is |
|---|---|---|
| `rule` | [`rules/`](rules/README.md) | A directive. It applies when a touched file matches its `paths:`, or everywhere without them. |
| `reference` | [`references/`](references/README.md) | The reasoning behind a rule. The agent opens it on demand. |
| `hook` | [`hooks/`](hooks/README.md) | A check that runs on an event. A non-zero exit blocks. |
| `permission` | `permissions/` | Allow, ask, and deny rules for tools. |
| `eval` | [`evals/`](evals/README.md) | A test that the charter changes agent behavior. |

A **capability** adds something the agent can do:

| Kind | Folder | What it is |
|---|---|---|
| `skill` | [`skills/`](skills/README.md) | A `SKILL.md` folder. It activates when its description matches the task. |
| `command` | [`commands/`](commands/README.md) | A unit of work the user asks for by name. |
| `subagent` | [`agents/`](agents/README.md) | A role the agent delegates to. |
| `tool` | `tools/` | A callable you write, over MCP, as a command, or as a host plugin. |
| `document` | `documents/` | A template for governed output, such as a plan or an ADR. |

A fresh install has no `permissions/`, `tools/`, or `documents/` folder. Run `keystone new <kind> <name>` to add the first one.

## Every primitive shares

```yaml
---
id: process/verification     # stable id, required
description: One line.       # required; the agent reads it in the index
corpus: [process/verification]  # optional: references that hold the why
---
```

The agent reads `.keystone/INDEX.lite.json` first, then opens a body only when it applies. After you add, move, or delete a primitive, run `keystone index && keystone project`. Or run `keystone watch` in a side terminal.

## Rules and references

A rule commands; a reference explains. Keep the rule short, because the agent reads it often. Put the history, trade-offs, and sources in a reference, and list its id under the rule's `corpus:`.

Rules have three tiers. Most rules are plain rules. A golden rule is a strong standard, and deviation needs a stated reason. An iron law is non-negotiable and rare by design. Mark the tier with `tier:` and with a `## Rules`, `## Golden rule`, or `## Iron law` section.

## Two flywheels

- **Learning adds.** The `learn` command writes a candidate to [`.keystone/learning/inbox/`](../.keystone/learning/README.md). The `synthesize` command promotes it into a rule, a reference, or both.
- **Pruning removes.** The `audit` command finds rules the code no longer follows. Stale files move to [`.keystone/archive/`](../.keystone/archive/README.md) with the reason.

## Lifecycle

Say "run task on `<ticket-id>`" to start a unit of work. The [`task`](skills/task/SKILL.md) skill walks spec, orient, implementation, check-drift, verify, and review. To run one step, ask for its command by name ("run verify").

Pacing has three modes: paired (the default), solo, and autopilot. Switch with the `mode` command.

## Policies

A policy is shared charter content, vendored read-only into `.keystone/policies/<name>/`. Declare policies in `keystone.json`. A fresh install declares none. Your files here override a policy, unless the policy lists the item under `strict:`.

## Assumptions

The charter works best when the project has these. Without one, the matching phase asks you instead.

- A way to track work, from an issue tracker to a `TODO.md`.
- Lint, test, and build commands for the verify phase.
- A pull request process for the review phase.
- A CI pipeline for the release phase.
