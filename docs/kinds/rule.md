# Rule

<!-- voice-register: informational -->
<!-- voice-english: ste -->

A rule is a directive the agent follows. It's a constraint, because it limits what the agent does. Use `paths:` to scope a rule to the files it governs; a rule without `paths:` applies everywhere. A rule can also be computational, which means it runs a command on an event instead of being read. The cost: an inferential rule is only text, so the agent can ignore it. Pair it with a hook when a violation must block.

## Example

```markdown
---
id: idioms/go/stdlib-first
description: Prefer Go stdlib over new third-party deps.
paths:
  - "cmd/**/*.go"
  - "internal/**/*.go"
  - "go.mod"
corpus:
  - idioms/go/stdlib-first
---

# Stdlib first

## Iron law

- No new direct dep without naming the stdlib package it replaces.

## Rules

- Serialization uses `encoding/json` and `gopkg.in/yaml.v3`. One YAML parser.

## Anti-patterns

- Two libraries that do the same job.
```

Save it as `.agents/rules/idioms/go/stdlib-first.md`, then run `keystone index && keystone project`.

## Where it lives

```
.agents/rules/<topic>/<name>.md
```

Keystone infers `kind: rule` from the `rules/` folder, so you don't need to declare it. The topic folder (`process`, `idioms`, `domain`, …) only groups your files, and it doesn't change when a rule applies.

## Frontmatter

| Key | Required | Meaning |
|---|---|---|
| `id` | yes | Stable id. Lint fails without it |
| `description` | yes | One sentence. Pathless rules show it in `AGENTS.md` |
| `paths` | no | Glob list. `**` matches any number of folders. Omit it to apply everywhere |
| `corpus` | no | Reference ids that hold the reasoning. Lint warns when one doesn't resolve |
| `tier` | no | `iron-law`, `golden-rule`, or `preference` (the default). Lint checks the value |
| `mode` | no | `computational` makes the rule a check. See below |
| `event` | computational only | The event that fires the check |
| `run` | computational only | The shell command. Lint fails when a computational rule has none |
| `tags` | no | Kebab-case labels for `keystone list --tag` |

The 4.x spellings still parse: `kind: guide` and `globs:`. `keystone lint` warns once per file, and `keystone lint --fix` rewrites them for you.

## Body

Write the rule as short directives, and keep the reasoning in a [reference](reference.md). Host shims copy only these `##` sections, in any case: `Iron law`, `Iron laws`, `Golden rule`, `Golden rules`, `Rules`, and `Anti-patterns`. Other sections stay put. The shim points to that file, so the agent can read it on demand.

## How each host gets it

- **A rule without `paths:`** goes into the `AGENTS.md` managed block as one line, with its path and description. The agent opens the body when it applies.
- **A rule with `paths:`** gets a shim on each host with a path-scoped rule format. Those are Claude Code (`.claude/rules/`), Cursor (`.cursor/rules/*.mdc`), Copilot (`.github/instructions/`), Continue (`.continue/rules/`), and Cline (`.clinerules/`).
- **On other hosts**, the agent finds the rule in `.keystone/INDEX.lite.json` and opens it itself.

See [`docs/hosts.md`](../hosts.md) for the full table.

## Computational rules

A computational rule records a deterministic tool that your stack already runs, such as a formatter or a linter. It replaces the 4.x computational guide.

```yaml
---
id: computational/gofmt
description: gofmt, canonical Go formatting.
mode: computational
event: PostToolUse
run: 'test -z "$(gofmt -l cmd internal)"'
---
```

`keystone hook fire` selects any primitive that has both `event:` and `run:`. So a computational rule fires through the same path as a [hook](hook.md), and a non-zero exit blocks. It gets no host shim and no `AGENTS.md` line. Its `paths:` are documentation only, because the hook layer doesn't filter on them. And without `event:`, the rule never runs.

## What a rule does not do

- An inferential rule doesn't enforce anything. The agent reads it and can still ignore it.
- `paths:` only narrows where a rule applies. It can't make a rule apply in more places.
- `tier:` doesn't change loading or projection. Only lint reads it today.
- Keystone doesn't merge two rules with the same id. Instead, lint reports the duplicate as an error.

## Failure modes

- **The rule never applies.** Its `paths:` match no file the agent touches. Check each pattern against real files, or write an [eval](eval.md) that asserts the rule fires.
- **The shim has no directives.** The body has none of the copied sections, so the shim only points to the source. Rename your sections to match the list above.
- **A computational rule never fires.** It has no `event:`, or the host has no hook bridge. Read the `keystone project` report, which names the hosts that don't run hooks.

## Authoring

```
keystone new rule <topic>/<name>
```

This writes `.agents/rules/<topic>/<name>.md` and a paired reference at `.agents/references/<topic>/<name>.md`. The rule's `corpus:` entry already points to that reference. `keystone new guide` exits with an error that names `keystone new rule`.
