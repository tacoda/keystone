# pi.dev — Activation

How ambient corpus content gets loaded into pi, where the menu lives, and how context resets work.

## The menu

Pi loads `AGENTS.md` at session start from three places, in precedence order:

1. **Global** — `~/.pi/agent/AGENTS.md`
2. **Project hierarchy** — walks parent directories up from CWD
3. **Local** — the current directory's `AGENTS.md`

Keystone drops an `AGENTS.md` at the consumer's repo root. Pi finds it via path-walk and loads it on every session.

If the project also needs a custom system prompt (replacing pi's default), use `.pi/SYSTEM.md`. If you only want to append to the default, use `.pi/APPEND_SYSTEM.md`. Keystone does **not** override the system prompt by default — the corpus is loaded via `AGENTS.md`, not `SYSTEM.md`.

## Where runtime config lives

- `.pi/settings.json` — project-scoped JSON settings (model selection, telemetry, compaction, etc.). User-owned.
- `.pi/SYSTEM.md` or `.pi/APPEND_SYSTEM.md` — optional system-prompt overrides. User-owned.

The installer does **not** create `.pi/`. Lifecycle actions are agent-agnostic playbooks at `charter/actions/<action>.md`; the agent reads them on demand when the user asks to run an action by name.

## Ambient loading

Pi's `AGENTS.md` precedence chain means the corpus pointer is always available regardless of where in the repo the user runs pi. The full corpus (`charter/`) is read on demand via pi's file-read capability — not autoloaded.

This is intentional: loading the entire corpus at every session start would burn context.

- `policies/universal/` is referenced from `AGENTS.md` so it's always available by path.
- `idioms/<stack>/` loads when the **orient** action detects the touched region.
- `domain/` loads at the start of any task.
- `state/` loads on demand from **orient**.
- `actions/<action>.md` loads when the user invokes that action by name.
- `guides/process/<phase>.md` loads when the agent enters that phase from inside an action playbook.

## Context-reset primitive

Pi supports two reset paths:

- `/compact` — summarizes the session and continues with a smaller context. Use after **synthesize** or **audit** wrote corpus changes, so the next turn can re-read the updated corpus.
- **Session branching** — pi's tree-structured session history lets you branch off rather than resetting. Useful when an exploration was worth keeping but a new direction is needed.

For pi, `/compact` is the right reset after corpus mutation, unless you want to keep the prior path alive on a branch.

## Lazy-by-region — manual

Pi has no glob-based rules system (unlike Cursor's `.cursor/rules/*.mdc`). The **orient** action playbook implements lazy-by-region manually: the agent reads the touched paths, walks `charter/corpus/state/CODEBASE_STATE.md` to map paths to stacks, consults `charter/corpus/state/GLOBS_INDEX.md` to gate per-guide loading on declared `globs:`, and reads the resulting guides with pi's file-read tool.
