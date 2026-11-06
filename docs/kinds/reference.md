# Reference

<!-- voice-register: informational -->
<!-- voice-english: ste -->

A reference holds the reasoning behind a rule: why it exists, what it guards against, and where it comes from. It's a constraint, and a keystone extension, because no host has a format for it. The agent opens a reference only when it needs the why, so a reference costs no context until then. The cost is that nothing loads it for you: a rule must link it, or the agent never finds it.

## Example

```markdown
---
id: idioms/go/stdlib-first
description: Why the Go code prefers the stdlib over new dependencies.
---

# Stdlib first, reasoning

Every direct dependency is code the team must audit and upgrade.
The Go stdlib already covers file I/O, paths, JSON, HTTP, and regexp.

## Anti-patterns

- Two libraries that do the same job.
- A dependency added for one helper function.

## Sources

- The Go standard library docs.

Back to the rule: [`rules/idioms/go/stdlib-first.md`](../../../rules/idioms/go/stdlib-first.md).
```

Save it as `.agents/references/idioms/go/stdlib-first.md`. Then link it from the rule:

```yaml
corpus:
  - idioms/go/stdlib-first
```

## Where it lives

```
.agents/references/<topic>/<name>.md
```

Keystone infers `kind: reference` from the `references/` folder, so you don't declare it. Use the same `<topic>/<name>` as the rule it explains, so the pair is easy to find.

## Frontmatter

| Key | Required | Meaning |
|---|---|---|
| `id` | yes | Stable id. A rule's `corpus:` entry names it |
| `description` | yes | One sentence on the reasoning this file holds |
| `tags` | no | Kebab-case labels for `keystone list --tag` |

A 4.x file with `kind: corpus` still parses. `keystone lint` warns, and `keystone lint --fix` rewrites it to `kind: reference`.

## Body

Write as much as the reasoning needs. Keep your rule short, because it loads often. A reference loads rarely, so it can carry the history, the trade-offs, and the sources. The scaffold gives three parts: the reasoning, `## Anti-patterns`, and `## Sources`, with a link back to the rule.

## How the agent gets it

Keystone doesn't project references to any host. The agent reaches one in two ways:

- **Through the rule.** Add the reference id to the rule's `corpus:` field. The agent follows that link when it needs the why.
- **Through MCP.** `keystone_get_reference id=<rule-id>` follows the rule's `corpus:` field and returns the linked reference bodies.

References also appear in `.keystone/INDEX.lite.json`, so the agent can browse them.

## What a reference does not do

- A reference doesn't load on its own. No host reads `.agents/references/`.
- A reference doesn't constrain anything by itself. The directive belongs in the rule.
- Keystone doesn't check that a rule and its reference agree. If they contradict each other, ask the user.

## Failure modes

- **The agent never reads it.** No rule names it in `corpus:`. Add the id to the rule.
- **A `corpus:` entry doesn't resolve.** The id is wrong or the reference is missing. `keystone lint` warns with the entry it can't resolve.
- **Directives drift into the reference.** The agent may not open it, so a directive there may never apply. Move it to the rule.

## Authoring

```
keystone new reference <topic>/<name>
```

This writes only the reference, for a rule that already exists or for reasoning that stands alone. `keystone new rule <topic>/<name>` writes a rule and its reference together. `keystone new corpus` exits with an error that names `keystone new reference`.
