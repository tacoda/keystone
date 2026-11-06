# Goose — Activation (rules binding)

How the charter's ambient content (guides and process docs) loads into Goose, and how the agent reaches corpus files on demand.

## The menu

Goose reads `.goosehints` at the repo root on every session — the file is treated as project-specific hints prepended to the system prompt. The installer drops a short `.goosehints` that points at `.charter/`, lists the five components, and states the iron laws.

Goose also reads a **global** hints file at `~/.config/goose/hints.md`. The charter uses only the workspace `.goosehints`; the global file is the user's territory.

## Where runtime config lives

```
.goosehints                       # workspace hints (auto-loaded; installed by keystone)
~/.config/goose/config.yaml       # global Goose config (extensions, provider, mode)
~/.config/goose/hints.md          # global hints (user-owned; charter does not write)
.goose/recipes/<name>.yaml        # per-action recipes (installed by keystone)
```

There is no project-scoped `config.yaml` for Goose equivalent to Continue's — Goose's main config is global. Recipes are the per-project mechanism.

## Required extensions

Goose's tool use comes from extensions. The charter assumes:

- **`developer`** (built-in) — shell execution, file read/write. Required.
- **`memory`** (optional) — long-term recall across sessions; useful for the **learn** flow.
- **`github`** (optional, MCP) — tracker fetch for GitHub Issues and PR operations.
- **`atlassian`** or **`linear`** (optional, MCP) — tracker fetch for the corresponding system.

Enable an extension globally:

```bash
goose configure --enable-extension <name>
```

Or per-session via the recipe's `extensions:` field.

Without the developer extension, sensors degrade to "agent surfaces commands; user runs them." That mode works but is not the charter's recommended shape.

## Ambient loading

Goose's ambient surface:

- **Auto-loaded every session:** `.goosehints` (workspace) and `~/.config/goose/hints.md` (global). Both prepend to the system prompt.
- **Auto-loaded via recipes:** anything listed in a recipe's `instructions:` field becomes part of the session prompt. The charter's recipes inline pointers to specific phase docs.
- **Reached on model decision:** the developer extension's `text_editor` tool reads files; the model uses it to pull in guides, corpus, and state on demand.

Guides should live in (or be reachable from) `.goosehints` so the model is always under them. Corpus files are reached via `text_editor` when a forward-link is followed.

## Lazy-by-region — not native

Goose has no glob-based auto-attach. Region-scoped idiom loading happens inside the **orient** action, which:

1. Reads `.charter/corpus/state/CODEBASE_STATE.md` to map touched paths to a stack.
2. Reads `.charter/corpus/state/GLOBS_INDEX.md` and gates per-guide loading on declared `globs:` (guides without `globs:` keep stack-based loading).
3. Reads `.charter/guides/idioms/<stack>/*.md` (rules) for the resulting set and, on demand, `.charter/corpus/idioms/<stack>/*.md` (reasoning).
4. Carries those rules forward in the session.

For sessions that touch one stack heavily, the user can add `.charter/guides/idioms/<stack>/` to `.goosehints` so the rules load automatically.

## Recipes vs. interactive sessions

Two invocation modes:

- **`goose session`** — interactive REPL. Hints auto-load; tools used on demand. Best for development work where the user pairs with the agent.
- **`goose run --recipe <path>`** — non-interactive. Recipe defines the prompt, extensions, and parameters; Goose runs to completion. Best for the lifecycle actions: `keystone-verify`, `keystone-review`, `keystone-audit`.

The charter ships one recipe per lifecycle action. Users can also chain recipes: a `keystone-cycle` recipe that runs `spec → orient → implementation → verify → review` end to end.

## MCP integration

Goose's extension system speaks MCP natively. To add a tracker integration:

```bash
goose configure --add-extension
```

…and follow the prompts for the MCP server (Atlassian, Linear, GitHub, etc.). The charter assumes presence when documenting tracker flows; without the extension, the **tracker-card-fetcher** sensor falls back to shell (`gh issue view`) or user paste.

## Context-reset primitive

- **`/exit`** in an interactive session, then `goose session` to start fresh. Drops history; reloads hints.
- **New session** — every `goose run` invocation is a fresh session by default; recipes implicitly reset.
- **`/clear`** — clears the current session's history without exiting (some Goose versions).

After a flywheel write that touched any `.charter/guides/` file, start a new session so the next turn re-reads the updated guides via `.goosehints` and any recipe instructions.

## Domain / state / process loading

- **`corpus/domain/`** — relevant at task start; the orient action reads it after state.
- **`corpus/state/CODEBASE_STATE.md`** — read by the **orient** action via the `text_editor` tool.
- **`guides/process/<phase>.md`** — read when entering each phase; recipes specify this in `instructions:`.
- **`guides/<layer>/<name>.md`** — reachable via the menu pointer in `.goosehints`; specific rule files read on demand.

## Capability gaps

- **No slash-command primitive.** Goose has only natural-language invocation in interactive mode, plus recipes for non-interactive runs.
- **No sub-agent parallelism.** Review runs sequentially.
- **No autonomy levels.** Interactive vs. recipe is the only axis; the charter's `paired`/`solo`/`autopilot` collapse approximately.
- **No native lazy-by-region.** Idiom loading is explicit, done inside the orient action.
- **Provider lock to the global config.** Goose's model provider is global, not per-project. A team standardizing on a model needs documentation outside the charter.
