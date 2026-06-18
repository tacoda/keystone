---
kind: guide
id: guides/idioms/harness-content/primitive-shape
description: 'Frontmatter, canonical paths, corpus/guide pairing, narrow-only globs.'
globs:
  - ".keystone/harness/**/*.md"
  - "internal/framework/templates/**/*.md"
---
# Primitive shape — rules

The rules from [`corpus/idioms/harness-content/primitive-shape.md`](../../../corpus/idioms/harness-content/primitive-shape.md).

## IRON LAW

- **Every primitive declares `kind`, `id`, and `description` in frontmatter.** No exceptions. `keystone index` ignores files that don't.
- **Every primitive lives at the canonical path for its kind** (table in the corpus). Off-path content is invisible.
- **`globs:` narrows, never broadens.** A guide with globs activates only when its ambient rule already said yes AND a touched file matches. Globs cannot force a guide to fire outside its topic.
- **No hand-edits to generated files.** `INDEX.json` and the `## Index` table in `GLOBS_INDEX.md` are regenerated; manual edits are overwritten.

## GOLDEN PATH

- New guide → write the paired corpus entry in the same change. Either side alone is a smell the **harness-debt** sensor flags.
- After any primitive add / move / delete → run `keystone index` so `INDEX.json` reflects the new shape. If the primitive is a skill / subagent / command, also run `keystone project` so `.claude/` regenerates.
- A guide's `description:` reads like a one-line answer to "should I open this file right now?" — the agent uses it as a gate.
- Use the same `id:` value as the path slug (e.g., `id: guides/idioms/go/stdlib-first` for `guides/idioms/go/stdlib-first.md`). Aliases hide intent.
- `severity: must` only when deviation actually warrants a sensor-grade complaint. Default to `should` and let the corpus carry the reasoning.

## Anti-patterns

- Placeholder `description:` (`'TODO'`, `'description'`, copy-paste from another primitive).
- A guide with `globs:` patterns that match no real files (typically copy-pasted from a different stack).
- A corpus file with no paired guide, or a guide whose body has no corpus link.
- A new primitive that lands without a `keystone index` run — `.keystone/INDEX.json` silently drifts until the next agent session that reads it.
