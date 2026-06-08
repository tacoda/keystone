# review

**Run the four review sensors on the current diff.** Semantic check, post-verify. Read [`harness/guides/process/review.md`](../guides/process/review.md).

## Activities

Run the four review sensors. If the agent supports parallel sub-agents (Claude Code, etc.), spawn them concurrently — **one sub-agent per sensor, all in a single message** — and combine their findings. Agents without sub-agent parallelism run the sensors sequentially.

| Sensor | Doc |
|---|---|
| review-functional | [`harness/sensors/review-functional.md`](../sensors/review-functional.md) |
| review-security | [`harness/sensors/review-security.md`](../sensors/review-security.md) |
| review-risk | [`harness/sensors/review-risk.md`](../sensors/review-risk.md) |
| review-deployment | [`harness/sensors/review-deployment.md`](../sensors/review-deployment.md) |

Also run **spec-adherence** ([`harness/sensors/spec-adherence.md`](../sensors/spec-adherence.md)) if a spec was authored — walk acceptance criteria line by line against the diff.

## Output

Single combined finding list, grouped by sensor, with severity tags.

## Gate

Don't ship until the review findings are addressed (fix, ignore-with-reason, or punt to a follow-up issue).
