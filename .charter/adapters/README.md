# Adapters

Per-agent bindings. The charter (corpus, guides, sensors, flywheels) is agent-agnostic — it speaks in **lifecycle actions** (`orient`, `verify`, `learn`, etc.). An adapter binds those abstract actions to a specific coding agent's invocation surface: slash commands, rules files, CLI subcommands, instructions fields, etc.

It also describes how the adapter projects **guides** into the agent's rules surface (so they load ambient) and how the agent reaches **corpus** files on demand.

## What an adapter ships

Each `adapters/<agent>/` directory contains three files:

| File | Answers |
|---|---|
| `lifecycle.md` | How each abstract action (`orient`, `check-drift`, `verify`, `review`, `learn`, `bootstrap`, `audit`, `synthesize`, `migrate`, `mode`) is invoked. |
| `sensors.md` | How sensors actually fire — autonomous shell execution, surfaced commands for the human to run, or no-op with checklist. |
| `activation.md` | Where ambient **guide** content is projected (root file convention or rules-surface mechanism), how lazy-by-region works for `guides/idioms/`, how the agent reaches **corpus** files on demand, and the agent's context-reset primitive. |

## Supported agents

Every adapter uses the same invocation model: the agent reads its menu file at session start, finds an action in the bulleted list, follows the link to `.charter/actions/<action>.md`, and executes the playbook. The "menu file" differs per agent — that's what this table records.

| Agent | Adapter | Menu file |
|---|---|---|
| Claude Code | [`claude-code/`](claude-code/) | `CLAUDE.md` at repo root |
| Codex CLI | [`codex/`](codex/) | `AGENTS.md` at repo root |
| [pi.dev](https://pi.dev) | [`pi/`](pi/) | `AGENTS.md` at repo root |
| Cursor | [`cursor/`](cursor/) | `.cursor/rules/keystone.mdc` (`alwaysApply: true`) |
| Aider | [`aider/`](aider/) | `CONVENTIONS.md` (loaded via `.aider.conf.yml` `read:`) |
| GitHub Copilot | [`github-copilot/`](github-copilot/) | `.github/copilot-instructions.md` |
| Continue | [`continue/`](continue/) | `.continuerules` + optional `config.yaml` |
| Cline / Roo Code | [`cline/`](cline/) | Custom instructions field + `.clinerules` / `.roorules` |
| Goose | [`goose/`](goose/) | `.goosehints` |
| (any other) | [`_generic/`](_generic/) | the floor: agent reads markdown on demand |

## How adapters degrade

Not every agent can do everything Claude Code can. The corpus is written so it still works under reduced capability — adapters document the degradation explicitly. The three common failure modes:

1. **No autonomous sensor execution.** Agent cannot run shell during a turn. Sensors degrade to "agent surfaces the command, human runs it, agent reads the output."
2. **No sub-agent parallelism.** `review` collapses from "spawn N reviewers in parallel" to "sequential reviewer passes" or "single combined reviewer prompt."
3. **No autonomy levels.** Pacing modes (paired / solo / autopilot) collapse to a single mode. The phases still run; the user-facing pace is fixed.

Each adapter calls out which of these apply.

## Adding a new adapter

1. Create `adapters/<agent>/` with the three files above.
2. Fill in the **lifecycle table** — every action in `.charter/guides/process/` must have a row, even if the row says "not supported on this agent."
3. Fill in the **sensors table** — every sensor in `.charter/sensors/` gets a binding (autonomous / surfaced / no-op).
4. Fill in **activation** — where ambient **guide** content projects to, how the agent reaches **corpus** files on demand, glob/region matching for `guides/idioms/`, context-reset primitive.
5. Add a row to the table above.
6. Add the matching installable artifacts to `targets/<agent>/` at repo root (separate from the charter — this is what gets dropped into a consumer project).

## The policy/charter separation invariant still applies

`.charter/adapters/<agent>/` is documentation about how the agent binds to the charter. The actual artifacts that get installed into a consumer's project live in `targets/<agent>/` at repo root. Do not duplicate content between the two; the adapter doc *describes*, the target *ships*.
