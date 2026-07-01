---
kind: sensor
id: risk-fingerprint
description: 'Computes complexity + coupling + coverage patterns per region.'
tags:
  - computational
mode: computational
on: Stop
run: true
---
# Sensor: risk-fingerprint

Computes complexity + coupling + coverage patterns per region.

- **Trigger** — **audit**, manually via **bootstrap**.
- **Inputs** — code metrics (cyclomatic complexity, fan-in/fan-out, churn from git), coverage data.
- **Exit condition** — fingerprints computed for all tracked regions.
- **Output** — risk score per region with the metric breakdown.
- **State writes** — proposes a diff to `corpus/state/risk-fingerprints.md`. User accepts or edits.
