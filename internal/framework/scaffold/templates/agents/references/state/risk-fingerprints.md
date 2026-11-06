---
kind: corpus
id: corpus/state/risk-fingerprints
description: 'Per-region risk score based on complexity + coupling + coverage.'
---
# Risk Fingerprints

> **Template.** The **bootstrap** or **audit** action will populate this from code metrics + coverage. Until then, leave as-is or fill in by hand.

Per-region risk score based on complexity + coupling + coverage. Identifies the parts of the codebase most likely to harbor defects under change.

## Table

| Region | Complexity (cyclomatic) | Coupling (fan-in/fan-out) | Coverage | Risk score | Notes |
|---|---|---|---|---|---|
| `<region-path>` | `<avg or max>` | `<in/out>` | `<percentage>` | `<low|medium|high|critical>` | `<optional notes>` |

## How risk score is computed

Heuristic combination:

- **complexity** — average or max cyclomatic complexity in the region.
- **coupling** — number of modules that import from this region (fan-in) and number this region imports from (fan-out).
- **coverage** — test coverage percentage.

Score:

- **critical** — high complexity + high coupling + low coverage.
- **high** — any two of the above.
- **medium** — any one of the above.
- **low** — none of the above flagged.

## How to use it

- **Critical regions** → prioritize for refactor or coverage backfill before further feature work.
- **High regions** → review every change carefully; the **review** action should pay extra attention.
- **Medium / low regions** → standard review.

## Pairing with traffic topology

Cross-reference with `traffic-topology.md`:

- Critical risk + high churn → fragile and active. Highest danger.
- Critical risk + low churn → fragile but stable. Don't touch unless you must.
- Low risk + high churn → healthy iteration. Keep going.
