---
kind: document
id: characterization
description: The empirical map of a region — what is true, before any guidance.
type: characterization
produced_by: plan
gates:
  - drafted
  - verified
  - landed
---

# characterization

How a system fails to guide agents is part of what the system *is* —
and the first thing a legacy kit fixes is that nothing says what is
true. Characterization runs before guidance: it records the region's
facts, so later work checks the map instead of re-deriving reality.

The state ledger's seeding step, as governed output: drafted (recorded),
verified (checked against the builds/tests and the region), landed (the
empirical map cites it, and CODEBASE_STATE references it).

## Current state

- **The build**: the entry point, and whether it runs from a clean clone.
- **The tests**: whether there are any, what they cover, and how long
  they take. When there are none, characterization tests stand where
  unit tests cannot yet: pin the behavior as-is, warts and all.
- **The regions**: top-level directories, what each holds, and which
  are hot (changes often) or cold.
- **The seams**: the polyglot boundaries — where stacks meet and the
  idioms change hands.
- **The risks**: the complexity, coupling, and coverage patterns the
  risk-fingerprint machinery reports per region.

## Done when

- The verified map cites this document, and every region an agent might
  touch has an as-written statement of what is true.
- Any claim that later proves wrong gets corrected in the same change
  that corrects the code — stale maps mislead worse than no map.
