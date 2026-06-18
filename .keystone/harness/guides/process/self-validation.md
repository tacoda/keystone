---
kind: guide
id: process/self-validation
description: 'The discipline of refusing to count the agent''s own claim as evidence.'
---
# Self-Validation

The discipline of refusing to count the agent's own claim as evidence. Operationalizes the verification IRON LAW for the specific case where the agent re-reads its own previous statement and treats it as proof.

## RULES

- **A prior message saying "tests pass" is not evidence.** The test runner saying tests pass — *in this turn* — is evidence. See `verification.md`.
- **Re-stating the conclusion is not the same as re-running the check.** "As I said, the linter is clean" — was the linter run *after the last edit*? If not, the claim is stale.
- **An agent's reading of its own diff is not a review.** The review phase exists because the writer is not a reliable reviewer of their own work. The agent is the writer; the review agents are separate.
- **A subagent's "done" report is not evidence.** See [[subagent-trust]].

## GOLDEN PATH

- **Aim to re-run sensors after every edit, not after every claim.** The freshness boundary is the *edit*, not the *statement*.
- **Aim to cite the tool output, not the tool name.** "Lint passes" is a claim; "`eslint .` exited 0 with 0 errors at HH:MM" is evidence.

## Why this is agent-specific

A coding agent generates text continuously. The agent's own previous output is in the context window, and the agent will read it back as if it were external. This produces a closed loop: the agent says "the tests pass," then later reads that message and concludes "the tests pass," with no actual test runner involvement. The behavior is not deception — it is the absence of an external check.

The remedy is to anchor evidence to *tool output*, not to *text in the conversation*. The verification IRON LAW does this at the phase level; this guide does it at the within-phase level.

## Sensors

The **verify** action surfaces fresh evidence; re-invoking it after any edit is the operational form of this rule.

## Anti-patterns

- "I already confirmed that earlier" — earlier is not now.
- Reading the agent's own prior summary of test output instead of the test output itself.
- Treating an analytical claim ("this function looks correct") as equivalent to a verification claim ("the test for this function passes").
- "I'm confident this works" — confidence is not evidence.

---

See also: `verification.md` (IRON LAW), [[subagent-trust]], [[grounding]].
