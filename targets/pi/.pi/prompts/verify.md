---
description: Run every verification sensor with fresh evidence
---

Run the **verify** action of the project harness.

Read `harness/guides/process/verification.md`. The IRON LAW applies: **no completion claims without fresh verification evidence**. Run every sensor in this turn. Stale output from a previous turn does not count.

Steps:

1. Read the tool commands from `harness/corpus/state/CODEBASE_STATE.md` (lint, type-check, test, build, coverage).
2. Run each sensor via shell:
   - `lint` — exit 0, 0 errors
   - `type-check` — exit 0, 0 errors (skip if no type checker)
   - `test` — exit 0, 0 failures
   - `build` — exit 0
3. Run the **drift** sensor (compare diff against loaded corpus rules).
4. Run the **commit-message** sensor on the proposed commit message: conventional format, title under 70 chars, no AI attribution.
5. Read the diff and propose state-layer updates (coverage delta, region recency) as a diff against `harness/corpus/state/CODEBASE_STATE.md`. Do not silently overwrite — propose, then confirm.

Report:

| Sensor | Status | Evidence |
|---|---|---|
| lint | pass/fail | command output excerpt |
| type-check | ... | ... |
| ... | ... | ... |

If any sensor fails, return to the implementation phase. Do not "fix forward" inside verify.
