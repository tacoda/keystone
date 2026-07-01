---
kind: guide
id: process/review
description: 'The phase that confirms a mechanically-correct change is also the right change.'
---
# Review

The phase that confirms a mechanically-correct change is also the *right* change. Where verification asks "does it work?", review asks "is it what we wanted, and is it good?"

## Entry condition

The verification phase gate passed. Mechanical sensors are green with fresh evidence. The diff exists and is ready for review.

## Activities

### 1. Locate the spec

Find the spec file — most recent under `docs/specs/` matching the task. If the spec references a `tracker_card`, carry that reference forward into PR descriptions and any review artifacts.

### 2. Spawn review agents

Spawn every available review agent (named `review-*`) in parallel on the current diff. The default set is `review-functional`, `review-security`, `review-risk`, and `review-deployment`; teams may add more (`review-performance`, `review-accessibility`, project-specific reviewers).

Where review agents live, and how they are spawned, is agent-specific — see `charter/adapters/<your-agent>/lifecycle.md`. Agents that do not support sub-agent parallelism collapse this into sequential review passes.

Each agent returns a numbered list of findings, graded:
- **blocker** — must be fixed before release.
- **major** — should be fixed; if deferred, requires an explicit follow-up tracker card.
- **minor** — cleanup; deferrable.

### 3. Check spec adherence

Walk through the spec's `## Acceptance criteria`. For each:

- Is it met by the diff?
- Is there evidence (a test, an output, a manual verification)?

A criterion not met is a blocker, regardless of how clean the sensors are.

### 4. Peer review (when a PR exists)

If the diff is in a PR, the agent can:

- Summarize the diff for human reviewers, with the tracker card linked.
- Address comments as they arrive: propose fix, apply with confirmation, mark thread resolved.
- Re-invoke the **verify** action after each comment-driven change (locked: fresh evidence per turn).

### 5. Combine findings

Merge review-agent findings, spec-adherence gaps, and PR comments into a single list:

| Severity | Source | Finding | Action |
|---|---|---|---|
| blocker | review-security | <description> | fix |
| blocker | spec adherence | criterion 3 not met | fix |
| major | review-functional | <description> | fix or follow-up |
| minor | review-functional | <description> | defer |

### 6. Resolve

For each finding:
- **blocker** → return to implementation. Re-invoke the **verify** action after the fix.
- **major** → fix in this PR, or open a follow-up tracker card and link it from the spec.
- **minor** → defer or fix; the user decides.

### 7. Record (optional)

If the team archives review findings, write to `docs/reviews/<spec-slug>.md`. Useful during postmortems and audits.

## Sensors

Review runs semantic sensors (the review agents — see `sensors.md`). The mechanical sensors do not re-run here by default; they already gated verification. They re-run only if implementation re-opens.

## Gate condition

To exit review:

1. No blocker findings remain.
2. Every acceptance criterion in the spec is met and verifiable.
3. PR comments (if any) are resolved or explicitly deferred with a follow-up tracker card linked.
4. Major findings are either fixed or have follow-up tracker cards linked.

## Artifacts

| Kind | Location |
|---|---|
| Review findings record | `docs/reviews/<spec-slug>.md` (optional) |
| Follow-up tracker cards | tracker (Jira/Linear/GitHub) |

## Pacing modes

- **paired** — user reviews findings before fix-or-defer decisions.
- **solo** — agent handles minor and major findings autonomously; raises blockers to the user.
- **autopilot** — agent fixes everything it can, defers what it cannot, surfaces the assumption log at the end.

## Anti-patterns

- Folding review back into verification — they answer different questions (mechanical vs semantic).
- Approving a PR without checking acceptance criteria from the spec.
- Resolving a blocker with "looks fine to me, ship it" instead of evidence.
- Deferring a major finding without a follow-up tracker card.
