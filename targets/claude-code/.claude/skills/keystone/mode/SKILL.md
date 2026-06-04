---
name: keystone:mode
description: Switch the harness pacing mode (paired / solo / autopilot)
argument-hint: "paired | solo | autopilot"
---

You are running the **mode** action. Read `harness/guides/process/modes.md`.

## Activities

1. If `$1` is `paired`, `solo`, or `autopilot`, set that as the active mode.
2. If `$1` is empty or `--show`, print the current active mode and a one-line description of each.
3. Update `harness/guides/process/modes.md` in place — replace the "Current mode" line. Do not rewrite the rest of the file.

## Mode summary

- **paired** — user is at the keyboard reviewing every step. Confirm before non-trivial edits.
- **solo** — user is around but not watching every step. Run sensors and proceed; ask only at genuine forks.
- **autopilot** — user is away. Execute end-to-end; pause only on iron-law violations or destructive actions.

## Iron law

**No silent overwrites of state files.** Show the modes.md diff before saving.
