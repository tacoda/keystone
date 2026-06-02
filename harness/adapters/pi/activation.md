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

- `.pi/settings.json` — project-scoped JSON settings (model selection, telemetry, compaction, etc.).
- `.pi/prompts/*.md` — project-scoped slash-command templates (one file per lifecycle action).
- `.pi/skills/<name>/SKILL.md` — project-scoped skills (none ship by default; teams can add their own).
- `.pi/SYSTEM.md` or `.pi/APPEND_SYSTEM.md` — optional system-prompt overrides.

The installer creates `.pi/prompts/` populated with one template per lifecycle action. It does **not** write `.pi/settings.json` or `.pi/SYSTEM.md` — those stay user-owned.

## Ambient loading

Pi's `AGENTS.md` precedence chain means the corpus pointer is always available regardless of where in the repo the user runs pi. The full corpus (`harness/`) is read on demand via pi's file-read capability inside prompt templates — not autoloaded.

This is intentional: loading the entire corpus at every session start would burn context.

- `principles/` is referenced from `AGENTS.md` so it's always available by path.
- `idioms/<stack>/` loads when the **orient** template detects the touched region.
- `domain/` loads at the start of any task.
- `state/` loads on demand from **orient**.
- `process/<phase>.md` loads when the agent enters that phase.

## Context-reset primitive

Pi supports two reset paths:

- `/compact` — summarizes the session and continues with a smaller context. Use after **synthesize** or **audit** wrote corpus changes, so the next turn can re-read the updated corpus.
- **Session branching** — pi's tree-structured session history lets you branch off rather than resetting. Useful when an exploration was worth keeping but a new direction is needed.

The corpus references "the agent's context-clear primitive." For pi, `/compact` is the right one after corpus mutation, unless you want to keep the prior path alive on a branch.

## Lazy-by-region — manual

Pi has no glob-based rules system (unlike Cursor's `.cursor/rules/*.mdc`). The **orient** prompt template implements lazy-by-region manually: it reads the touched paths, walks `harness/corpus/state/CODEBASE_STATE.md` for matching idioms, and reads them with pi's file-read tool.

## Skills as an alternative

Lifecycle actions ship as prompt templates because they're lighter and don't require directory structure per action. If you want richer behavior — embedded scripts, helper assets — promote an action to a skill at `.pi/skills/<action>/SKILL.md`. The invocation changes from `/<action>` to `/skill:<action>`, but the corpus path it reads stays the same.
