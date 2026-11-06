# Process rules

The workflow phases and the discipline that applies across them. These rules declare no `paths:`, so each one applies everywhere. Keystone lists each one in the `AGENTS.md` managed block, and the agent opens the one that fits the phase it is in.

## Phases

Each phase file states its entry condition, activities, checks, gate, and artifacts.

| File | Entry | Gate |
|---|---|---|
| `spec.md` | An idea or a tracker card | An approved spec with explicit acceptance criteria |
| `planning.md` | Spec approved | Plan approved |
| `implementation.md` | Plan approved | Tests pass, checks clean |
| `verification.md` | Implementation done | Every check green this turn |
| `review.md` | Verification passed | No blockers, acceptance criteria met |
| `release.md` | Review passed | Change deployed and recorded |

`modes.md` sets the pacing: paired, solo, or autopilot. A mode changes how the phases run, not what happens in them.

## Discipline across every phase

- `sensitive-files.md`: files the agent never reads, writes, or commits.
- `dangerous-actions.md`: operations that need explicit confirmation first.
- `scoping.md`: size limits on commits and PRs.
- `ci-failure.md`: what to do when CI fails.
- `escalation.md`: when to stop and ask.
- `runtime-resolution.md`: how to look up an answer, from rules to references to outside sources.

## Agent failure modes

These counter the common ways a coding agent goes wrong:

- `grounding.md`: check that a function, package, flag, or config key exists before using it.
- `pushback.md`: disagree when the user is wrong.
- `self-validation.md`: tool output is evidence, and the agent's own claim isn't.
- `subagent-trust.md`: a subagent's "done" is a claim to verify.
- `context-budget.md`: read what the touched region needs, and no more.
- `surgical-edits.md`: touch only what serves the task.

## Commands per phase

| Phase | Command |
|---|---|
| Spec | `spec` |
| Planning | `orient` |
| Implementation exit | `check-drift` |
| Verification | `verify` |
| Review | `review` |
| Release | `learn` |

The [`task`](../../skills/task/SKILL.md) skill chains them for one unit of work.

## Tracker cards

The spec phase records the tracker card id when one exists. Every later phase links back to it: the PR, the postmortem, and any learning capture. Without a card, each phase asks you for its input.

## Artifact locations

| Artifact | Location |
|---|---|
| Spec | `docs/specs/YYYY-MM-DD-<topic>.md` |
| Plan | `docs/plans/YYYY-MM-DD-<topic>.md` |
| ADR | `docs/adrs/NNNN-<slug>.md` |
| Postmortem | `docs/postmortems/YYYY-MM-DD-<slug>.md` |
