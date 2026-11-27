---
name: keystone-new
description: 'Scaffold a new primitive (rule, reference, corpus, hook, permission, eval, skill, command, agent, document) at the canonical path. Triggers: keystone new; keystone:new; /keystone:new; add a rule; add a hook; add a skill; scaffold a primitive.'
---

# keystone:new — scaffold a primitive

`keystone new` writes a skeleton at the conventional path with valid
frontmatter. Fill in the TODOs, then run `keystone index && keystone project`.

## Pick the kind

Ten kinds, no categories. The triad: **rule** = how, **reference** = why, **corpus** = what.

| Kind | Command | Writes |
|---|---|---|
| rule | `keystone new rule <topic>/<name>` | `.agents/rules/<topic>/<name>.md` + its reference |
| reference | `keystone new reference <topic>/<name>` | `.agents/references/<topic>/<name>.md` |
| corpus | `keystone new corpus <id>` | `.agents/corpus/<id>.md` |
| hook | `keystone new hook <id>` | `.agents/hooks/<id>.md` |
| permission | `keystone new permission <id>` | `.agents/permissions/<id>.md` |
| eval | `keystone new eval <id>` | `.agents/evals/<id>/` |
| skill | `keystone new skill <id>` | `.agents/skills/<id>/SKILL.md` |
| command | `keystone new command <id>` | `.agents/commands/<id>.md` |
| agent | `keystone new agent <id>` | `.agents/agents/<id>.md` |
| document | `keystone new document <id>` | `.agents/documents/<id>.md` |

The `tool` kind is retired in 6.0: keystone manages context, not callables.

## After scaffolding

1. Replace every `TODO` in the frontmatter and body.
2. A rule: add `paths:` to scope it to matching files, or leave it out to
   apply everywhere.
3. A hook: set `event:` and `run:`. A non-zero exit blocks the agent.
4. Run `keystone lint`, then `keystone index && keystone project`.

`keystone new` never overwrites: it refuses when the file exists.
