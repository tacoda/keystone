---
kind: sensor
id: sast
description: 'Static application security testing — pattern-based detection of insecure code in the source tree (SQL injection, command injection, weak...'
host_triggers:
  - phase: Stop
    command: true
    timeout: 30
tags:
  - computational
---
# Sensor: sast

Static application security testing — pattern-based detection of insecure code in the source tree (SQL injection, command injection, weak crypto, unsafe deserialization, etc.).

- **Trigger** — **verify** (diff-scoped, gate) and **audit** (codebase-scoped, advisory).
- **Inputs** — the SAST command from `corpus/state/CODEBASE_STATE.md` (e.g., `semgrep ci`, `bandit`, `brakeman`, `gosec`). Tool choice is the project's; the sensor describes the contract.
- **Exit condition** — exit code 0 and no findings at or above the project's severity threshold (recorded in `CODEBASE_STATE.md`; defaults to `error`).
- **Output** — pass/fail. On fail: file, line, rule ID, message, severity.
- **State writes** — none.

## Diff-scoped vs. codebase-scoped

- **Verify** runs SAST on the diff only. This is fast and catches new insecurity introduced by the current change.
- **Audit** runs SAST across the entire source tree. Slower, but catches insecurity that pre-existed the harness or that the **drift** sensor doesn't catch because the diff didn't touch it.

If your tool has a "diff mode" (e.g., `semgrep ci --baseline-ref`), use it for verify. Otherwise scope manually by passing only the changed paths.

## Relationship to review-security

[`review-security`](review-security.md) is the **inferential** sibling of this sensor: an agent reasoning over the diff. SAST is pattern-based, fast, and catches the things tools are good at. The reviewer catches the things tools miss — design flaws, authorization gaps, business-logic abuse.

Run both. Treat their findings as complements, not duplicates.
