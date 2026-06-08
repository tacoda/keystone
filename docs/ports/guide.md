# Port: Guide

**Activation:** Ambient — every guide in the resolved cascade is loaded into the agent's context at session start.
**Purpose:** The rules the agent must follow. The *what* and *what not*. Reasoning lives in the paired corpus file.

## Path convention

```
harness/guides/<topic>/<name>.md                              # project-owned
harness/plugins/<plugin>/guides/<topic>/<name>.md             # plugin-owned (read-only)
```

`<topic>` groups related guides (`process`, `principles`, `idioms`, `domain`, `computational`). Topic directories are open-ended — adding a new topic is just creating a new directory.

## Required shape

```markdown
# <Name> — rules

<one-sentence framing of what this guide governs>

## IRON LAW(S)

<non-negotiable rules; omit this section when nothing here qualifies>

## GOLDEN RULES

<strong, explicit standards; omit when nothing here qualifies>

## RULES

<regular rules — the default tier; most directives live here>

For reasoning, see [`corpus/<topic>/<name>.md`](corpus/<topic>/<name>.md).
```

- **H1 title** — required. Format: `# <Name> — rules` (the `— rules` suffix is convention, not enforced).
- **Frontmatter** — none required. Reserved keys (if added later) live in the frontmatter; bare-content guides remain valid.
- **Forward-link to corpus** — required when a paired corpus file exists. Harness-root-relative path (no `../` segments; `keystone doctor` enforces).
- **Length** — short. A guide is rules; long-form belongs in corpus. Rough ceiling: one screen.

### Rules tiers

Rules in a guide are organized by strength. Only `## RULES` is mandatory — omit the special tiers when no rule warrants them (the common case).

| Tier              | When to use it                                                                                                         |
| ----------------- | ---------------------------------------------------------------------------------------------------------------------- |
| `## IRON LAW(S)`  | Non-negotiable. Violation causes real damage (data loss, security breach, contract break). Rare by design.             |
| `## GOLDEN RULES` | Stronger than regular rules; **deviation requires reasoning**. May be concrete prescriptions or aspirational ideals.   |
| `## RULES`        | Regular rules. The default tier — most directives live here.                                                           |

The tier framing is part of the agent's reading discipline: iron laws short-circuit any conflicting instruction; golden rules can be overridden only with explicit reasoning; regular rules can be overridden when a stronger rule applies.

## Cascade behavior

For a given `guides/<topic>/<name>.md`:

1. The project's `harness/guides/<topic>/<name>.md` always wins by default.
2. Otherwise, among plugins, outer plugins (shallower in `keystone.json`) win over plugins nested inside them — the first occurrence in a pre-order walk of `plugins[]`.
3. A `strict.guides: [<name>]` declaration on any tree node makes that item absolute — nothing else can override it, not the project, not any other plugin. `keystone verify` reports a violation if any layer attempts to shadow a strict item.

The framework never composes overlapping guides for the same name. Exactly one file loads.

## Drift and corpus pairing

Every guide *should* have a paired corpus file at `corpus/<topic>/<name>.md`. The `drift` sensor flags rules in guides that have no corresponding corpus reasoning, and corpus entries that no guide references.

A guide without a corpus pair is permitted (some rules are self-evident), but `keystone doctor` reports them so authors can decide.

## Example

```markdown
# Surgical edits — rules

Touch only what the task requires. Don't refactor adjacent code, don't reformat
files you didn't change, don't "improve" comments you didn't introduce.

## RULES

- Every changed line traces directly to the user's request.
- Orphaned imports/variables your change created: remove.
- Pre-existing dead code: mention, don't delete (unless asked).

For reasoning, see [`corpus/process/surgical-edits.md`](corpus/process/surgical-edits.md).
```

## Authoring

```
keystone new guide <topic>/<name>
```

Scaffolds the guide and a paired corpus stub at the conventional paths.
