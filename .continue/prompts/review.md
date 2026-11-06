---
name: review
description: Run the four review subagents on the current diff.
invokable: true
---

# review

**Run the four review subagents on the current diff.** Semantic check, post-verify. Read [`.agents/rules/process/review.md`](../rules/process/review.md).

## Activities

Run the four review subagents. If the agent supports parallel subagents (Claude Code, etc.), spawn them concurrently: **all four in a single message**. Then combine their findings. Agents without subagent parallelism run them sequentially.

| Subagent | Doc |
|---|---|
| review-functional | [`.agents/agents/review-functional.md`](../agents/review-functional.md) |
| review-security | [`.agents/agents/review-security.md`](../agents/review-security.md) |
| review-risk | [`.agents/agents/review-risk.md`](../agents/review-risk.md) |
| review-deployment | [`.agents/agents/review-deployment.md`](../agents/review-deployment.md) |

Also run **spec-adherence** ([`.agents/hooks/spec-adherence.md`](../hooks/spec-adherence.md)) if a spec was authored — walk acceptance criteria line by line against the diff.

## Output

Single combined finding list, grouped by subagent, with severity tags.

## Gate

Don't ship until the review findings are addressed (fix, ignore-with-reason, or punt to a follow-up issue).
