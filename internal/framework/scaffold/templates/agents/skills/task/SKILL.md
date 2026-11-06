---
kind: playbook
id: task
description: 'End-to-end task workflow.'
---
# task

**End-to-end task workflow.** Orchestrates `spec → orient → (implementation) → check-drift → verify → review` and an optional `learn` pass at the end. The canonical phrase to kick off a unit of work.

**Invoke as:** "run task on `<ticket-or-description>`" — or just "run the task workflow."

## Activities

The agent walks through each phase below in order. After each phase, **pause for the user's acceptance** before proceeding to the next. Do not race ahead.

1. **spec** — read [`spec.md`](actions/spec.md). Author the spec, restate intent, list acceptance criteria, list non-goals, flag uncertainty. Save it. **Gate:** explicit user acceptance.
2. **orient** — read [`orient.md`](actions/orient.md). Read `CODEBASE_STATE.md`, load idioms for the touched region, sketch a plan. **Gate:** explicit user acceptance of the plan.
3. **implementation** — read [`charter/guides/process/implementation.md`](guides/process/implementation.md). Make the changes inside the loaded idioms. Iron law: surgical edits only — touch what the spec requires.
4. **check-drift** — read [`check-drift.md`](actions/check-drift.md). Fast diff-vs-guides comparison before running heavyweight sensors.
5. **verify** — read [`verify.md`](actions/verify.md). Run lint, type-check, test, build, drift, commit-message sensors in this turn. Iron law: no completion claims without fresh evidence.
6. **review** — read [`review.md`](actions/review.md). Run functional / security / risk / deployment review (parallel sub-agents if supported) plus spec-adherence against the acceptance criteria.
7. **learn** *(conditional)* — if something surprising came up during the task, read [`learn.md`](actions/learn.md) and capture it to `learning/inbox/`. Otherwise skip.

## Iron laws (carry across every phase)

- **No proceeding without explicit acceptance criteria.** The spec gate is real.
- **No completion claims without fresh verification evidence.** Sensors must run this turn.
- **No commits with failing sensors.** Never `--no-verify`.
- **No AI attribution** in commits, PRs, or tracker comments.
- **No silent overwrites** of state files.

## Pacing

Read [`charter/guides/process/modes.md`](guides/process/modes.md) to learn the current pacing mode (paired / solo / autopilot). In `paired`, confirm before non-trivial edits inside implementation; in `solo`, proceed and ask only at genuine forks; in `autopilot`, execute end-to-end and pause only on iron-law violations or destructive actions.

## When *not* to use task

- One-off questions or short edits that don't need a spec. Skip directly to the relevant action.
- The very first session on a new project. Run [`bootstrap.md`](actions/bootstrap.md) first so `CODEBASE_STATE.md` exists.
- Periodic charter hygiene. Use [`audit.md`](actions/audit.md) instead.
