---
description: Run the drift sensor on the current diff
---

Run the **check-drift** action of the project harness.

Read `harness/guides/process/implementation.md` (the "Check drift" section).

Steps:

1. Run `git diff` to see the working changes.
2. For each touched file, identify which rules apply. Always-loaded sources: `harness/guides/**/*.md` (project rules) and `harness/policies/*/guides/**/*.md` (policy rules — universal + any installed org policies). On-demand corpus reasoning: `harness/corpus/idioms/<stack>/*.md` for matching stacks, `harness/corpus/domain/*.md`, and `harness/policies/*/corpus/**/*.md` when a rule's forward-link is followed.
3. Compare the diff against the loaded rules.
4. Report findings:
   - **IRON LAW violations** → block; return to implementation.
   - **GOLDEN RULE deviations** → warn with reasoning; user decides.
   - **No violations** → report clean and continue.
