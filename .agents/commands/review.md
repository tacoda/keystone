---
kind: command
id: review
description: 'Run the four review sensors on the current diff.'
---
# review

**Run the four review sensors on the current diff.** Semantic check, post-verify. Read [`.agents/rules/process/review.md`](../rules/process/review.md).

## Activities

Run the four review sensors. If the agent supports parallel sub-agents (Claude Code, etc.), spawn them concurrently — **one sub-agent per sensor, all in a single message** — and combine their findings. Agents without sub-agent parallelism run the sensors sequentially.

| Sensor | Doc |
|---|---|
| review-functional | [`.agents/hooks/review-functional.md`](../agents/review-functional.md) |
| review-security | [`.agents/hooks/review-security.md`](../agents/review-security.md) |
| review-risk | [`.agents/hooks/review-risk.md`](../agents/review-risk.md) |
| review-deployment | [`.agents/hooks/review-deployment.md`](../agents/review-deployment.md) |

Also run **spec-adherence** ([`.agents/hooks/spec-adherence.md`](../hooks/spec-adherence.md)) if a spec was authored — walk acceptance criteria line by line against the diff.

## Output

Single combined finding list, grouped by sensor, with severity tags.

## Gate

Don't ship until the review findings are addressed (fix, ignore-with-reason, or punt to a follow-up issue).
