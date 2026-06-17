---
kind: sensor
id: drift
description: 'Compares the diff (or full codebase, during audit) against loaded guide rules.'
---
# Sensor: drift

Compares the diff (or full codebase, during audit) against loaded guide rules. Finds violations across all three rule tiers — regular **RULES**, **GOLDEN PATH**, and **IRON LAWS**.

- **Trigger** — **check-drift** (implementation), **verify** (verification), **audit** (discipline).
- **Inputs** — current diff (or file set for audit), loaded guides (`guides/principles/`, `guides/idioms/`, `guides/domain/`, `guides/process/`), and the per-guide `globs:` from frontmatter (if present) that say which paths the guide claims.
- **Glob filtering** — for each loaded guide, compare findings only against files matching its `globs:`. Findings outside the guide's globs are dropped before tier classification — a rule cannot flag a violation in a file it does not claim. Guides without `globs:` apply to every file in the input set (today's behavior).
- **Exit condition** — no IRON LAW violations.
- **Output** — pass/fail. Findings list each violation with rule reference, file, and reason. Severity by tier:
  - **IRON LAW** violation → fail.
  - **GOLDEN RULE** violation → warning (strong; deviation requires reasoning in the diff).
  - **RULES** (regular) violation → warning.
- **State writes** — discrepancies discovered during the **audit** action may become Pruning flywheel candidates (archive proposals). Promotions go through the audit flow, not silent writes.
