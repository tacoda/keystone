---
description: Capture novel judgment calls to the Learning flywheel inbox
---

Run the **learn** action of the project harness.

Read `harness/guides/process/release.md` (the "Capture learnings" section).

If during this work you encountered a situation not covered by the corpus and made a judgment call, write a candidate to `harness/learning/inbox/<timestamp>-<slug>.md` with this shape:

```markdown
---
captured_at: <ISO timestamp>
phase: <spec|planning|implementation|verification|review|release>
confidence: <low|medium|high>
proposed_target: <principles|idioms|domain|state|process>
---

# <Short title>

## Situation
What you encountered that wasn't covered.

## Decision
What you chose to do.

## Reasoning
Why. Especially: what would have happened with the alternative.

## Proposed rule
A draft of the corpus rule this should become, if confirmed.
```

If nothing in this work was novel — every decision was covered by existing corpus content — skip the write and report "no learnings to capture."

The Learning flywheel doesn't reload until the next `/keystone-synthesize` cycle. `/keystone-learn` writes to the inbox only.
