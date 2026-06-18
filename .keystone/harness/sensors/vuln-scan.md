---
kind: sensor
id: vuln-scan
description: 'Scans declared dependencies for known vulnerabilities (CVEs, advisories).'
---
# Sensor: vuln-scan

Scans declared dependencies for known vulnerabilities (CVEs, advisories).

- **Trigger** — **verify** (gate, when the lockfile changed) and **audit** (advisory, always).
- **Inputs** — the vuln-scan command from `corpus/state/CODEBASE_STATE.md` (e.g., `trivy fs`, `npm audit`, `pip-audit`, `bundler-audit`, `cargo audit`, `govulncheck`). Tool choice is the project's; the sensor describes the contract.
- **Exit condition** — exit code 0 and no findings at or above the project's severity threshold (recorded in `CODEBASE_STATE.md`; defaults to `high`).
- **Output** — pass/fail. On fail: dependency name, version, advisory ID, fixed version.
- **State writes** — none.

## Severity threshold

Projects set their own threshold in `CODEBASE_STATE.md`. Default: fail on `high` or `critical`. Production-facing services usually drop to `medium`; tooling and CLIs sometimes raise to `critical`.

## When a vuln has no fixed version

If the advisory has no published fix, the failure cannot be resolved by an upgrade. Two paths:

1. **Mitigate** — confirm the vulnerable code path is unreachable in this project; document in `corpus/state/debt.md` as a `deliberate` debt item with the advisory ID as the revisit trigger.
2. **Replace** — swap the dependency.

Either way, suppress the specific advisory in the tool's config (don't broaden the severity threshold) and record why.

## Scope

This sensor scans declared dependencies. It does not scan vendored or transitively-copied code — those are caught by **sast** on the source-tree pass.
