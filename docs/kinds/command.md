# Command

<!-- voice-register: informational -->
<!-- voice-english: ste -->

A command is a unit of work the user asks for by name, such as `/spec` or `run verify`. Where the host has slash commands, keystone writes the command as one. Where it doesn't, keystone lists the command in `AGENTS.md`, and the agent opens the body when asked. The cost is that a command does nothing until someone calls it, so if the work should start from the task itself, write a [skill](skill.md).

## Example

```markdown
---
kind: command
id: spec
description: Capture intent and acceptance criteria for a unit of work.
args:
  - name: card
    type: string
    required: false
    description: The tracker card id to read the work from.
---

# spec

1. Read the card, or ask the user to describe the work.
2. Write the acceptance criteria as a numbered list.
3. Stop and ask the user to approve them.
```

Save your command as `.agents/commands/spec.md`, then run `keystone index && keystone project`. On Claude Code, the user then types `/spec`.

## Where it lives

```
.agents/commands/<id>.md
```

Commands are flat, with one file per command and no topic folders, because the id alone names the command across the charter. The body is the prompt. The host gives that prompt to the agent when the user calls the command by its id.

## Frontmatter

| Key | Required | Meaning |
|---|---|---|
| `kind` | no | `command`. Keystone infers it from the `commands/` folder. |
| `id` | yes | The name the user calls. The projected name is the id in kebab case. |
| `description` | yes | One line. Keystone copies it into each host's command file. |
| `args` | no | A list of `name`, `type`, `required`, and `description`. A bare string is short for `name`. |
| `model` | no | The model to run under, where the host supports it (opencode). |
| `phase` | no | The lifecycle phase. `keystone explain` prints it. |
| `produces` | no | The [document](document.md) ids this command writes. |
| `consumes` | no | The document ids this command reads. |
| `stop` | no | The done-condition, in plain words. |

`keystone lint` warns when a `produces` or `consumes` id names no document. That finding is a warning and not an error, because a policy installed later can supply the document.

## How each host gets it

| Host | What keystone writes |
|---|---|
| Claude Code | `.claude/commands/<id>.md`: the description, then the body. |
| Cursor | `.cursor/skills/<id>/SKILL.md` with `disable-model-invocation: true`, so only the user starts it. |
| GitHub Copilot | `.github/prompts/<id>.prompt.md`, with the args as `argument-hint`. |
| Continue | `.continue/prompts/<id>.md`, marked `invokable`. |
| opencode | `.opencode/commands/<id>.md`, with `model` when set. |
| Pi | `.pi/prompts/<id>.md`, with the args as `argument-hint`. |
| Codex, Cline, Goose, Aider | A line in `AGENTS.md`. The agent opens the body when the user asks by name. |

`keystone project` overwrites these files on every run. Edit the source instead. For the other kinds on each host, and for the sources behind each row, read the full matrix in [hosts.md](../hosts.md).

## What a command does not do

- It doesn't run a shell command, because the body is a prompt for the agent. To run a script, write a [tool](tool.md) or a [hook](hook.md).
- It doesn't fire on an event. Only the user starts it.
- Keystone doesn't check its args. Only GitHub Copilot and Pi get them, as an `argument-hint`.
- It doesn't create the documents it `produces`. Instead, the agent writes them under `.keystone/work/` as the command body tells it to.

## Failure modes

- **The command shadows a host built-in.** A command id can match a slash command the host already ships. The 5.0 plan flags `/review` on Claude Code as a likely clash (unverified), so pick an id the host doesn't use.
- **The host has no slash commands.** On Codex, Cline, Goose, and Aider, the user asks in plain words. The agent then finds the command in `AGENTS.md`.
- **A command and a skill do the same work.** The agent may pick either. Keep one, or make the skill call the command.

## Authoring

```
keystone new command <id>
```

This writes `.agents/commands/<id>.md` with one sample arg and a TODO description. Replace both before you run `keystone project`, or delete the arg if you don't need it.

Older keystone docs called a command an action. The kind and the `commands/` folder didn't change.
