---
kind: rule
id: legacy/characterize-before-touch
description: 'No behavior change in a characterized region without a characterization test pinning it first — the legacy kit''s first rule.'
paths:
  - "**"
---
# Characterize before touch

Intro
Legacy

In a characterized region, no change of behavior runs without a
characterization test behind it — pinning the code as it is, warts and
all — before the behavior moves. In an uncharacterized region, the
characterization document runs first, and no re-derivation happens: what
is true is recorded, never re-derived from scratch by every agent that
passes through.

Exit condition: the behavior change lands behind a test that pins it,
and the pinning tests stay as-is while the code under them moves.
