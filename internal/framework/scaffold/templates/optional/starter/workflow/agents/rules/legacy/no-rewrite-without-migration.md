---
kind: rule
id: legacy/no-rewrite-without-migration
description: 'Large-scale refactorings get a migration document (gates: proposed, sequenced, in-flight, landed) — no big-bang rewrite.'
paths:
  - "**"
---
# No rewrite without a migration

No large-scale refactoring runs without a migration document behind it:
the target architecture, the ordered steps, and which regions still run
the old way. A rewrite with no document is not a refactoring — it is a
bet the whole team loses when the transition stalls.

The migration document tracks the transition across commits, promotes
forward and refuses backward, and holds the done-condition the agents
check before touching a region. The corpus query flow surfaces it.
