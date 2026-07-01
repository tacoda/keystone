# Charter

You **must** read [`CHARTER.md`](CHARTER.md) now, before doing anything in this repo. It carries the iron laws and the ambient rules that govern the charter; they apply whether or not this file restates them. Do not proceed without loading it.

## On this host — opencode

- **Subagents** — spawn charter agents (`.charter/agents/`) as subagents for review/scout work.
- **Slash commands** — charter commands and playbooks surface as `/keystone-<id>`.
- **Skills** — auto-activate by their `triggers:`.
- **No automatic hooks** — run the checks a hook would fire (lint, type-check, test, build) yourself before claiming done.
