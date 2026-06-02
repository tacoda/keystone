# Codex CLI — Activation

How ambient corpus content gets loaded into Codex, where the menu lives, and how context resets work.

## The menu

Codex reads `AGENTS.md` at the repo root on session start. Keystone drops one there pointing at `harness/README.md`.

There is no convention for parent-directory walking or global `AGENTS.md` like pi has — Codex looks at the CWD's `AGENTS.md`.

## Where runtime config lives

- `~/.codex/` — global config (model selection, API keys, etc.). User-owned, the installer does not touch.
- `AGENTS.md` at repo root — project menu pointing at the corpus.

There is no `.codex/` directory convention for project-scoped commands, skills, or prompts. Codex has no slash-command or skill system; lifecycle actions are invoked by natural language.

## Ambient loading

`AGENTS.md` is read at session start. The full corpus (`harness/`) is read on demand by Codex's file-read capability inside the conversation — not autoloaded.

- `principles/` is referenced from `AGENTS.md` so the agent knows where to find it.
- `idioms/<stack>/` loads when the **orient** action detects the touched region.
- `domain/` loads at the start of any task.
- `state/` loads on demand from **orient**.
- `process/<phase>.md` loads when the agent enters that phase.

## Context-reset primitive

Codex has no in-session `/clear`. To reset context after a flywheel write (**synthesize** or **audit**), the user exits and starts a new `codex` session. The next session reads the freshly written corpus.

## Lazy-by-region — manual

Codex has no glob-based rules system. The **orient** action implements lazy-by-region manually: the agent reads the touched paths, walks `harness/corpus/state/CODEBASE_STATE.md` for matching idioms, and reads them on demand.

## Approval modes ↔ pacing modes

Codex's approval flags (`--ask-for-approval`, `--auto-edit`, `--full-auto`) are the closest analog to the harness's pacing modes. The flag is chosen at session start, not switched mid-session — so to change pacing mid-task, the user updates `harness/guides/process/modes.md` and starts a new Codex session with the matching flag.
