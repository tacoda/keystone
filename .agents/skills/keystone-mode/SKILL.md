---
name: keystone-mode
description: 'Switch the charter pacing mode (paired / solo / autopilot). Triggers: keystone mode; keystone:mode; /keystone:mode; switch pacing mode; go autopilot; go solo; go paired.'
allowed-tools: Read Write Edit
metadata:
  model: sonnet
---

# keystone:mode — switch pacing mode

Toggles how often the agent stops to ask. Three modes:

- **paired** — agent confirms at every gate; default for new projects.
- **solo** — agent runs through phases, asks at unclear forks.
- **autopilot** — agent runs end-to-end; asks only on hard blockers.

Mode definitions
live in `.agents/rules/process/modes.md`.

## Run

**Switch the charter pacing mode** (paired / solo / autopilot). Read [`.agents/rules/process/modes.md`](../../rules/process/modes.md).

## Activities

1. If the user named `paired`, `solo`, or `autopilot`, set that as the active mode.
2. If no mode was named, print the current mode and a one-line description of each.
3. Update `.agents/rules/process/modes.md` in place: replace the `**Active mode:**` line. Do not rewrite the rest of the file.

## Mode summary

- **paired** — user is at the keyboard reviewing every step. Confirm before non-trivial edits.
- **solo** — user is around but not watching every step. Run hooks and proceed; ask only at genuine forks.
- **autopilot** — user is away. Execute end-to-end; pause only on iron-law violations or destructive actions.

## Iron law

**No silent overwrites of state files.** Show the `modes.md` diff before saving.

## When to trigger

- Start of a task: pick paired for risky work, autopilot for routine sweeps.
- Mid-task: dial in or out as confidence shifts.
