---
name: task
description: End-to-end task workflow.
---
# task

**End-to-end task workflow.** Orchestrates `spec → orient → (implementation) → check-drift → verify → review` and an optional `learn` pass at the end. The canonical phrase to kick off a unit of work.

**Invoke as:** "run task on `<ticket-or-description>`" — or just "run the task workflow."

## Activities

The agent walks through each phase below in order. After each phase, **pause for the user's acceptance** before proceeding to the next. Do not race ahead.

1. **spec** — read [`spec.md`](../../skills/keystone-spec/SKILL.md). Author the spec, restate intent, list acceptance criteria, list non-goals, flag uncertainty. Save it. **Gate:** explicit user acceptance.
2. **orient** — read [`orient.md`](../../skills/keystone-orient/SKILL.md). Read `CODEBASE_STATE.md`, load idioms for the touched region, sketch a plan. **Gate:** explicit user acceptance of the plan.
3. **implementation** — read [`.agents/rules/process/implementation.md`](../../rules/process/implementation.md). Make the changes inside the loaded idioms. Iron law: surgical edits only — touch what the spec requires.
4. **check-drift** — read [`check-drift.md`](../../skills/keystone-check-drift/SKILL.md). Fast diff-vs-rules comparison before running heavyweight hooks.
5. **verify** — read [`verify.md`](../../skills/keystone-verify/SKILL.md). Run the lint, type-check, test, build, drift, and commit-message hooks in this turn. Iron law: no completion claims without fresh evidence.
6. **review** — read [`review.md`](../../skills/keystone-review/SKILL.md). Run functional / security / risk / deployment review (parallel sub-agents if supported) plus spec-adherence against the acceptance criteria.
7. **learn** *(conditional)* — if something surprising came up during the task, read [`learn.md`](../../skills/keystone-learn/SKILL.md) and capture it to `.keystone/learning/inbox/`. Otherwise skip.

## Iron laws (carry across every phase)

- **No proceeding without explicit acceptance criteria.** The spec gate is real.
- **No completion claims without fresh verification evidence.** Hooks must run this turn.
- **No commits with failing hooks.** Never `--no-verify`.
- **No AI attribution** in commits, PRs, or tracker comments.
- **No silent overwrites** of state files.

## Pacing

Read [`.agents/rules/process/modes.md`](../../rules/process/modes.md) to learn the current pacing mode (paired / solo / autopilot). In `paired`, confirm before non-trivial edits inside implementation; in `solo`, proceed and ask only at genuine forks; in `autopilot`, execute end-to-end and pause only on iron-law violations or destructive actions.

## When *not* to use task

- One-off questions or short edits that don't need a spec. Skip directly to the relevant command.
- The very first session on a new project. Run [`bootstrap.md`](../../skills/keystone-bootstrap/SKILL.md) first so `CODEBASE_STATE.md` exists.
- Periodic charter hygiene. Use [`audit.md`](../../skills/keystone-audit/SKILL.md) instead.
