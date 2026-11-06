# Skills

Each folder here is a skill: a workflow the agent loads when the task matches its description. A skill is a `SKILL.md` file written to the Agent Skills spec, so most hosts read it with no translation. The host decides when a skill fires.

## Shape

```markdown
---
name: release-notes
description: Draft release notes from merged PRs since the last tag. Use when the user asks for release notes.
allowed-tools: Read Bash Grep   # optional
---

# release-notes

The steps the agent follows.
```

- `name` must match the folder name. Keystone uses it as the id.
- `description` says what the skill does and when to use it. The host matches the task against this text.
- Scripts or templates can sit next to `SKILL.md`. They travel with the folder.

The shipped [`task`](task/SKILL.md) skill runs the lifecycle end to end.

## Add one

Run `keystone new skill <name>`, fill in the body, then run `keystone index && keystone project`. `keystone lint` checks the name and field lengths, and `keystone lint --fix` moves 4.x keys such as `triggers:` into spec form.
