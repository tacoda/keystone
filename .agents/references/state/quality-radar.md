---
kind: reference
id: corpus/state/quality-radar
description: 'Five-dimension scorecard for the codebase (or for the most recent reviewed diff).'
---
# Quality Radar

> **Template.** The **audit** or **review** command will populate this from the underlying hooks. Until then, leave as-is or fill in by hand.

Five-dimension scorecard for the codebase (or for the most recent reviewed diff). Captured periodically by the [quality-radar subagent](../../agents/quality-radar.md).

## Latest scorecard

| Dimension | Score | Trend vs. last | Citations |
|---|---|---|---|
| Type safety | `<green|yellow|red>` | `<→|↑|↓>` | `<hook + finding refs>` |
| Test quality | `<green|yellow|red>` | `<→|↑|↓>` | `<hook + finding refs>` |
| Readability | `<green|yellow|red>` | `<→|↑|↓>` | `<hook + finding refs>` |
| Security | `<green|yellow|red>` | `<→|↑|↓>` | `<hook + finding refs>` |
| Performance | `<green|yellow|red>` | `<→|↑|↓>` | `<hook + finding refs>` |

## How to use it

- **Two or more reds** → block further feature work until the worst is addressed.
- **One red** → flag in the next planning pass; do not start adjacent work without a containment plan.
- **All-yellow** → drift signal. Run **audit** to see whether the harness needs new rules.
- **All-green** → keep iterating.

## History

Append a new dated section per audit pass. Do not overwrite — trend matters more than any single score.

```
## YYYY-MM-DD

| Dimension | Score | Citations |
|---|---|---|
| ... | ... | ... |
```

## Scoring conventions

Per-dimension thresholds live in [`.agents/agents/quality-radar.md`](../../agents/quality-radar.md). If your project needs different thresholds (e.g., a research codebase where coverage drops are expected), edit the subagent and document the deviation here.
