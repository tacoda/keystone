---
kind: skill
id: keystone:verify
description: Check vendored plugins for drift and the strict cascade for project-layer violations.
triggers:
  - keystone verify
  - keystone:verify
  - /keystone:verify
  - verify the harness cascade
  - audit plugin drift
  - check strict violations
---

# keystone:verify — cascade + drift check

Walks every vendored plugin under `.keystone/harness/plugins/`,
compares per-file hashes to the lockfile, and reports project files
that shadow strict-locked plugin items.

Exits non-zero on cascade violation. Drift alone exits zero but is
reported and the affected plugin is reset (re-install needed to
repopulate).

## Run

```
keystone verify
```

## When to trigger

- Before committing any harness change — catches accidental shadows of
  strict plugin items.
- After `keystone install` or `keystone plugin update` — confirms the
  install is clean.
- In CI to gate merges on cascade integrity.

## Interpreting output

- `✓ keystone verify clean` — pass.
- `✗ N strict violation(s)` — remove the offending project file(s) or
  surface the conflict to the plugin author.
- `▸ drift detected — resetting N plugin(s)` — vendored files diverged
  from the lockfile; re-run `keystone install` to restore.
