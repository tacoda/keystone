---
kind: guide
id: process/subagent-trust
description: 'The discipline of treating subagent reports as claims to verify, not as evidence to accept.'
---
# Subagent Trust

The discipline of treating subagent reports as claims to verify, not as evidence to accept.

## GOLDEN RULE

- **Aim to verify subagent work by reading the diff, not the report.** A subagent's "done!" describes its intent, not the result. `git diff` is the result.
- **Aim to run sensors after a subagent edits the code.** The parent agent's verify action runs against the actual state, not the subagent's narrative of it.
- **Aim to spawn subagents with structured success criteria.** "Make it green" is a wish; "all tests in `tests/auth/` pass with no skips" is a check.

## RULES

- **No reporting subagent success without checking the diff.** The implementation phase IRON LAW for subagent work: read the changes before you claim them.
- **No accepting a subagent's "verified" claim as fresh evidence.** The verify action re-runs in the parent's turn. Subagent evidence is stale by the time the report arrives.
- **No chaining subagent reports.** If subagent A spawned subagent B, the parent verifies *both* — not just A's summary of B's summary.
- **No silent subagent results.** If a subagent says "I tried X but it did not work," that is information the parent must surface, not absorb.

## Why this is agent-specific

Multi-agent orchestration is a recent capability. The natural failure mode is for the parent to trust the child's self-report — partly because the child is also an agent, partly because the alternative (re-reading every diff) feels redundant. It is not redundant. A subagent has its own context, its own biases, its own room for the same hallucinations and [[self-validation]] failures the parent guards against.

The remedy: the same fresh-evidence rule that protects the parent from itself protects the parent from its children. The diff is the truth.

## Sensors

- The **verify** action runs in the *parent's* turn after subagent work returns. State propagated from the subagent is verified, not inherited.
- The **drift sensor** runs against the post-subagent diff, with the same rules as any other edit.

## Anti-patterns

- "Subagent reported success" as the last word in a PR. Read the diff.
- Accepting "tests pass" from a subagent that did not show the runner output.
- Spawning a subagent without a success criterion. The subagent will declare success on its own terms.
- Forwarding the subagent's summary to the user as a finding without verifying it.

---

See also: `verification.md` (IRON LAW), [[self-validation]], `implementation.md`.
