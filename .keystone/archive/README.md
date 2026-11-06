# Archive

Charter files that the pruning flywheel removed, kept with the reason they were removed. Files here sit outside `.agents/`, so keystone doesn't index them and no host loads them.

## Why not delete?

A rule that held for two years and then went wrong is worth more as a record than as an absence. When someone asks "did we ever consider X?", the archive answers. The `audit` command also reads it, so it doesn't propose a recently pruned idea again.

## What gets pruned

- **Rules churn with the code.** A rule that names a removed API, contradicts a newer rule, or no longer matches how the team writes code is stale. Audit rules on a cadence, such as quarterly or alongside a major refactor.
- **References change rarely.** A reference goes stale when the team's thinking changes, not when the code changes shape. Prune one only on purpose.

## Layout

Mirror the source path under `.agents/`:

```
archive/
├── rules/<topic>/<name>.md
├── references/<topic>/<name>.md   (rare)
└── hooks/<id>.md                  (very rare)
```

Keystone itself archives here too. For example, `keystone migrate` to 5.0 moves unedited 4.x scaffold skills to `archive/5.0/skills/`.

## Entry format

The archived file keeps its content, plus a header added at archive time:

```markdown
---
archived_at: <ISO date>
archived_by: <command or person>
source: <rules | references | hooks>
reason: <factually-wrong | aspirationally-stale | domain-stale | process-stale | strategy-shift>
replaced_by: <path to the new file, or "(none)">
---
```

For a reference, use `strategy-shift` and say what the team now believes that it didn't before.

## After you archive

Run `keystone index && keystone project` so the index and the host copies drop the file. A running session may still hold the archived rule, so start a new one.
