---
kind: skill
id: keystone:synthesize
description: Promote accepted learning-inbox candidates into the right corpus / guide / sensor layer, then refresh INDEX.json.
triggers:
  - keystone synthesize
  - keystone:synthesize
  - /keystone:synthesize
  - run synthesize
  - promote learning candidates
  - promote inbox to guides
---

# keystone:synthesize — promote inbox candidates

The Learning flywheel's promotion step. Walks
`.keystone/harness/learning/inbox/`, promotes `status: accepted`
candidates to the right layer (guide / corpus / sensor), regenerates
the globs index, and refreshes Cursor projections if present.

Canonical playbook: `.keystone/harness/actions/synthesize.md`.

## Run

Open `.keystone/harness/actions/synthesize.md` and execute every
activity in order. Every promotion is a proposed diff — never overwrite
silently. After promotions land, run `keystone index` so
`.keystone/INDEX.json` reflects the new shape.

## When to trigger

- After accepting candidates from `/keystone:learn`.
- Periodically (weekly) as part of harness hygiene.

## Iron law

No silent overwrites. Propose every diff before applying.
