# Skill

<!-- voice-register: informational -->
<!-- voice-english: ste -->

A skill is a workflow the agent loads when the task matches its description, so it adds something the agent can do. Each skill is a folder with a `SKILL.md` file written to the [Agent Skills spec](https://agentskills.io/specification), which means most hosts read it with no translation. The cost is control: the host decides when a skill fires, and keystone can't force it to activate.

## Example

```markdown
---
name: release-notes
description: Draft release notes from the merged PRs since the last tag. Use when the user asks for release notes or a changelog entry.
allowed-tools: Read Bash Grep
metadata:
  model: sonnet
---

# release-notes

1. Find the last tag with `git describe --tags --abbrev=0`.
2. List the merged PRs since that tag.
3. Group them by type and write one line per change.
```

Save it as `.agents/skills/release-notes/SKILL.md`, then run `keystone index && keystone project`.

## Where it lives

```
.agents/skills/<name>/SKILL.md
```

The folder name is the skill's name. You can also put scripts or templates in the folder next to `SKILL.md`. Hosts that mirror skills take the whole folder, so your files travel with it. The file declares no `kind`. Keystone infers `kind: skill` from the `skills/` folder.

## Frontmatter

These are the six keys the spec allows, and keystone reads each one into the same descriptor that other kinds use.

| Key | Required | Meaning |
|---|---|---|
| `name` | yes | Lowercase letters, digits, and single hyphens, max 64 characters. Must match the folder name. Keystone uses it as the id. |
| `description` | yes | What the skill does and when to use it, max 1024 characters. The host matches the task against this text, so write it for the task. |
| `license` | no | The license of the skill. |
| `compatibility` | no | Environment needs, max 500 characters. |
| `metadata` | no | A string map. Keystone reads `model` and `tags` from it. |
| `allowed-tools` | no | A space-separated tool list. Keystone reads it as the tool allow-list. |

`keystone lint` checks these limits for you. It reports an error for a bad `name`, a name that doesn't match its folder, or an over-long field. A missing `name` is only a warning.

### The 4.x keys

Keystone still reads the 4.x keys `kind`, `id`, `triggers`, `tools`, `model`, and `tags`. But lint warns about them, because strict hosts reject keys outside the spec. To move them, run `keystone lint --fix`. That fix does four things:

- It takes `name` from the folder name.
- It folds `triggers` into the end of `description`, or into `metadata.triggers` when the result would pass 1024 characters.
- It turns `tools` into `allowed-tools`.
- It moves `model` and `tags` into `metadata`.

If your skill carries a key the fix can't map, the fix refuses it. Move that key by hand.

A 4.x playbook or pattern is now a skill. The 5.0 migration moves each one to `.agents/skills/<id>/SKILL.md`, and [the upgrade guide](../upgrade-4.x-to-5.0.md) has the details.

## How each host gets it

`keystone project` writes these, using the folder name as the host name with no `keystone-` prefix added.

| Host | What keystone does |
|---|---|
| Claude Code | Links the folder to `.claude/skills/<name>/`. It copies when `link_mode` is `copy`, and on Windows by default. |
| Continue | Copies the folder to `.continue/skills/<name>/`. |
| Cline | Copies the folder to `.cline/skills/<name>/`. |
| Cursor, GitHub Copilot, opencode, Codex, Goose, Pi | Nothing. These hosts read `.agents/skills/` by themselves. |
| Aider | Lists the skill in `AGENTS.md`. The agent opens the body by hand. |

For the other kinds on each host, and for the sources behind each row, read the full matrix in [hosts.md](../hosts.md).

## What a skill does not do

- It doesn't run by itself: the host loads the body into the session, and then the agent follows the steps it reads.
- It doesn't guarantee activation, because a description that doesn't match how people ask never loads the skill.
- It doesn't gate anything. To block an action before it runs, write a [hook](hook.md) with a `run:` check.
- It isn't a slash command, so if the user must start the work by typing its name, write a [command](command.md).

## Failure modes

- **The skill never fires.** The description is too vague. Name the task and the words your users type.
- **A strict host rejects the file.** The frontmatter carries keys outside the spec, which `keystone lint --fix` moves.
- **The agent sees the skill twice.** Two folders the host reads hold the same skill, and `keystone project` warns about that duplicate.
- **A flat `skills/<id>.md` file doesn't project.** Hosts that mirror folders need `<name>/SKILL.md`, so give the skill its own folder.

## Authoring

```
keystone new skill <name>
```

This writes `.agents/skills/<name>/SKILL.md` with `name` set and a TODO description. Fill in the description before you run `keystone project`, because lint warns while it still says TODO.
