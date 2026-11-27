<!-- keystone:start -->
# Charter

This repo is governed by a **keystone charter**: the codified context of the repository — everything a coding agent should be able to look up about how this codebase works and how to work in it.

The sources live in `.agents/`; keystone's index and state live in `.keystone/`. Host files (`.claude/`, `.cursor/`, the region of this file between the keystone markers) are generated from `.agents/`: edit the sources, never the generated files.

## Read first

`.keystone/INDEX.lite.json` lists every primitive (kind, id, description). Browse it, then open a body only when it applies to the task.

## Query the charter

Nothing loads eagerly except this block and the index; open bodies on demand.

- CLI — `keystone charter query <files…> [--task "…"]` (add `--full` for bodies); also `list` / `show <kind> <id>` / `search <query>`
- MCP (`keystone mcp serve`) — `keystone_charter_query` / `keystone_list_primitives` / `keystone_get_primitive` / `keystone_search`
- Charter hygiene — `keystone learn | synthesize | audit | drift | mode` (the same verbs as `keystone_*` MCP tools); dashboard: `keystone web serve`

## On each host

- **Claude Code** — loads rules, hooks (inferential), skills, commands, agents by itself.
- **Cursor** — loads rules, skills, commands, agents by itself.
- **GitHub Copilot** — loads rules, skills, commands, agents by itself.
- **Continue** — loads rules, skills, commands by itself; open agents from `.agents/` yourself; does **not** run hooks — run the checks in `.agents/hooks/` yourself before claiming done.
- **opencode** — loads skills, commands, agents by itself; open rules from `.agents/` yourself.
- **Codex** — loads skills, agents by itself; open rules, commands from `.agents/` yourself.
- **Aider** — open rules, skills, commands, agents from `.agents/` yourself; does **not** run hooks — run the checks in `.agents/hooks/` yourself before claiming done.
<!-- keystone:end -->
