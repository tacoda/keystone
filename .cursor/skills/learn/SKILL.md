---
name: learn
description: Capture a learning candidate from a surprise, an incident, or a review finding.
disable-model-invocation: true
---

# learn

**Capture a learning candidate** from a surprise, an incident, or a review finding. Writes to `.keystone/learning/inbox/` for later synthesis. Read [`.keystone/learning/README.md`](../../../.keystone/learning/README.md).

## Activities

Write a candidate to `.keystone/learning/inbox/<timestamp>-<slug>.md` where:

- `<timestamp>` is `YYYY-MM-DD-HHMM` in UTC
- `<slug>` is a short kebab-case description of the insight

The candidate file uses this shape:

```markdown
---
captured: <ISO date>
source: <what triggered this — review finding, surprise during implementation, etc.>
proposed-layer: <references/principles | rules/idioms/<stack> | rules/process | hook>
proposed-globs:                       # optional; see below
  - "src/billing/**"
  - "tests/billing/**"
---

## What happened

<concrete observation — code, output, or interaction>

## Why it matters

<the principle, idiom, or rule this implies>

## Proposed change

<the smallest charter edit that would prevent the next incident>
```

### `proposed-globs:` — record the paths the lesson came from

When the surprise happened in a specific region of the codebase, list the paths in `proposed-globs:`. These are the touched files (or their parent directories' globs) from the interaction that produced the candidate. **Synthesize** uses this as signal when deciding the rule's actual `paths:`.

- If the lesson is cross-cutting (would apply to any file in any stack), **omit the field**. Synthesize will default to no `paths:` on the promoted rule.
- If the lesson is regional, list the patterns that match the affected files. Prefer existing region globs from `.agents/references/state/CODEBASE_STATE.md` over hand-invented patterns. The globs should reflect the real codebase, not the abstraction the rule is about.
- Never set `proposed-globs:` wider than where the surprise actually occurred. Synthesize narrows on user confirmation; widening at learn-time loses the evidence.

## Gate

Learn writes only to the inbox. Promotion into references / rules happens in **synthesize**, not here. Capture liberally; prune aggressively later.

## Index refresh

Run `keystone index` after writing the candidate so the new entry appears in `.keystone/INDEX.json`. The `keystone-index` skill wraps the CLI invocation.
