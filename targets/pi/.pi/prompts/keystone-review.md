---
description: Run the review phase — spec adherence + functional + security review
---

Run the **review** action of the project harness.

Read `harness/guides/process/review.md`.

Pi does not support parallel sub-agents natively. Run the review steps sequentially:

1. **Locate the spec.** Find the most recent file under `docs/specs/` matching the current task.
2. **Spec adherence** — walk through the spec's `## Acceptance criteria`. For each criterion, decide: met by the diff (with evidence)? unmet? not yet covered? Report per-criterion.
3. **Functional review** — read the diff as a careful peer reviewer. Look for: logic errors, edge cases, dead code, broken abstractions, missing tests. List findings with severity (blocker / major / minor).
4. **Security review** — read the diff for OWASP-style risks: injection, secrets, auth bypass, unsafe deserialization, broken access control, sensitive data exposure. List findings with severity.
5. **Combine** — merge the three lists into one severity-sorted table.

If any blocker remains, return to implementation. After fixing, re-run `/keystone-verify` (fresh evidence) and then re-run `/keystone-review`.

GOLDEN RULE: do not approve a PR by inspection alone — the acceptance criteria must be verifiable.
