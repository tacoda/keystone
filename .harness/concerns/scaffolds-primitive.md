---
kind: concern
id: scaffolds-primitive
description: Reusable concern for keystone-new-* skills — declares the tools needed to scaffold a new primitive at its canonical path and the idempotency invariants every scaffold respects.
tools:
  - Read
  - Write
  - Edit
  - Glob
  - Bash
tags:
  - scaffold
  - composition
---

# Concern: scaffolds-primitive

Composed into every `keystone-new-<kind>` skill. Establishes the
shared protocol every scaffold follows so a user typing
`keystone:new-action` and `keystone:new-sensor` lands files with the
same shape, the same frontmatter discipline, and the same recovery
behavior on re-run.

## Tools

`Read`, `Write`, `Edit`, `Glob`, `Bash` — the minimum surface to
create a new file at the canonical path, refuse to clobber an
existing primitive, run `keystone index` after the write, and report
the resulting file paths.

## Invariants

- **Canonical path.** Every primitive kind has exactly one place it
  lives (see `corpus/idioms/harness-content/primitive-shape.md`). The
  scaffold refuses to write off-path.
- **No clobber.** If a file already exists at the target path, the
  scaffold reports it and exits — the user must `keystone show` /
  edit / delete first. Idempotent re-runs are safe.
- **Frontmatter completeness.** Every scaffolded file lands with
  `kind`, `id`, and `description` populated. Placeholder
  `description: TODO` is a lint warning by design — surfaces
  unfinished work in `keystone lint` and the debt sensor.
- **Index after write.** Every scaffold ends by running
  `keystone index` so `INDEX.json` and `INDEX.lite.json` reflect the
  new primitive immediately — no manual step.
- **Paired corpus for guides.** When the new primitive is a guide,
  the scaffold also lays down the paired `corpus/<topic>/<slug>.md`
  stub. The harness-debt sensor reports orphan guides and orphan
  corpus; the scaffold avoids creating them.

## What this concern does NOT do

- Does not author the body content beyond frontmatter — that's the
  user's job in the editor.
- Does not run `keystone project` — the projection pipeline is a
  separate user action (or `keystone watch`).
- Does not register the primitive in `keystone.json` — primitives
  live in `.keystone/harness/`, not in the project config.
