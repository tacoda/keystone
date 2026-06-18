---
kind: guide
id: guides/idioms/harness-content/README
description: 'Harness content rules — entry point.'
globs:
  - ".keystone/harness/**/*.md"
  - "internal/framework/templates/**/*.md"
---
# Harness content rules

Rules for authoring markdown primitives under `.keystone/harness/` and the install-time templates that ship them. Paired with [`../../../corpus/idioms/harness-content/`](../../../corpus/idioms/harness-content/) (reasoning).

## Activation

`globs:` above narrows ambient stack activation. These rules fire when a touched file matches `.keystone/harness/**/*.md` or `internal/framework/templates/**/*.md`.

## Rule files

- [`primitive-shape.md`](primitive-shape.md) — frontmatter, paths, pair convention, narrow-only globs.
- [`state-files.md`](state-files.md) — authoring rules for `harness/corpus/state/`.

(More files accumulate via the **learn** → **synthesize** flywheel.)
