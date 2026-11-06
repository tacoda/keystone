---
name: keystone-allocate
description: 'Move an existing agent setup into the charter and cut always-loaded context: run keystone allocate, review its classifications, then apply. Triggers: keystone allocate; keystone:allocate; /keystone:allocate; convert my .claude setup; migrate .cursorrules; trim CLAUDE.md; reduce context bloat.'
---

# keystone:allocate — move an existing setup into the charter

`keystone allocate` reports where each piece of an existing agent setup
belongs in `.agents/`: host files (`.claude/`, `.cursor/`, Copilot, Cline,
Continue, opencode) and the sections of the always-loaded root files
(`CLAUDE.md`, `AGENTS.md`, `GEMINI.md`, `.goosehints`). The root-file
classification is a heuristic. Your job is to check it.

## Run

1. Run `keystone allocate` and read the report. It writes nothing.
2. Check every root-file section it proposes to move:
   - A **rule** must apply only to the listed `paths:`. If it applies
     everywhere, it stays in the root file.
   - A **command** is a procedure the user asks for by name. If the agent
     should start it unasked, make it a skill instead.
   - A **hook** runs its command on `Stop` and blocks on failure. Confirm
     the command and the event.
   - A **reference** is reasoning, opened on demand. A directive the agent
     must always follow stays in the root file.
3. Check every item note: a subagent with no `tools:` needs a list, and a
   Claude Code hook with a matcher stays in `settings.json`.
4. Present the proposed moves to the user, with the tokens each one saves.
5. When the user agrees, run `keystone allocate --apply`. It writes new
   files under `.agents/` only and changes no original.
6. Fix each converted file you flagged, then run `keystone lint`.
7. Run `keystone allocate --overwrite` only when the user explicitly asks
   for it. It removes moved sections from `CLAUDE.md` and `AGENTS.md`,
   removes moved hooks from `settings.json`, and lets keystone replace
   the converted host files.

Never run `--overwrite` on your own judgment: it changes the user's files.
