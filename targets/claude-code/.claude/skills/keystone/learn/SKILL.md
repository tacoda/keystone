---
name: keystone:learn
description: Capture a learning candidate into harness/learning/inbox/ for later synthesis
argument-hint: "<short slug describing the insight>"
---

You are running the **learn** action. Read `harness/guides/process/review.md` (learn flywheel section) and `harness/learning/README.md`.

## Activities

Write a candidate learning to `harness/learning/inbox/<timestamp>-<slug>.md` where:

- `<timestamp>` is `YYYY-MM-DD-HHMM` in UTC
- `<slug>` is a short kebab-case description (`$1` if provided)

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

Learn writes only to the inbox. Promotion into corpus/guides happens in **synthesize**, not here. Capture liberally; prune aggressively later.
