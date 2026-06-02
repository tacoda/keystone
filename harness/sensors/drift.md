# Sensor: drift

Compares the diff (or full codebase, during audit) against loaded guide rules. Finds violations of IRON LAWs and GOLDEN RULES.

- **Trigger** — **check-drift** (implementation), **verify** (verification), **audit** (discipline).
- **Inputs** — current diff (or file set for audit), loaded guides (`guides/principles/`, `guides/idioms/`, `guides/domain/`, `guides/process/`), the file paths each rule applies to.
- **Exit condition** — no IRON LAW violations.
- **Output** — pass/fail. Findings list each violation with rule reference, file, and reason. GOLDEN RULE violations surface as warnings, do not fail.
- **State writes** — discrepancies discovered during the **audit** action may become Pruning flywheel candidates (archive proposals). Promotions go through the audit flow, not silent writes.
