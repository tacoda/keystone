---
kind: guide
id: guides/idioms/charter-content/README
description: 'Charter content rules — entry point.'
globs:
  - ".charter/**/*.md"
  - "internal/framework/scaffold/templates/**/*.md"
---
# Charter content rules

Rules for authoring markdown primitives under `.charter/` and the install-time templates that ship them. Paired with [`../../../corpus/idioms/charter-content/`](../../../corpus/idioms/charter-content/) (reasoning).

## Activation

`globs:` above narrows ambient stack activation. These rules fire when a touched file matches `.charter/**/*.md` or `internal/framework/scaffold/templates/**/*.md`.

## Rule files

- [`primitive-shape.md`](primitive-shape.md) — frontmatter, paths, pair convention, narrow-only globs.
- [`state-files.md`](state-files.md) — authoring rules for `.charter/corpus/state/`.

(More files accumulate via the **learn** → **synthesize** flywheel.)
