---
name: keystone-synthesize
description: 'Promote accepted learning-inbox candidates into the right reference / rule / hook layer, then refresh INDEX.json. Triggers: keystone synthesize; keystone:synthesize; /keystone:synthesize; run synthesize; promote learning candidates; promote inbox to rules.'
---

# keystone:synthesize — promote inbox candidates

The Learning flywheel's promotion step. Walks
`.keystone/learning/inbox/`, promotes `status: accepted`
candidates to the right layer (rule / reference / hook), regenerates
the globs index, and refreshes Cursor projections if present.

Canonical command: `.agents/commands/synthesize.md`.

## Run

Open `.agents/commands/synthesize.md` and execute every
activity in order. Every promotion is a proposed diff — never overwrite
silently. After promotions land, run `keystone index` so
`.keystone/INDEX.json` reflects the new shape.

## When to trigger

- After accepting candidates from `/keystone:learn`.
- Periodically (weekly) as part of charter hygiene.

## Iron law

No silent overwrites. Propose every diff before applying.
