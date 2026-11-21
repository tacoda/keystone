# Corpus

<!-- voice-register: informational -->
<!-- voice-english: ste -->

A corpus entry is **the what**: facts, general patterns, product
information. It is standalone information — no rule needs to cite it. The
triad: a [rule](rule.md) is **how** and specifics; a
[reference](reference.md) is **why**; a corpus entry is **what**.

Corpus is the new kind in 6.0. It is also where adopted repository
documents land in the index (6.0's `charter.roots` indexes docs that live
outside `.agents/` — GLOSSARY.md, design docs, plans — in place, without
moving or editing them).

## Example

```markdown
---
id: product/pricing-model
description: How subscriptions, seats, and overages are priced as of 6.0.
tags: [product, billing]
---

# Pricing model

Seats are billed monthly …
```

Save it as `.agents/corpus/product/pricing-model.md`, then run
`keystone index`. The entry appears in `.keystone/INDEX.json` and
`keystone list`; the agent opens the body when a fact about pricing is
relevant.

## Where it lives

```
.agents/corpus/<id>.md
```

A bare id is a single file; `<topic>/<name>` nests a folder. Both work.

## Frontmatter

| Key | Required | Meaning |
|---|---|---|
| `kind` | no | `corpus`. Keystone infers it from the `corpus/` folder. |
| `id` | yes | Stable identity in the index. Derived from the path when omitted. |
| `description` | yes | One line. The agent reads it in the index to decide whether to open the body. |
| `tags` | no | Kebab-case labels for `keystone list --tag`. |
| `paths` | no | Scopes the entry to file patterns, so `keystone charter query` surfaces it only for matching work. |

## What a corpus entry does not do

- It doesn't direct the agent. It informs. If the agent must do something,
  write a [rule](rule.md) and cite this entry from the rule's `corpus:`
  field or a [reference](reference.md).
- It doesn't load by itself. Corpus is on-demand, like every body.

## Failure modes

- **Stale facts.** Product information drifts. When the fact changes,
  update the entry in the same change that changes the behavior.
- **Everything becomes corpus.** If the agent must *do* something, that is
  a rule. If it explains *why*, that is a reference. Corpus is only the what.

## Authoring

```
keystone new corpus <id>
```

This writes `.agents/corpus/<id>.md` with a TODO description. Fill the
description in — it is the agent's only	signal for whether to open the body.
