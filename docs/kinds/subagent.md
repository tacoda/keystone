# Subagent

<!-- voice-register: informational -->
<!-- voice-english: ste -->

A subagent is a role the agent delegates to, such as a reviewer or a planner. Each one has its own system prompt and tool list. It's a capability, because the main agent can hand a narrow job to it and get a short report back. The cost is that you must trust that report less than your own checks, and only some hosts enforce the tool list.

## Example

```markdown
---
kind: subagent
id: drift-reviewer
description: Flags places the current diff drifts from loaded charter rules.
tools:
  - Read
  - Grep
  - Bash
model: sonnet
---

# drift-reviewer

Compare the diff against the rules under `.agents/rules/`. Report only
deviations, one per line, with the rule path. Do not propose unrelated
improvements.
```

Save it as `.agents/agents/drift-reviewer.md`, then run `keystone index && keystone project`. On Claude Code, the main agent can then spawn `drift-reviewer` by name and hand it the diff to review.

## Where it lives

```
.agents/agents/<id>.md
```

The folder is `agents/`, the name every host with subagents uses for its own folder. The body is the system prompt. Write it as instructions to the subagent, not as notes about it.

## Frontmatter

| Key | Required | Meaning |
|---|---|---|
| `kind` | no | `subagent`. Keystone infers it from the `agents/` folder. |
| `id` | yes | The name the agent spawns it by. The projected name is the id in kebab case. |
| `description` | yes | One line. Keystone copies it into each host's file. |
| `tools` | yes | The tool allow-list, as the host's tool names (`Read`, `Grep`, `Bash`). |
| `model` | no | The model to run under, on hosts that accept one. |
| `tags` | no | Kebab-case labels for `keystone list --tag`. |

`keystone lint` reports an error when `tools` is missing or empty, so declare the list even when the subagent needs only `Read`.

## How each host gets it

| Host | What keystone writes |
|---|---|
| Claude Code | `.claude/agents/<id>.md` with `name`, `description`, `tools`, and `model`, then the body |
| Cursor | `.cursor/agents/<id>.md` with `name`, `description`, and `model` |
| GitHub Copilot | `.github/agents/<id>.agent.md` with `name` and `description` |
| opencode | `.opencode/agents/<id>.md` with `mode: subagent` and `model` |
| Codex | `.codex/agents/<id>.toml`, with the body as `developer_instructions` |
| Goose | `.goose/agents/<id>.md` with `name`, `description`, and `model` |
| Continue, Cline, Pi, Aider | A line in `AGENTS.md`. The agent opens the body from `.agents/agents/` by hand. |

Only the Claude Code file carries `tools`. On every other host, the projected file has no tool list, so the host decides which tools the subagent gets. Read [hosts.md](../hosts.md) for the full matrix, which also lists the host docs behind each file format.

## What a subagent does not do

- It doesn't produce evidence. Its report is a claim, so read the diff and run the checks yourself before you accept it.
- It doesn't run on an event by itself. To dispatch a review when an event fires, write an inferential [hook](hook.md) that names this subagent in its `agent:` field.
- It doesn't restrict tools on hosts that get no tool list. On those hosts, add a `deny` rule in a [permission](permission.md) where the host supports one.

## Failure modes

- **The host spawns the wrong subagent.** Two descriptions overlap, so give each subagent one job and name that job in its description.
- **The subagent does more than its job.** The prompt names no limits. State what it must not do.
- **The tool list doesn't hold.** Your host isn't Claude Code.
- **The report is long and vague.** The prompt names no output format. State the format, the length, and what the subagent must leave out of its report.

## Authoring

```
keystone new subagent <id>
```

This writes `.agents/agents/<id>.md` with `tools: [Read, Grep]` and a TODO description. Replace the description before you run `keystone project`, and widen the tool list only as far as the job needs.

Older keystone docs called the framework form of a subagent a persona. The kind and folder didn't change.
