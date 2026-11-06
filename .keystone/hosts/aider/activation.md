# Aider — Activation (rules binding)

How the corpus's ambient content (principles, idioms, process docs) gets loaded into Aider.

## The menu

Aider reads `CONVENTIONS.md` at the repo root on every session — the file is treated as conventions Aider must respect. The installer drops a short `CONVENTIONS.md` that points at `.charter/` and lists the five layers and the iron laws.

`CONVENTIONS.md` is the only thing Aider auto-loads. Everything else under `.charter/` is read on demand via `/add`, `/read-only`, or the agent's file-read tool.

## Where runtime config lives

```
.aider.conf.yml        # optional; see "Suggested config" below
CONVENTIONS.md         # always-loaded menu pointer (installed by keystone)
```

There is no `.aider/` directory by convention. Aider's state lives in `.aider.chat.history.md`, `.aider.input.history`, and `.aider.tags.cache.v3/` (all gitignored by default).

## Suggested `.aider.conf.yml`

```yaml
read:
  - CONVENTIONS.md
  - .charter/README.md
auto-commits: false
test-cmd: <project test command>
lint-cmd: <project lint command>
```

The `read:` list auto-attaches files at session start *without* including them in the editable set. `CONVENTIONS.md` plus `.charter/README.md` is enough to orient Aider; specific phase docs and idiom files are loaded on demand.

`auto-commits: false` is recommended — the charter's commit-message sensor inspects the proposed message before the commit, and Aider's default auto-commit fires before that gate can run.

## Ambient loading

Aider's load surface is intentionally narrow:

- **Auto-loaded at session start:** files in `.aider.conf.yml`'s `read:` list. The charter puts the menu pointer here.
- **Manually attached:** `/add <path>` adds a file to the editable set; `/read-only <path>` adds it as reference. Use these to bring in `.charter/corpus/principles/<file>.md`, `.charter/corpus/idioms/<stack>/*.md`, `.charter/guides/process/<phase>.md` as the work needs.
- **Read on demand:** Aider can read files mentioned in the chat (it has file-read access). Inline reference to `.charter/corpus/state/CODEBASE_STATE.md` will cause it to read that file.

The pattern this charter recommends: keep the read-set small (menu + .charter/README.md) and use the lifecycle actions to instruct Aider to read each phase doc and the matching idioms when they are needed.

## Lazy-by-region — not native

Aider has no glob-based auto-attachment. Region-scoped idiom loading is done by the **orient** action, which:

1. Looks at the files the user is editing — the **touched-files set**.
2. Cross-references `.charter/corpus/state/CODEBASE_STATE.md` to find the stack and the idiom folder.
3. Reads `.charter/corpus/state/GLOBS_INDEX.md` and keeps only the guides whose `globs:` match at least one touched file (guides without `globs:` are still loaded by stack).
4. Reads `.charter/corpus/idioms/<stack>/*.md` for the resulting set and brings those rules into the conversation.

The user can streamline this by adding the stack-relevant idiom files to `.aider.conf.yml`'s `read:` list for sessions that touch one stack heavily.

## Context-reset primitive

- **`/clear`** — clears chat history. Use after **synthesize** or **audit** so the next turn re-reads the updated corpus.
- **`/drop`** — removes a specific file from context. Useful when a file you `/add`'d is no longer relevant.
- **`/reset`** — `/clear` plus `/drop` of all files. The hard reset.

After a flywheel write that changed any file in Aider's read-set, run `/clear` and let Aider re-read `CONVENTIONS.md` on the next prompt.

## Domain / state / process loading

- `domain/` — relevant at task start; the orient action's first read is the domain layer.
- `state/CODEBASE_STATE.md` — loaded by the **orient** action.
- `process/<phase>.md` — loaded by the lifecycle action when the agent enters that phase. The user can pre-attach with `/read-only .charter/guides/process/<phase>.md`.

## Capability gaps

- **No autonomous tracker integration.** Paste the card content or use the shell to fetch (`gh issue view`, `curl`, etc.) via `/run`.
- **No sub-agent parallelism.** The **review** action runs each concern sequentially.
- **Single autonomy mode.** Aider applies edits as soon as it suggests them; `/undo` is the only retreat. Treat all sessions as `paired`.
- **`auto-commits: true`** (the default) bypasses the commit-message sensor. The charter requires it off.
