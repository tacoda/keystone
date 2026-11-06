---
name: keystone-bootstrap
description: 'One-time initial charter scaffold — detect stack, seed state, scaffold idioms, classify hooks, build the globs index. Triggers: keystone bootstrap; keystone:bootstrap; /keystone:bootstrap; run bootstrap; bootstrap the charter; bootstrap the keystone charter; first-time setup.'
---

# keystone:bootstrap — make the charter specific to your codebase

The first thing to run after `keystone init`. Walks the repo, detects
the real stack, and replaces every template placeholder under
`.agents/references/state/` with code-grounded values.

The canonical procedure is the **bootstrap command** at
`.agents/commands/bootstrap.md`. This skill is a thin
trigger that points the agent at it. Open the command and execute its
activities in order.

## Run

Open `.agents/commands/bootstrap.md` and execute every activity
listed under `## Activities`. Every activity must land a real file
change before moving on; narration does not count.

A short summary of the work, copied from the command:

1. Detect the stack (`package.json`, `go.mod`, `pyproject.toml`, …).
2. Seed `.agents/references/state/CODEBASE_STATE.md` with real tool commands +
   region map.
3. Scaffold `.agents/references/idioms/<stack>/` + paired `.agents/rules/idioms/<stack>/`
   with `paths:` derived from the region map.
4. Inventory `.agents/rules/computational/` from `.editorconfig`, LSP config,
   formatter configs.
5. Classify every hook under `.agents/hooks/` as runnable or not for the
   selected adapter.
6. Build `.agents/references/state/GLOBS_INDEX.md`: the reverse index from
   glob → rules.
7. Project rules to the agent's native surface (`.cursor/rules/`,
   `.claude/skills/`, or the index-driven path for pointer-style
   adapters).

After every state-file write or new primitive, run `keystone index` so
`.keystone/INDEX.json` reflects the new shape. If skills, subagents,
or commands changed, run `keystone project` to regenerate `.claude/`.

## When to trigger

- Immediately after `keystone init`, before the first real task.
- After a major stack change (new language, framework upgrade,
  monolith → services split) — re-bootstrap to refresh the region map
  and idiom seeds.

## Completion check

Bootstrap is **not** done until:

- `grep -E '<[^>]+>' .agents/references/state/CODEBASE_STATE.md`
  returns no template placeholders.
- `last_reconciled` in `CODEBASE_STATE.md` is today's date.
- `.agents/references/idioms/<stack>/` and `.agents/rules/idioms/<stack>/` exist for
  each detected stack.
- Every seeded idiom rule declares real `paths:`.
- `.agents/references/state/GLOBS_INDEX.md` lists at least one row per seeded
  rule (or is explicitly empty).

See the command for the full check list.
