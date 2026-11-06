---
kind: skill
id: keystone:learn
description: Capture a learning candidate (surprise, incident, review finding) into .charter/learning/inbox/ for later synthesis.
triggers:
  - keystone learn
  - keystone:learn
  - /keystone:learn
  - run the learn action
  - capture a learning candidate
  - log a surprise
model: opus
tools:
  - Read
  - Write
  - Edit
  - Grep
  - Glob
---

# keystone:learn — capture a learning candidate

The Learning flywheel's additive step. Write the smallest record of what
surprised you to `.charter/learning/inbox/<timestamp>-<slug>.md`
so a later `synthesize` can promote it to a guide, corpus entry, or
sensor.

Canonical playbook: `.charter/actions/learn.md`. Open it and
follow the activities — file shape, frontmatter, and proposed-layer
selection all live there.

## Run

Open `.charter/actions/learn.md` and execute every activity.

## When to trigger

- During implementation when something behaves unexpectedly.
- After an incident or review comment that revealed a missing rule.
- Mid-task when an agent ran into a gap the charter should have closed.

## Followups

- Review `.charter/learning/inbox/` at the dashboard's
  `/inbox` view. Mark candidates accepted / rejected.
- Run `/keystone:synthesize` to promote accepted candidates.
