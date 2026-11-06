---
name: keystone-review
description: 'Run the four review subagents on the current diff (functional, security, risk, deployment). Triggers: keystone review; keystone:review; /keystone:review; review the diff; run review subagents; semantic review.'
allowed-tools: Read Grep Glob Bash Task
metadata:
  model: opus
---

# keystone:review — semantic review of the current diff

Phase 5 of the six-phase lifecycle. Fires the four review
subagents against the in-progress diff:

- **review-functional** — does the change implement the spec?
- **review-security** — auth, input validation, secrets, injection.
- **review-risk** — blast radius, data integrity, observability.
- **review-deployment** — migrations, rollback, config, infra deps.

Each subagent returns a structured finding list. Semantic check — runs
after computational hooks (lint / type / test / build) pass.

Full
discipline at `.agents/rules/process/review.md`.

## Run

**Run the four review subagents on the current diff.** Semantic check, post-verify. Read [`.agents/rules/process/review.md`](../../rules/process/review.md).

## Activities

Run the four review subagents. If the agent supports parallel subagents (Claude Code, etc.), spawn them concurrently: **all four in a single message**. Then combine their findings. Agents without subagent parallelism run them sequentially.

| Subagent | Doc |
|---|---|
| review-functional | [`.agents/agents/review-functional.md`](../../agents/review-functional.md) |
| review-security | [`.agents/agents/review-security.md`](../../agents/review-security.md) |
| review-risk | [`.agents/agents/review-risk.md`](../../agents/review-risk.md) |
| review-deployment | [`.agents/agents/review-deployment.md`](../../agents/review-deployment.md) |

Also run **spec-adherence** ([`.agents/hooks/spec-adherence.md`](../../hooks/spec-adherence.md)) if a spec was authored — walk acceptance criteria line by line against the diff.

## Output

Single combined finding list, grouped by subagent, with severity tags.

## Gate

Don't ship until the review findings are addressed (fix, ignore-with-reason, or punt to a follow-up issue).

## When to trigger

- After `/keystone:verify` is clean.
- Before opening a PR.
- Anytime a change feels load-bearing and a second pass is cheap.
