# Continue — Activation (rules binding)

How the charter's ambient content (guides and process docs) loads into Continue, and how the agent reaches corpus files on demand.

## The menu

Continue auto-loads `.continuerules` at the repo root on every session. The installer drops a short `.continuerules` that points at `charter/`, lists the five components, and states the iron laws.

`.continuerules` is the legacy mechanism. Newer Continue versions also read `config.yaml` (or `config.json`) at `<repo>/.continue/config.yaml` for slash commands, context providers, and tool/MCP configuration. The charter ships `.continuerules` because every Continue version reads it; richer config is opt-in via `config.yaml`.

## Where runtime config lives

```
.continuerules                 # always-loaded menu pointer (installed by keystone)
.continue/config.yaml          # slash commands, context providers, MCP servers (optional)
.continue/config.json          # legacy alternative to config.yaml
```

There is also a global config at `~/.continue/config.yaml`. The project file overrides the global file for repo-specific bindings.

## Suggested `config.yaml`

```yaml
name: keystone-charter
version: 0.3.0
schema: v1

# Slash commands (full list in lifecycle.md).
prompts:
  - name: spec
    description: "Spec phase"
    prompt: "Read charter/guides/process/spec.md and run the spec action for: {{{ input }}}"
  # ... (see lifecycle.md for the full set)

# Always-on context.
context:
  - provider: file
  - provider: codebase
  - provider: diff
  - provider: terminal
  - provider: folder
    params:
      folder: charter/guides

# Optional: MCP servers for tracker integration.
mcpServers:
  - name: atlassian
    command: npx
    args: ["-y", "@modelcontextprotocol/server-atlassian"]
```

The `folder` provider on `charter/guides/` keeps rule content in reach without forcing the user to `@`-mention each file.

## Ambient loading

Continue's ambient surface is layered:

- **Auto-loaded every session:** `.continuerules` (the menu).
- **Auto-attached via context providers:** anything listed in `config.yaml`'s `context:` — in the charter's recommended config, that includes `charter/guides/` (via `folder` provider) and the current diff (`diff` provider, on demand).
- **Reached on `@`-mention or model decision:** any file under `charter/corpus/` is read when the agent follows a forward-link from a guide or when the user `@`-mentions it.

Guides are ambient; corpus is on-demand. This maps cleanly onto Continue's provider model — keep guides under a `folder` provider, leave corpus reachable via `file` / `codebase`.

## Lazy-by-region — partial

Continue has no glob-based auto-attachment, but two providers approximate it:

- **`codebase` provider** — semantic search the model invokes on demand. When the **orient** action lands in a region, it asks "find related idioms" and the model walks `charter/corpus/idioms/<stack>/` results into context. Orient also reads `charter/corpus/state/GLOBS_INDEX.md` and narrows the candidate set to guides whose `globs:` match at least one touched file.
- **`folder` provider scoped to a single idiom** — e.g., `folder: charter/guides/idioms/<stack>` for sessions that touch one stack heavily.

The general pattern: keep the `folder` provider scoped narrowly to `charter/guides/` so always-loaded content stays small, and let the model reach for corpus and stack-specific idioms via `codebase` search.

## Context-reset primitive

- **New chat** — the "+" button in the side panel. Drops conversation history; reloads `.continuerules` and configured context providers.
- **Clear chat** — from the chat options menu; same effect for the current panel.

After a flywheel write that touched any `charter/guides/` file, start a new chat so the next turn re-reads the updated guides.

## Domain / state / process loading

- **`corpus/domain/`** — relevant at task start; the orient action's first read after state.
- **`corpus/state/CODEBASE_STATE.md`** — loaded by the **orient** action (via `@file charter/corpus/state/CODEBASE_STATE.md` or model's file-read).
- **`actions/<action>.md`** — loaded when the user invokes that action by name.
- **`guides/process/<phase>.md`** — loaded when the agent enters that phase from inside an action playbook.
- **`guides/<layer>/<name>.md`** — auto-loaded via the `folder` context provider on `charter/guides/`.

## Capability gaps

- **No autonomous tracker integration without MCP.** If the user doesn't configure an Atlassian / Linear MCP server, tracker fetch falls back to `cmd` steps (`gh issue view`, `curl`).
- **No sub-agent parallelism.** The **review** action runs each concern sequentially.
- **Single autonomy mode.** Agent mode applies edits as it suggests them; the user reviews each edit before it lands. Treat all sessions as `paired`.
- **`.continuerules` only.** Continue versions that predate `config.yaml` cannot host MCP servers — they fall back to typing the action name in chat against the agent-agnostic playbooks in `charter/actions/`.
