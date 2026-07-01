# Goose — Lifecycle binding

How each abstract lifecycle action is invoked in Block's Goose.

## Invocation

Goose is a CLI / desktop agent. It reads `.goosehints` at the repo root (workspace scope) and `~/.config/goose/hints.md` (global). Every action is invoked via natural language: "run task on TICKET-123," "run verify," "do a review pass." The agent finds the action in the menu's bulleted list, follows the link to `.charter/actions/<action>.md`, and executes the playbook.

The canonical kickoff phrase is **"run task on `<ticket-id>`"** (or "run the task workflow") — `.charter/actions/task.md` orchestrates `spec → orient → implementation → check-drift → verify → review`.

## Optional: recipes for one-shot invocation

Power users who want non-interactive runs can define a Goose recipe per action — each recipe just invokes the canonical playbook. Recipes aren't required; natural language against the menu works without them.

## Required extensions

The charter assumes the **developer extension** is enabled in Goose. It provides:

- Shell execution (`bash` tool) — needed by every sensor.
- File read/write — needed for reading guides, corpus, and writing state files.

Enable with:

```bash
goose configure --enable-extension developer
```

Or via the desktop app: Settings → Extensions → Developer → Enable.

Without the developer extension, Goose has no native tool use; sensor commands degrade to "agent surfaces, user runs in a separate terminal."

## Sub-agent support

None. Goose runs a single session per invocation; the **review** action runs each review concern sequentially.

Goose's `sub-agent` extension (experimental) can spawn focused child agents, but the children execute sequentially in the session — not in parallel.

## Modes

Goose has no autonomy levels in the charter sense. There is:

- **Interactive mode** (default for `goose session`) — pauses for user input between turns.
- **Non-interactive mode** (`goose run --recipe`) — runs the recipe to completion without prompts.

The charter's `paired`/`solo`/`autopilot` collapse approximately:

| Charter mode | Goose invocation |
|---|---|
| **paired** | `goose session` (interactive); approve commands as they come up |
| **solo** | `goose session` with `--with-extension` for ones you trust; manual approval for risky ops |
| **autopilot** | `goose run --recipe <name>` — non-interactive, runs to completion |

## Capability matrix

| Capability | Supported? |
|---|---|
| Autonomous shell execution | ✓ (developer extension's `bash` tool) |
| Sub-agent parallelism | ✗ — sub-agent extension is sequential and experimental |
| Autonomy levels | partial — interactive vs. `goose run --recipe` |
| Lazy-by-region | ✗ — Goose has no glob-based auto-attach |
| Context-reset primitive | new session (`goose session`); or `/exit` and restart in interactive mode |
| Tracker integration | via MCP extension (e.g., GitHub, Atlassian) if installed and configured |
| GitHub integration | via the `github` MCP extension or shell `gh` |
