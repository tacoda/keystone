# Claude Code — Activation

How ambient corpus content gets loaded into Claude Code, where the menu lives, and how context resets work.

## The menu

Claude Code reads `CLAUDE.md` at the project root on every session start. That file is the menu — it points the agent at `harness/` and tells it the corpus is the source of truth.

The installer drops a `CLAUDE.md` at root with this minimal content:

```markdown
# CLAUDE.md

This project uses a project harness. Read [harness/README.md](harness/README.md) before starting work.

The corpus there defines:
- Principles (universal engineering rules)
- Idioms (stack-specific patterns)
- Domain (business invariants for this project)
- State (the empirical map of the codebase)
- Process (six phases: spec → planning → implementation → verification → review → release)

Adapter bindings for Claude Code live at `harness/adapters/claude-code/`.
```

If the project already had a `CLAUDE.md`, the installer asks before overwriting. The user can also choose to merge — the shipped `CLAUDE.md` content is short and copy-paste-friendly.

## Where runtime config lives

- `.claude/settings.json` — permissions, MCP servers, hooks (optional, user-owned).
- `.claude/agents/*.md` — project-local sub-agents (review-functional, review-security, etc.). Optional; only needed if the user wants to customize review sub-agents beyond what `harness/actions/review.md` describes.

The installer does **not** create or populate `.claude/`. Lifecycle actions live in `harness/actions/<action>.md` and are invoked via natural language — no slash commands, no skills, no policy install required.

## Ambient loading

Claude Code automatically loads `CLAUDE.md` from the repo root *and* from every parent directory up to `~/.claude/CLAUDE.md`. The harness places one short pointer at root; the corpus and action playbooks are read via the Read tool on demand, not autoloaded.

This is intentional — loading the entire corpus at every session start would burn context. Instead:

- `policies/universal/` is referenced from `CLAUDE.md` so it's always available by path.
- `idioms/<stack>/` loads lazily when the **orient** action detects the touched region.
- `domain/` loads at the start of any task (always relevant).
- `state/` loads on demand from the **orient** action.
- `actions/<action>.md` loads when the user invokes that action by name.
- `guides/process/<phase>.md` loads when the agent enters that phase from inside an action playbook.

## Context-reset primitive

Claude Code provides two reset paths:

- `/clear` — wipes the session entirely. Use after **synthesize** or **audit** writes changed the corpus.
- `/compact` — summarizes the session and continues with a smaller context. Use when the session has been productive but is getting long.

The corpus references these as "the agent's context-clear primitive." For Claude Code, `/clear` is the right one after corpus mutation.

## Lazy-by-region

Claude Code does not have a built-in "load this file when editing this path" mechanism (unlike Cursor's `.cursor/rules/*.mdc` globs). The harness implements lazy-by-region inside the **orient** action playbook: the playbook tells the agent to read the touched paths, walk `harness/corpus/state/CODEBASE_STATE.md` to map paths to stacks, consult `harness/corpus/state/GLOBS_INDEX.md` to gate per-guide loading on declared `globs:`, and load the resulting guides via Read.
