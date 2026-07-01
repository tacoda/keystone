---
name: keystone-bootstrap
description: One-time initial charter scaffold — detect stack, seed state, scaffold idioms, classify sensors, build the globs index.
tools:
  - Read
  - Grep
  - Glob
  - Bash
  - Write
  - Edit
  - Task
model: opus
---

# keystone:bootstrap — make the charter specific to your codebase

The first thing to run after `keystone init`. Walks the repo, detects
the real stack, and replaces every template placeholder under
`.charter/corpus/state/` with code-grounded values.

The canonical playbook is the **bootstrap action** at
`.charter/actions/bootstrap.md`. This skill is a thin
trigger that points the agent at it — open the action and execute its
activities in order.

## Run

Open `.charter/actions/bootstrap.md` and execute every activity
listed under `## Activities`. Every activity must land a real file
change before moving on; narration does not count.

A short summary of the work, copied from the action:

1. Detect the stack (`package.json`, `go.mod`, `pyproject.toml`, …).
2. Seed `corpus/state/CODEBASE_STATE.md` with real tool commands +
   region map.
3. Scaffold `corpus/idioms/<stack>/` + paired `guides/idioms/<stack>/`
   with `globs:` derived from the region map.
4. Inventory `guides/computational/` from `.editorconfig`, LSP config,
   formatter configs.
5. Classify every sensor under `sensors/` as runnable or not for the
   selected adapter.
6. Build `corpus/state/GLOBS_INDEX.md` — the reverse index from
   glob → guides.
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

- `grep -E '<[^>]+>' .charter/corpus/state/CODEBASE_STATE.md`
  returns no template placeholders.
- `last_reconciled` in `CODEBASE_STATE.md` is today's date.
- `corpus/idioms/<stack>/` and `guides/idioms/<stack>/` exist for
  each detected stack.
- Every seeded idiom guide declares real `globs:`.
- `corpus/state/GLOBS_INDEX.md` lists at least one row per seeded
  guide (or is explicitly empty).

See the action for the full check list.
