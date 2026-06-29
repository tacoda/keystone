---
name: keystone-learn
description: Capture a learning candidate (surprise, incident, review finding) into .keystone/harness/learning/inbox/ for later synthesis.
tools:
  - Read
  - Write
  - Edit
  - Grep
  - Glob
model: opus
---

# keystone:learn — capture a learning candidate

The Learning flywheel's additive step. Write the smallest record of what
surprised you to `.keystone/harness/learning/inbox/<timestamp>-<slug>.md`
so a later `synthesize` can promote it to a guide, corpus entry, or
sensor.

Canonical playbook: `.keystone/harness/actions/learn.md`. Open it and
follow the activities — file shape, frontmatter, and proposed-layer
selection all live there.

## Run

Open `.keystone/harness/actions/learn.md` and execute every activity.

## When to trigger

- During implementation when something behaves unexpectedly.
- After an incident or review comment that revealed a missing rule.
- Mid-task when an agent ran into a gap the harness should have closed.

## Followups

- Review `.keystone/harness/learning/inbox/` at the dashboard's
  `/inbox` view. Mark candidates accepted / rejected.
- Run `/keystone:synthesize` to promote accepted candidates.
