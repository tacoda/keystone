# Port: Corpus

**Activation:** On-demand — loaded when a guide forward-links to a corpus entry, or when an action explicitly references one.
**Purpose:** The reasoning, anti-patterns, history, and references behind a rule. The *why*. Rules live in the paired guide.

## Path convention

```
harness/corpus/<topic>/<name>.md                              # project-owned
harness/plugins/<plugin>/corpus/<topic>/<name>.md             # plugin-owned (read-only)
```

`<topic>` mirrors the topic structure under `guides/`. Each corpus file pairs (by topic + name) with a guide.

## Required shape

```markdown
# <Name> — reasoning

<long-form explanation of why the rule exists>

## Anti-patterns
<failure modes the rule guards against>

## References
<links to source material — papers, books, posts>

Back to the rules: [`guides/<topic>/<name>.md`](../../guides/<topic>/<name>.md).
```

- **H1 title** — required. Format: `# <Name> — reasoning` (convention).
- **Frontmatter** — none required.
- **Back-link to guide** — required when a paired guide file exists.
- **Length** — long-form welcomed. Corpus carries the depth a guide cannot.

## Cascade behavior

For a given `corpus/<topic>/<name>.md`:

1. The project's `harness/corpus/<topic>/<name>.md` always wins by default.
2. Otherwise, among plugins, outer plugins (shallower in `keystone.json`) win over plugins nested inside them — the first occurrence in a pre-order walk of `plugins[]`.
3. A `strict.corpus: [<name>]` declaration on any tree node makes that item absolute — nothing else can override it. `keystone verify` reports a violation if any layer attempts to shadow it.

Exactly one file loads per `<topic>/<name>`.

## Drift and guide pairing

The `drift` sensor flags corpus entries that no guide references, and guides whose paired corpus file is missing. `keystone doctor` reports both.

## Example

```markdown
# Surgical edits — reasoning

The cost of an adjacent change is rework risk for the reviewer: every line they
must read to confirm the change is correct, plus every line they must read to
confirm an unrelated line wasn't broken by accident. A focused diff bounds
that cost; a drive-by formatting pass does not.

## Anti-patterns
- Reformatting an entire file in a one-line fix.
- "While I'm here" refactors that double the diff size.
- Renaming variables in code you didn't otherwise touch.

## References
- *A Philosophy of Software Design*, John Ousterhout — chapter 5 on information leakage.

Back to the rules: [`guides/process/surgical-edits.md`](../../guides/process/surgical-edits.md).
```

## Authoring

```
keystone new guide <topic>/<name>
```

Scaffolds both the guide and its paired corpus stub. The corpus generator is also available standalone:

```
keystone new corpus <topic>/<name>
```
