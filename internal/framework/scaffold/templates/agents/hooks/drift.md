---
kind: hook
mode: inferential
returns: review-findings
id: drift
description: 'Compares the diff (or full codebase, during audit) against loaded rules.'
---
# Hook: drift

Compares the diff (or full codebase, during audit) against loaded rules. Finds violations across all three rule tiers — regular **RULES**, **GOLDEN RULE**, and **IRON LAWS**.

- **Trigger** — **check-drift** (implementation), **verify** (verification), **audit** (discipline).
- **Inputs** — current diff (or file set for audit), loaded rules (`.agents/rules/idioms/`, `.agents/rules/domain/`, `.agents/rules/process/`), and each rule's `paths:` from frontmatter (if present), which say which files the rule claims.
- **Path filtering** — for each loaded rule, compare findings only against files matching its `paths:`. Drop findings outside the rule's paths before tier classification: a rule cannot flag a violation in a file it does not claim. Rules without `paths:` apply to every file in the input set (today's behavior).
- **Exit condition** — no IRON LAW violations.
- **Output** — pass/fail. Findings list each violation with rule reference, file, and reason. Severity by tier:
  - **IRON LAW** violation → fail.
  - **GOLDEN RULE** violation → warning (strong; deviation requires reasoning in the diff).
  - **RULES** (regular) violation → warning.
- **State writes** — discrepancies discovered during the **audit** action may become Pruning flywheel candidates (archive proposals). Promotions go through the audit flow, not silent writes.
