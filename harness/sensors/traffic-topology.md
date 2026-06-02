# Sensor: traffic-topology

Combines git churn + recency + business criticality into a map of where attention concentrates.

- **Trigger** — **audit**, manually via **bootstrap**.
- **Inputs** — git log per region (last N months), business criticality flags from `corpus/state/CODEBASE_STATE.md` if the consumer has marked any.
- **Exit condition** — topology computed for all tracked regions.
- **Output** — heat map: per-region churn count, last-touched date, criticality tag.
- **State writes** — proposes a diff to `corpus/state/traffic-topology.md`. User accepts or edits.
