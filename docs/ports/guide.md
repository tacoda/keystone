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

<rules: bullets, numbered lists, or short prose paragraphs>

For reasoning, see [`corpus/<topic>/<name>.md`](../../corpus/<topic>/<name>.md).
```

- **H1 title** — required. Format: `# <Name> — rules` (the `— rules` suffix is convention, not enforced).
- **Frontmatter** — none required. Reserved keys (if added later) live in the frontmatter; bare-content guides remain valid.
- **Forward-link to corpus** — required when a paired corpus file exists. Relative path, kept resolvable.
- **Length** — short. A guide is rules; long-form belongs in corpus. Rough ceiling: one screen.

## Cascade behavior

For a given `guides/<topic>/<name>.md`:

1. The project's `harness/guides/<topic>/<name>.md` always wins.
2. Otherwise, the first occurrence in a pre-order walk of `keystone.json`'s `plugins[]` tree wins.
3. A `strict.guides: [<name>]` declaration on a tree node blocks any deeper node from shipping `guides/<topic>/<name>.md`. Violations fail `keystone verify` at install time.

The framework never composes overlapping guides for the same name. Exactly one file loads.

## Drift and corpus pairing

Every guide *should* have a paired corpus file at `corpus/<topic>/<name>.md`. The `drift` sensor flags rules in guides that have no corresponding corpus reasoning, and corpus entries that no guide references.

A guide without a corpus pair is permitted (some rules are self-evident), but `keystone doctor` reports them so authors can decide.

## Example

```markdown
# Surgical edits — rules

Touch only what the task requires. Don't refactor adjacent code, don't reformat
files you didn't change, don't "improve" comments you didn't introduce.

- Every changed line traces directly to the user's request.
- Orphaned imports/variables your change created: remove.
- Pre-existing dead code: mention, don't delete (unless asked).

For reasoning, see [`corpus/process/surgical-edits.md`](../../corpus/process/surgical-edits.md).
```

## Authoring

```
keystone new guide <topic>/<name>
```

Scaffolds the guide and a paired corpus stub at the conventional paths.
