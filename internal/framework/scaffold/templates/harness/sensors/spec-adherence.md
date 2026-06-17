---
kind: sensor
id: spec-adherence
description: 'Walks the spec''s acceptance criteria against the current diff.'
---
# Sensor: spec-adherence

Walks the spec's acceptance criteria against the current diff.

- **Trigger** — **review** (review phase).
- **Inputs** — the spec (`docs/specs/<file>.md`) and the diff.
- **Exit condition** — every criterion is met *with evidence* (a test, an output, a manual check).
- **Output** — per-criterion pass/fail with evidence link. A criterion missing evidence fails.
- **State writes** — none.
