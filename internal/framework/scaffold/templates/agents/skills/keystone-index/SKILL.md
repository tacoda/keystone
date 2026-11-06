---
kind: skill
id: keystone:index
description: Regenerate .charter/INDEX.json — the primitive descriptor index every agent reads at session start.
triggers:
  - keystone index
  - keystone:index
  - /keystone:index
  - refresh the keystone index
  - regenerate INDEX.json
  - reindex primitives
---

# keystone:index — refresh the primitive descriptor index

The agent's session-start path is "read `.charter/INDEX.json`, then
open primitive bodies on demand." When canonical charter files change,
the index goes stale and the agent's view of what is available diverges
from what is on disk.

This skill regenerates the index by invoking the keystone CLI.

## Run

From the project root:

```
keystone index
```

The command writes `.charter/INDEX.json` and prints the primitive
count. Idempotent — safe to run repeatedly.

## When to trigger

- After authoring or editing any file under `.charter/`.
- After running `keystone migrate` (the migrator regenerates the index
  automatically, but a re-run is cheap if drift is suspected).
- Before reading `.charter/INDEX.json` later in the same session if
  any primitive file has changed since session start.

## Failure modes

- "keystone: command not found" — the binary is not on PATH. Confirm
  with `which keystone`.
- "parse frontmatter: ..." on stderr — a primitive file has malformed
  YAML frontmatter; the indexer reports the path. Fix the file and
  re-run.
