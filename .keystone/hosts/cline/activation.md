# Cline / Roo Code — Activation (rules binding)

How the charter's ambient content (guides and process docs) loads into Cline, and how the agent reaches corpus files on demand.

## The menu

Cline reads two sources at session start:

1. **Custom instructions field** (extension settings: `cline.customInstructions` for Cline, `rooCode.customInstructions` for Roo Code). Always loaded into the system prompt. This is the canonical menu location.
2. **`.clinerules`** (Cline) or **`.roorules`** (Roo Code) at the workspace root. Auto-loaded by newer extension versions. Older versions ignore it.

The installer ships `cline-instructions.md` — the **paste-into-settings text**. The user copies it into the extension's custom-instructions field once. The same content can be dropped at `.clinerules` to cover newer extension versions.

## Where runtime config lives

```
.clinerules            # auto-loaded by newer Cline versions (workspace scope)
.roorules              # same, for Roo Code
                       # (the charter ships cline-instructions.md as paste source)
```

There is no project-scoped `config.json` for Cline equivalent to Continue's `config.yaml`. Custom workflows live in VS Code's extension storage; MCP server config lives in the extension's MCP settings panel (per-workspace and per-global).

## Suggested workflows

Cline's **workflows** feature stores reusable prompts triggered from the side panel. The charter recommends one workflow per lifecycle action — see [`lifecycle.md`](lifecycle.md) for the full list. Users add these via the workflows menu in the side panel; the installer cannot create them programmatically.

## Ambient loading

Cline's ambient surface:

- **Auto-loaded every session:** the custom-instructions field (extension settings) and `.clinerules` (workspace, if newer Cline). The charter uses both — the settings field for safety, `.clinerules` for users on supported versions.
- **Auto-loaded into context on every task:** the **open editor tabs** and the **active file** are surfaced to the model implicitly. Cline does not auto-attach files by path glob.
- **Reached on model decision:** the model uses the `read_file` tool to pull specific files. Forward-links from guides to corpus are followed this way.

Guides should land in the menu (custom-instructions field) so the model is always under them. Corpus stays on disk and is reached via `read_file` when a forward-link is followed.

## Lazy-by-region — not native

Cline has no glob-based auto-attachment. Region-scoped idiom loading happens inside the **orient** action, which:

1. Reads `.charter/corpus/state/CODEBASE_STATE.md` to map the touched paths to a stack.
2. Reads `.charter/corpus/state/GLOBS_INDEX.md` and gates per-guide loading on declared `globs:` (guides without `globs:` keep stack-based loading).
3. Reads `.charter/guides/idioms/<stack>/*.md` and `.charter/corpus/idioms/<stack>/*.md` for the resulting set as needed.
4. Carries those rules forward in the task's context.

The user can streamline this by attaching the relevant idiom files to the task via the "+ Add Files" button in the side panel, which forces them into context.

## MCP integration

Cline has first-class MCP support. The charter recommends configuring at minimum:

- **Tracker MCP** — Atlassian for Jira, Linear's official server, or GitHub for Issues. Used by the **tracker-card-fetcher** sensor inside the **spec** action.
- **GitHub MCP** — for opening PRs and reading review threads during the **release** and **review** actions.

MCP servers are configured in the extension's MCP settings panel (the gear icon). The charter does not install MCP configuration — the user enables and configures servers themselves; the charter assumes presence when documenting tracker / GitHub flows.

## Context-reset primitive

- **New Task** — the "+" button in the side panel. Drops the entire task history; reloads the custom-instructions field. Equivalent to Claude Code's `/clear`.
- **Subtask** — focuses a child task with a tight scope; on completion, control returns to the parent with a summary. Not a context reset for the parent.

After a flywheel write that touched any `.charter/guides/` file, start a new task so the next turn re-reads the updated guides via the custom-instructions field.

## Domain / state / process loading

- **`corpus/domain/`** — relevant at task start; the orient action reads it after state.
- **`corpus/state/CODEBASE_STATE.md`** — read by the **orient** action via `read_file`.
- **`guides/process/<phase>.md`** — read when entering each phase. The Keystone workflows make this explicit ("read .charter/guides/process/<phase>.md").
- **`guides/<layer>/<name>.md`** — should be reachable via the custom-instructions menu pointer. The menu file lists the five components and the iron laws; specific rule files are read on demand.

## Auto-approve and the verify cycle

Cline's per-command approval can stall sensor execution. The charter recommends auto-approving at minimum:

- File reads (model needs to read corpus and guides freely).
- Read-only git commands (`git diff`, `git log`, `git status`).
- Project-scoped test/lint/build commands (whitelist via Cline's settings).

Without these, the **verify** action becomes a clicker game. With them, sensors run in one chat turn.

## Capability gaps

- **No slash-command primitive.** Cline's workflows are the substitute; they require manual setup.
- **No sub-agent parallelism.** Review runs sequentially. Subtasks help with scope but not parallelism.
- **Roo Code drift.** Roo Code forks Cline and diverges over time. The charter adapter applies to both, but specific feature names (workflows, custom-instructions field key) may differ — check the Roo Code docs if a binding does not work.
- **Custom-instructions field is global per workspace, not per-task.** Switching projects in the same VS Code window requires re-pasting the menu (per-workspace settings help).
