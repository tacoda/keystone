---
name: keystone-mode
description: Switch the charter pacing mode (paired / solo / autopilot).
invokable: true
---

# mode

**Switch the charter pacing mode** (paired / solo / autopilot). Read [`.agents/rules/process/modes.md`](../rules/process/modes.md).

## Activities

1. If the user named `paired`, `solo`, or `autopilot`, set that as the active mode.
2. If no mode was named, print the current mode and a one-line description of each.
3. Update `.agents/rules/process/modes.md` in place — replace the "Current mode" line. Do not rewrite the rest of the file.

## Mode summary

- **paired** — user is at the keyboard reviewing every step. Confirm before non-trivial edits.
- **solo** — user is around but not watching every step. Run sensors and proceed; ask only at genuine forks.
- **autopilot** — user is away. Execute end-to-end; pause only on iron-law violations or destructive actions.

## Iron law

**No silent overwrites of state files.** Show the `modes.md` diff before saving.
