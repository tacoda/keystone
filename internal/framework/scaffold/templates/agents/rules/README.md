# Rules

Each file here is a rule: a directive the agent follows. A rule is short and carries only the directives. Its reasoning lives in a reference, which the rule names under `corpus:`.

## When a rule applies

- **With `paths:`**, a rule applies when a touched file matches one of its globs. Hosts with a path-scoped rule format get a shim, such as `.claude/rules/` or `.cursor/rules/*.mdc`.
- **Without `paths:`**, a rule applies everywhere. Keystone lists it in the `AGENTS.md` managed block as one line, and the agent opens the body when it fits the phase.

`paths:` only narrows. It can't make a rule apply in more places. The topic folder (`process/`, `idioms/`, `domain/`) only groups files, and it doesn't change when a rule applies.

## Shape

```markdown
---
id: idioms/go/stdlib-first
description: Prefer Go stdlib over new third-party deps.
paths:
  - "cmd/**/*.go"
  - "internal/**/*.go"
corpus:
  - idioms/go/stdlib-first
tier: preference       # optional: preference | golden-rule | iron-law
---

# Stdlib first

## Iron law

- Only when violation causes real damage. Usually omit this section.

## Rules

- The default tier. Most directives land here.
```

Host shims copy only these sections: `Iron law`, `Golden rule`, `Rules`, and `Anti-patterns`.

**Tier discipline.** A golden rule means deviation needs a stated reason. An iron law is non-negotiable. Both get their force from being rare, so default new rules to `## Rules` and promote only when a person confirms.

## Topics

| Folder | What lives here | Ships |
|---|---|---|
| [`process/`](process/README.md) | The workflow phases and cross-cutting discipline. No `paths:`. | Populated |
| [`idioms/`](idioms/README.md) | Stack-specific rules, usually scoped by `paths:`. | Empty |
| [`domain/`](domain/README.md) | Business-rule constraints. | Empty |
| [`computational/`](computational/README.md) | Rules that run a command instead of being read. | Empty |

## Add one

Run `keystone new rule <topic>/<name>`. It scaffolds the rule and its paired reference. Then run `keystone index && keystone project`.
