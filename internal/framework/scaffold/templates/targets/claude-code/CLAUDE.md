@CHARTER.md

You **must** read [`CHARTER.md`](CHARTER.md) before doing anything in this repo — it carries the iron laws and the ambient rules that govern the charter. The import above loads it; do not proceed without it.

## On this host — Claude Code

- **Subagents** — spawn charter agents (`.charter/agents/`) as subagents via the Task tool for review/scout work.
- **Slash commands** — charter commands and playbooks surface as `/keystone-<id>`.
- **Skills** — auto-activate by their `triggers:`.
- **Hooks** — charter hooks fire automatically on Claude Code lifecycle events.
