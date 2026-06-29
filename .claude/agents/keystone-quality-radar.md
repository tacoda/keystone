---
name: keystone-quality-radar
description: Aggregates verification and review sensors into a five-dimension scorecard for the current diff (or the whole codebase, on audit).
tools:
  - Read
  - Grep
---
# Sensor: quality-radar

Aggregates verification and review sensors into a five-dimension scorecard for the current diff (or the whole codebase, on **audit**).

- **Trigger** — **audit** (codebase-wide) and **review** (diff-scoped).
- **Inputs** — outputs of [lint](lint.md), [type-check](type-check.md), [test](test.md), [coverage](coverage.md), [review-security](review-security.md), [review-functional](review-functional.md), plus complexity metrics from [risk-fingerprint](risk-fingerprint.md).
- **Exit condition** — every dimension has a score and at least one cited finding (or an explicit "no findings").
- **Output** — one score per dimension (`green | yellow | red`) with the citations that drove the score.
- **State writes** — proposes a diff to `corpus/state/quality-radar.md`. User accepts or edits.

## Dimensions

| Dimension | Driven by | Green | Yellow | Red |
|---|---|---|---|---|
| **Type safety** | `type-check` exit + signature churn | no errors | errors confined to one region | errors across regions or `any`/`unknown` proliferation |
| **Test quality** | `test` exit + `coverage` + test-to-code ratio in diff | passes; coverage stable or up; new code tested | passes; coverage flat with new code | failures, coverage drop, or new code untested |
| **Readability** | `lint` + complexity from `risk-fingerprint` | lint clean; no new critical-complexity regions | lint warns; one new high-complexity region | lint errors or multiple new critical regions |
| **Security** | `review-security` findings | none | low-severity only | medium or higher |
| **Performance** | hot-path heuristics from `review-functional` + diff size on perf-critical regions (flagged in `CODEBASE_STATE.md`) | no perf-critical touch or no concerns | perf-critical touch, no benchmarks | perf-critical touch with regression signal |

## Not a gate

This is a **scorecard**, not a verification gate. Red on a dimension is a signal to discuss, not an automatic block — the **review** action decides what to do with it. The actual block is whatever the underlying sensor already imposes (a failed `test` blocks regardless of how the radar shades the diff).
