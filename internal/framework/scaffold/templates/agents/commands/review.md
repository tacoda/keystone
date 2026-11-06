---
kind: command
id: review
description: 'Run the four review hooks on the current diff.'
---
# review

**Run the four review hooks on the current diff.** Semantic check, post-verify. Read [`.agents/rules/process/review.md`](../rules/process/review.md).

## Activities

Run the four review hooks. If the agent supports parallel subagents (Claude Code, etc.), spawn them concurrently: **one subagent per hook, all in a single message**. Then combine their findings. Agents without subagent parallelism run the hooks sequentially.

| Hook | Doc |
|---|---|
| review-functional | [`.agents/hooks/review-functional.md`](../hooks/review-functional.md) |
| review-security | [`.agents/hooks/review-security.md`](../hooks/review-security.md) |
| review-risk | [`.agents/hooks/review-risk.md`](../hooks/review-risk.md) |
| review-deployment | [`.agents/hooks/review-deployment.md`](../hooks/review-deployment.md) |

Also run **spec-adherence** ([`.agents/hooks/spec-adherence.md`](../hooks/spec-adherence.md)) if a spec was authored — walk acceptance criteria line by line against the diff.

## Output

Single combined finding list, grouped by hook, with severity tags.

## Gate

Don't ship until the review findings are addressed (fix, ignore-with-reason, or punt to a follow-up issue).
