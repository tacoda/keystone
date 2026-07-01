---
kind: skill
id: keystone:verify
description: Check vendored policies for drift and the strict cascade for project-layer violations.
triggers:
  - keystone verify
  - keystone:verify
  - /keystone:verify
  - verify the charter cascade
  - audit policy drift
  - check strict violations
model: opus
tools:
  - Read
  - Bash
  - Grep
  - Glob
---

# keystone:verify — cascade + drift check

Walks every vendored policy under `.charter/policies/`,
compares per-file hashes to the lockfile, and reports project files
that shadow strict-locked policy items.

Exits non-zero on cascade violation. Drift alone exits zero but is
reported and the affected policy is reset (re-install needed to
repopulate).

## Run

```
keystone verify
```

## When to trigger

- Before committing any charter change — catches accidental shadows of
  strict policy items.
- After `keystone install` or `keystone policy update` — confirms the
  install is clean.
- In CI to gate merges on cascade integrity.

## Interpreting output

- `✓ keystone verify clean` — pass.
- `✗ N strict violation(s)` — remove the offending project file(s) or
  surface the conflict to the policy author.
- `▸ drift detected — resetting N policies` — vendored files diverged
  from the lockfile; re-run `keystone install` to restore.
