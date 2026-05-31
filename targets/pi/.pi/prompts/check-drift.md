---
description: Run the drift sensor on the current diff
---

Run the **check-drift** action of the project harness.

Read `harness/process/implementation.md` (the "Check drift" section).

Steps:

1. Run `git diff` to see the working changes.
2. For each touched file, identify which corpus rules apply (load `harness/principles/*.md` always; `harness/idioms/<stack>/*.md` for matching stacks; `harness/domain/*.md` always).
3. Compare the diff against the loaded rules.
4. Report findings:
   - **IRON LAW violations** → block; return to implementation.
   - **GOLDEN RULE deviations** → warn with reasoning; user decides.
   - **No violations** → report clean and continue.
