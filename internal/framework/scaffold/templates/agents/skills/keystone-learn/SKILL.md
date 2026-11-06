---
name: keystone-learn
description: 'Capture a learning candidate (surprise, incident, review finding) into .keystone/learning/inbox/ for later synthesis. Triggers: keystone learn; keystone:learn; /keystone:learn; run the learn command; capture a learning candidate; log a surprise.'
---

# keystone:learn — capture a learning candidate

The Learning flywheel's additive step. Write the smallest record of what
surprised you to `.keystone/learning/inbox/<timestamp>-<slug>.md`
so a later `synthesize` can promote it to a rule, reference, or
hook.

Canonical command: `.agents/commands/learn.md`. Open it and
follow the activities — file shape, frontmatter, and proposed-layer
selection all live there.

## Run

Open `.agents/commands/learn.md` and execute every activity.

## When to trigger

- During implementation when something behaves unexpectedly.
- After an incident or review comment that revealed a missing rule.
- Mid-task when an agent ran into a gap the charter should have closed.

## Followups

- Review `.keystone/learning/inbox/` at the dashboard's
  `/inbox` view. Mark candidates accepted / rejected.
- Run `/keystone:synthesize` to promote accepted candidates.
