# Adapters

Per-agent bindings. The corpus (principles, idioms, domain, state, process) is agent-agnostic — it speaks in **lifecycle actions** (`orient`, `verify`, `learn`, etc.). An adapter binds those abstract actions to a specific coding agent's invocation surface: slash commands, rules files, CLI subcommands, instructions fields, etc.

## What an adapter ships

Each `adapters/<agent>/` directory contains three files:

| File | Answers |
|---|---|
| `lifecycle.md` | How each abstract action (`orient`, `check-drift`, `verify`, `review`, `learn`, `bootstrap`, `audit`, `synthesize`, `migrate`, `mode`) is invoked. |
| `sensors.md` | How sensors actually fire — autonomous shell execution, surfaced commands for the human to run, or no-op with checklist. |
| `activation.md` | Where ambient corpus content is loaded from (root file convention), how lazy-by-region works, and the agent's context-reset primitive. |

## Supported agents

| Agent | Adapter | Rules surface |
|---|---|---|
| Claude Code | [`claude-code/`](claude-code/) | slash commands + `CLAUDE.md` + `.claude/` |
| Codex CLI | [`codex/`](codex/) | `AGENTS.md` at repo root |
| [pi.dev](https://pi.dev) | [`pi/`](pi/) | `AGENTS.md` + `.pi/prompts/` + `.pi/settings.json` |
| Cursor | [`cursor/`](cursor/) (stub) | `.cursor/rules/*.mdc` with glob frontmatter |
| Aider | [`aider/`](aider/) (stub) | `CONVENTIONS.md` + `.aider.conf.yml` |
| GitHub Copilot CLI | [`github-copilot-cli/`](github-copilot-cli/) (stub) | `.github/copilot-instructions.md` |
| Continue | [`continue/`](continue/) (stub) | `.continuerules` + yaml config |
| Cline / Roo Code | [`cline/`](cline/) (stub) | custom instructions field |
| Goose | [`goose/`](goose/) (stub) | `.goosehints` + config |
| (any other) | [`_generic/`](_generic/) | the floor: agent reads markdown on demand |

## How adapters degrade

Not every agent can do everything Claude Code can. The corpus is written so it still works under reduced capability — adapters document the degradation explicitly. The three common failure modes:

1. **No autonomous sensor execution.** Agent cannot run shell during a turn. Sensors degrade to "agent surfaces the command, human runs it, agent reads the output."
2. **No sub-agent parallelism.** `review` collapses from "spawn N reviewers in parallel" to "sequential reviewer passes" or "single combined reviewer prompt."
3. **No autonomy levels.** Pacing modes (paired / solo / autopilot) collapse to a single mode. The phases still run; the user-facing pace is fixed.

Each adapter calls out which of these apply.

## Adding a new adapter

1. Create `adapters/<agent>/` with the three files above.
2. Fill in the **lifecycle table** — every action in `harness/process/` must have a row, even if the row says "not supported on this agent."
3. Fill in the **sensors table** — every sensor in `harness/process/sensors.md` gets a binding (autonomous / surfaced / no-op).
4. Fill in **activation** — where ambient content loads from, glob/region matching mechanism, context-reset primitive.
5. Add a row to the table above.
6. Add the matching installable artifacts to `targets/<agent>/` at repo root (separate from the corpus — this is what gets dropped into a consumer project).

## The plugin/corpus separation invariant still applies

`harness/adapters/<agent>/` is documentation about how the agent binds to the corpus. The actual artifacts that get installed into a consumer's project live in `targets/<agent>/` at repo root. Do not duplicate content between the two; the adapter doc *describes*, the target *ships*.
