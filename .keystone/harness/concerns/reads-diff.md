---
kind: concern
id: reads-diff
description: Reusable concern for review personas — declares the tools needed to read a git diff and the section that documents the protocol.
tools:
  - Read
  - Grep
  - Bash
tags:
  - review
  - composition
---

# Concern: reads-diff

Composed into review personas that inspect the working diff before
emitting findings. Provides the minimum tool surface (`Read`, `Grep`,
`Bash`) and codifies the read protocol so every reviewer behaves the
same way.

## Read protocol

The diff is the source of truth — not the agent's prior summary,
not the tracker card, not the PR description. Read it directly:

```
git diff --stat                       # scope check first
git diff HEAD                         # full unstaged diff
git diff --staged                     # what's about to commit
git diff <base>...HEAD                # PR-style range diff
```

For large diffs (>500 lines) read the stat first, then per-file diffs
on demand. Never assume what the diff says — every claim a reviewer
makes must trace to a line the reviewer actually read.

## What this concern does NOT do

- Does not declare `severity:` — that's the host primitive's choice.
- Does not provide LLM-judgment heuristics — each reviewer persona
  owns its own review lens.
- Does not run a sensor — `keystone verify --sensor drift` is the
  computational equivalent and lives elsewhere.
