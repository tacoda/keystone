# GitHub Copilot — Activation (rules binding)

How the corpus's ambient content (principles, idioms, process docs) gets loaded into GitHub Copilot.

## The menu

Copilot reads `.github/copilot-instructions.md` on every session (both VS Code and CLI). The installer drops this file with a short body pointing at `charter/` and listing the five layers, the lifecycle actions, and the iron laws.

A single `.github/copilot-instructions.md` covers both Copilot in VS Code and the Copilot CLI — Microsoft / GitHub designed them to share the same convention.

## Where runtime config lives

```
.github/
└── copilot-instructions.md     # always-loaded menu pointer (installed by keystone)
```

There is no `.copilot/` directory by convention; Copilot reads project context from `.github/copilot-instructions.md` and from open editor buffers (in VS Code).

## Ambient loading

Copilot loads:

- **`.github/copilot-instructions.md`** at session start — auto, no configuration needed.
- **Currently-open editor buffers** (VS Code only) — visible context for inline suggestions.
- **Files referenced in the chat** — read on demand when the agent or user names them.

The charter places the short menu in `.github/copilot-instructions.md` and lets the lifecycle actions pull in `charter/corpus/principles/<file>.md`, `charter/corpus/idioms/<stack>/*.md`, and `charter/guides/process/<phase>.md` as needed. Loading the entire corpus into the always-loaded menu would dilute every chat.

## Lazy-by-region — not native

Copilot has no glob-based file auto-attachment (unlike Cursor's `.mdc` rules). The **orient** action implements lazy-by-region inside the chat: the agent reads `charter/corpus/state/CODEBASE_STATE.md` to find the stack for the touched paths, consults `charter/corpus/state/GLOBS_INDEX.md` to keep only the guides whose `globs:` match at least one touched file (guides without `globs:` are still loaded by stack), then reads `charter/corpus/idioms/<stack>/*.md` for the resulting set.

For VS Code specifically, the open-file context partially compensates — if the user has the relevant idiom files open in the editor, Copilot uses them as context. But this is incidental, not architectural.

## Context-reset primitive

- **VS Code:** "New chat" button in the Copilot chat panel (also `Cmd+Shift+I` to open a new chat session).
- **CLI:** Start a new `copilot` session.

After any **synthesize** or **audit** writes changed the corpus, reset the context so the next turn re-reads `.github/copilot-instructions.md` and any updated phase docs.

## Domain / state / process loading

- `domain/` — read at the start of any task via "read charter/corpus/domain/" or similar in the chat.
- `state/CODEBASE_STATE.md` — read by the **orient** action.
- `process/<phase>.md` — read by the matching lifecycle action.

The user can pre-warm a session by asking Copilot to "read charter/README.md, charter/corpus/domain/, and charter/corpus/state/CODEBASE_STATE.md" at the start. This brings the always-relevant context into the chat without ceremony.

## VS Code extension considerations

- **Workspace trust** — Copilot respects VS Code's workspace trust model. Untrusted workspaces have reduced functionality; the charter's shell-based sensors will refuse to run.
- **Extension settings** — `github.copilot.chat.codeGeneration.useInstructionFiles: true` is the default in current versions and is required for `.github/copilot-instructions.md` to be honored. Older versions or restrictive policies may have it off; verify in the team's enterprise policy if relevant.
- **Custom prompts** — `.github/prompts/*.prompt.md` (a newer feature) can store reusable prompts; the lifecycle actions are good candidates if your version of Copilot supports them. Not all environments do; the charter does not rely on this feature.

## Capability gaps

- **No tracker integration outside GitHub Issues.** Jira / Linear / Asana cards must be pasted into the chat or fetched via `gh` workarounds.
- **No sub-agent parallelism.** The **review** action runs each concern sequentially.
- **Single autonomy mode** with per-command approval. No `autopilot`.
- **No in-place context compression.** "New chat" is the only reset.
