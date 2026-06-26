---
kind: guide
id: process/modes
description: 'Active mode: paired'
---
# Pacing Modes

**Active mode: paired**

How the agent paces work. Orthogonal to the four phases — modes change *how* phases run, not *what* happens in them.

Set by invoking the **mode** action with `<paired|solo|autopilot>`. The `**Active mode:**` line above is the source of truth and is updated in place. The binding for your agent lives in `harness/adapters/<your-agent>/lifecycle.md`. Agents that don't have a notion of autonomy levels may ignore modes entirely.

## paired

Pair programming. The user drives the pace.

- Ask for feedback at each phase boundary.
- Ask at confirmation prompts in commands.
- Stop and surface trade-offs at any decision point with non-trivial consequences.

The right mode for: unfamiliar code, security-sensitive work, design decisions, work where the user wants visibility.

## solo

The agent works independently but raises a hand on hard problems.

- Do **not** ask for feedback at routine phase boundaries.
- Iterate until the work is done.
- **Stop and report findings** — do not guess — when:
  - Requirements are ambiguous.
  - Rules conflict.
  - A test failure is inscrutable.
  - A choice has non-trivial consequences and no clear winner.

The right mode for: well-scoped work, code the agent has touched before, work the user is OK reviewing only at the end.

## autopilot

Maximally autonomous. The user engages at kickoff and at the final review.

- Do **not** ask for feedback at routine phase boundaries.
- Do **not** stop on ambiguity — make the best-effort decision and continue. Note the assumption inline so the final-review reader can audit it.
- Iterate end-to-end.
- Surface every assumption, fallback, and judgment call in a structured **assumption log** at the end of the run, alongside the diff and any verification results.

The right mode for: trivial work, repetitive tasks, work the user cannot supervise for hours.

## What all modes preserve

Three things are mode-invariant:

- **Plan-first behavior.** Planning still happens; only the user-facing "ask" is skipped in solo and autopilot. There is no mode where "skip planning" is the answer.
- **Scaffolding safety contract.** Never silently overwrite a consumer's file. All modes ask before overwriting; that is a safety rule, not a pace-setting checkpoint.
- **User ownership of commits and pushes.** Every mode iterates on *work*; the user still owns the *publish*. Even in autopilot, the agent does not push to remote without explicit user instruction.

## IRON LAW

**A MODE NEVER OVERRIDES A SAFETY RULE.**

If a mode would skip a sensor, skip a gate, or bypass the scaffolding safety contract — that is the mode being wrong, not the safety rule. Modes change the conversational pace. They do not change what makes the work correct.

## Switching modes mid-task

Allowed. Invoking the **mode** action with a new value takes effect on the next phase boundary. The current phase finishes in the current mode; the next phase enters the new one.

A common pattern: start in paired for planning, switch to solo for implementation, return to paired for verification.
