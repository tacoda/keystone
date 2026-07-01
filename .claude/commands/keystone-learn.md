---
description: Capture a learning candidate from a surprise, an incident, or a review finding.
---
# learn

**Capture a learning candidate** from a surprise, an incident, or a review finding. Writes to `.charter/learning/inbox/` for later synthesis. Read [`.charter/learning/README.md`](learning/README.md).

## Activities

Write a candidate to `.charter/learning/inbox/<timestamp>-<slug>.md` where:

- `<timestamp>` is `YYYY-MM-DD-HHMM` in UTC
- `<slug>` is a short kebab-case description of the insight

The candidate file uses this shape:

```markdown
---
captured: <ISO date>
source: <what triggered this — review finding, surprise during implementation, etc.>
proposed-layer: <corpus/principles | guides/idioms/<stack> | guides/process | sensor>
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

When the surprise happened in a specific region of the codebase, list the paths in `proposed-globs:`. These are the touched files (or their parent directories' globs) from the interaction that produced the candidate. **Synthesize** uses this as signal when deciding the guide's actual `globs:`.

- If the lesson is cross-cutting (would apply to any file in any stack), **omit the field**. Synthesize will default to no `globs:` on the promoted guide.
- If the lesson is regional, list the patterns that match the affected files. Prefer existing region globs from `corpus/state/CODEBASE_STATE.md` over hand-invented patterns — globs should reflect the real codebase, not the abstraction the rule is about.
- Never set `proposed-globs:` wider than where the surprise actually occurred. Synthesize narrows on user confirmation; widening at learn-time loses the evidence.

## Gate

Learn writes only to the inbox. Promotion into corpus / guides happens in **synthesize**, not here. Capture liberally; prune aggressively later.

## Index refresh

Run `keystone index` after writing the candidate so the new entry appears in `.keystone/INDEX.json`. The `keystone:index` skill wraps the CLI invocation.
