---
kind: skill
id: keystone:new-corpus
description: Scaffold a corpus reasoning entry (the *why*) at the canonical path.
triggers:
  - keystone new corpus
  - keystone:new-corpus
  - /keystone:new-corpus
  - add corpus reasoning
  - scaffold a corpus entry
---

# keystone:new-corpus — scaffold a corpus entry

A **corpus** entry is the long-form reasoning behind a rule. It lives
at `.keystone/harness/corpus/<topic>/<name>.md` and is opened only when
a guide's `traces:` (or a prose forward-link) points at it.

Most authors use `keystone:new-guide`, which scaffolds both the guide
and its paired corpus in one shot. Reach for `keystone:new-corpus`
only when adding reasoning to an existing guide that lacks one, or
when the corpus stands alone (rare).

## Run

```
keystone new corpus <topic>/<name>
```

Example:

```
keystone new corpus idioms/rails/migrations
# writes .keystone/harness/corpus/idioms/rails/migrations.md
```

## After scaffolding

1. Fill in the body — the failure modes the paired rule guards against,
   the trade-offs, the references that informed it.
2. Confirm the paired guide forward-links to this corpus entry (either
   via `traces:` in frontmatter or a prose `For reasoning, see ...`
   footer).
3. Run `keystone index` to refresh the descriptor surface.

Full port contract:
[`docs/ports/corpus.md`](../../../../docs/ports/corpus.md).
