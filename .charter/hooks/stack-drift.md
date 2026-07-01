---
kind: hook
id: stack-drift
description: 'Detects when the empirical reality of the repo has diverged from what CODEBASESTATE.'
tags:
  - computational
mode: computational
event: Stop
run: true
---
# Sensor: stack-drift

Detects when the empirical reality of the repo has diverged from what `CODEBASE_STATE.md` records. Re-running bootstrap is **not** the answer — bootstrap is a one-time initial scaffold. When this sensor flags drift, it's a trigger for **audit** to reconcile the relevant parts.

- **Trigger** — **audit** (full sweep) and on-demand when a change touches a region whose stack signals look off.
- **Inputs** — `.charter/corpus/state/CODEBASE_STATE.md` (declared stacks, regions, tool commands) and the actual repo (`package.json`, `go.mod`, `pyproject.toml`, `Gemfile`, `Cargo.toml`, `requirements.txt`, `build.gradle`, language file extensions per region).
- **Exit condition** — every declared stack and region has been compared against the repo and either matches or has a drift entry.
- **Output** — per-divergence row: declared value, observed value, recommended reconciliation.
- **State writes** — none directly. Findings feed the [charter-debt sensor](charter-debt.md) as `drifted-state` items; **audit**'s Pruning flywheel turns them into ledger entries.

## What counts as drift

| Signal | Drift if |
|---|---|
| **Declared stack present** | manifest file for the declared stack no longer exists, or the language has zero files in the region it was attached to. |
| **Undeclared stack present** | a new manifest exists (e.g., a `pyproject.toml` was added) and no stack of that kind is declared. |
| **Region map** | a top-level dir in `CODEBASE_STATE.md` doesn't exist on disk, or a top-level dir exists but isn't mapped. |
| **Tool command** | recorded command exits non-zero with "not found" (binary missing), or its config file (`.eslintrc`, `tsconfig.json`, `pyproject.toml`) is gone. |
| **CI platform** | recorded platform's config dir (`.github/workflows/`, `.gitlab-ci.yml`) is missing or empty, or a different CI's config exists. |

## What does **not** count as drift

- A stack adopting a new framework within itself (e.g., a TypeScript project switching from Jest to Vitest). That's a tool-command change, not stack drift; bootstrap doesn't need to re-run, the relevant tool entry just gets updated by **audit**.
- Adding a new directory under an existing region. Region map drift only fires for *top-level* changes.

## Why no rebootstrap action

[`bootstrap`](actions/bootstrap.md) is a one-time initial seed of state, idioms, and sensor classification. Re-running it would either overwrite human edits or no-op against existing content — neither is useful. Stack drift is an **incremental** problem the **audit** action is built for: reconcile only the parts that drifted, leave the rest alone.
