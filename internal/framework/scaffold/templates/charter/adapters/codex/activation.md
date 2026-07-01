# Codex CLI — Activation

How ambient corpus content gets loaded into Codex, where the menu lives, and how context resets work.

## The menu

Codex reads `AGENTS.md` at the repo root on session start. Keystone drops one there pointing at `charter/README.md`.

Codex actually walks a cascade — `~/.codex/AGENTS.md` (global) first, then every `AGENTS.md` from the git root down to the current working directory, concatenated in that order (closer to cwd = higher precedence, because it appears later in the prompt). At each level, `AGENTS.override.md` is checked before `AGENTS.md` and wins when present. The combined load is capped at `project_doc_max_bytes` (32 KiB default); accumulation stops once the cap is hit. The charter only ships the repo-root file; the cascade is bonus capability the user can layer their own content into without us prescribing.

## Where runtime config lives

- `~/.codex/AGENTS.md` (and `~/.codex/AGENTS.override.md`) — global instructions. User-owned, the installer does not touch.
- `~/.codex/` — other global config (model selection, API keys, etc.). Same: user-owned.
- `AGENTS.md` at repo root — project menu pointing at the corpus (the only file the installer writes).
- Optional `AGENTS.override.md` at any level — local override for the matching `AGENTS.md`. Not used by the charter; available to the user.

There is no `.codex/` directory convention for project-scoped commands, skills, or prompts. Codex has no slash-command or skill system; lifecycle actions are invoked by natural language.

## Ambient loading

`AGENTS.md` is read at session start. The full corpus (`charter/`) is read on demand by Codex's file-read capability inside the conversation — not autoloaded.

- `principles/` is referenced from `AGENTS.md` so the agent knows where to find it.
- `idioms/<stack>/` loads when the **orient** action detects the touched region.
- `domain/` loads at the start of any task.
- `state/` loads on demand from **orient**.
- `process/<phase>.md` loads when the agent enters that phase.

## Context-reset primitive

Codex has no in-session `/clear`. To reset context after a flywheel write (**synthesize** or **audit**), the user exits and starts a new `codex` session. The next session reads the freshly written corpus.

## Lazy-by-region — manual

Codex has no glob-based rules system. The **orient** action implements lazy-by-region manually: the agent reads the touched paths, walks `charter/corpus/state/CODEBASE_STATE.md` to map paths to stacks, consults `charter/corpus/state/GLOBS_INDEX.md` to gate per-guide loading on declared `globs:`, and reads the resulting guides on demand.

## Approval modes ↔ pacing modes

Codex's approval flags (`--ask-for-approval`, `--auto-edit`, `--full-auto`) are the closest analog to the charter's pacing modes. The flag is chosen at session start, not switched mid-session — so to change pacing mid-task, the user updates `charter/guides/process/modes.md` and starts a new Codex session with the matching flag.
