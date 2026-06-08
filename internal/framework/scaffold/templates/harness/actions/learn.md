# learn

**Capture a learning candidate** from a surprise, an incident, or a review finding. Writes to `harness/learning/inbox/` for later synthesis. Read [`harness/learning/README.md`](learning/README.md).

## Activities

Write a candidate to `harness/learning/inbox/<timestamp>-<slug>.md` where:

- `<timestamp>` is `YYYY-MM-DD-HHMM` in UTC
- `<slug>` is a short kebab-case description of the insight

The candidate file uses this shape:

```markdown
---
captured: <ISO date>
source: <what triggered this — review finding, surprise during implementation, etc.>
proposed-layer: <corpus/principles | guides/idioms/<stack> | guides/process | sensor>
---

## What happened

<concrete observation — code, output, or interaction>

## Why it matters

<the principle, idiom, or rule this implies>

## Proposed change

<the smallest harness edit that would prevent the next incident>
```

## Gate

Learn writes only to the inbox. Promotion into corpus / guides happens in **synthesize**, not here. Capture liberally; prune aggressively later.
