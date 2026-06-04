---
name: keystone:review
description: Run the four review sensors in parallel sub-agents on the current diff
---

You are running the **review** action. Read `harness/guides/process/review.md`.

## Activities

Spawn the four review sensors in parallel sub-agents using the Agent tool. **One Agent call per sensor, all in a single message** so they run concurrently. Each sub-agent reads its own sensor doc and reports findings.

| Sub-agent | Sensor doc |
|---|---|
| review-functional | `harness/sensors/review-functional.md` |
| review-security | `harness/sensors/review-security.md` |
| review-risk | `harness/sensors/review-risk.md` |
| review-deployment | `harness/sensors/review-deployment.md` |

Also run **spec-adherence** (`harness/sensors/spec-adherence.md`) if a spec was authored — walk acceptance criteria line by line against the diff.

## Output

Collect the sub-agent reports. Present a single combined finding list, grouped by sensor, with severity tags.

## Gate

Don't ship until the review findings are addressed (fix, ignore-with-reason, or punt to a follow-up issue).
