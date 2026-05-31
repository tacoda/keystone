---
last_reconciled: <YYYY-MM-DD>
---

# Traffic Topology

> **Template.** The **bootstrap** or **audit** action will populate this from `git log` and your codebase. Until then, leave it as-is or fill in by hand.

Where attention concentrates in the codebase. Combines git churn + recency + business criticality.

## Heat map

| Region | Churn (commits, last 90d) | Last touched | Criticality | Notes |
|---|---|---|---|---|
| `<region-path>` | `<count>` | `<YYYY-MM-DD>` | `<low|medium|high|critical>` | `<optional notes>` |

## How to read it

- **Critical + high churn** → fragile hot zone. Refactor priority.
- **Critical + low churn** → stable core. Treat as risky to change.
- **Low criticality + high churn** → experimentation area. Light review.
- **Low criticality + low churn** → idle / forgotten. Audit for dead code.

## Criticality tags

- **critical** — the project is broken or unsafe if this region is broken.
- **high** — user-facing or revenue-impacting.
- **medium** — supports critical paths.
- **low** — tooling, scripts, internal utilities.
