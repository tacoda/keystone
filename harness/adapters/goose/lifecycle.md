# Goose — Lifecycle binding

How each abstract lifecycle action is invoked in Block's Goose.

## Action → invocation

Goose is a CLI / desktop agent. It reads `.goosehints` at the repo root (workspace scope) and `~/.config/goose/hints.md` (global). Lifecycle actions are invoked by **typing the action name as natural language in the chat**.

Goose supports **recipes** — saved, parameterized sessions defined in `recipe.yaml`. The harness defines one recipe per lifecycle action; users invoke them with `goose run --recipe <name>`.

| Action | Invocation | What happens |
|---|---|---|
| **spec** | `goose run --recipe keystone-spec` or "Start the spec phase for `<task>`." | Goose reads `harness/guides/process/spec.md` and follows its activities. Tracker fetch via MCP extension or shell (`gh issue view`). |
| **orient** | `goose run --recipe keystone-orient` or "Orient for work in `<region>`." | Goose reads `harness/corpus/state/CODEBASE_STATE.md` and the matching idioms; sketches a plan. |
| **check-drift** | `goose run --recipe keystone-check-drift` or "Check the diff for drift." | Goose runs `git diff` via the developer extension and compares against loaded guide rules. |
| **verify** | `goose run --recipe keystone-verify` or "Run the verify action." | Goose invokes sensors via the developer extension's shell tool; reports results. |
| **review** | `goose run --recipe keystone-review` or "Run the review action." | Goose walks the diff against spec AC, then runs functional and security review concerns **sequentially**. |
| **learn** | `goose run --recipe keystone-learn` or "Capture learnings from this work." | Goose writes a candidate to `harness/learning/inbox/<timestamp>-<slug>.md`. |
| **bootstrap** | `goose run --recipe keystone-bootstrap` or "Bootstrap the harness." | One-time; detects stack, frameworks, and libraries; seeds corpus (idioms/<stack>/, state/), paired guides (idioms/<stack>/), and confirms sensor commands. |
| **audit** | `goose run --recipe keystone-audit` or "Audit the harness." | Full Learning + Pruning flywheel pass. |
| **synthesize** | `goose run --recipe keystone-synthesize` or "Synthesize the inbox." | Promotes inbox items into the right corpus and/or guide. |
| **mode** | Edit `harness/guides/process/modes.md` directly. | Goose has no autonomy levels in the harness sense; the file is informational. |

## Required extensions

The harness assumes the **developer extension** is enabled in Goose. It provides:

- Shell execution (`bash` tool) — needed by every sensor.
- File read/write — needed for reading guides, corpus, and writing state files.

Enable with:

```bash
goose configure --enable-extension developer
```

Or via the desktop app: Settings → Extensions → Developer → Enable.

Without the developer extension, Goose has no native tool use; sensor commands degrade to "agent surfaces, user runs in a separate terminal."

## Recipes

Goose recipes are YAML files that define a reusable session with extensions, prompt, and parameters. The harness ships one per lifecycle action under `targets/goose/recipes/` (the installer drops them into `<project>/.goose/recipes/`).

Sample recipe (`keystone-verify.yaml`):

```yaml
version: 1.0.0
title: Keystone — verify
description: Run every sensor against the current change
instructions: |
  Read harness/guides/process/verification.md and run the verify action.
  Sensor commands live in harness/corpus/state/CODEBASE_STATE.md.
  Report each sensor's pass/fail and the structured output on failure.
extensions:
  - name: developer
prompt: |
  Run the verify action for the current diff.
```

Run with `goose run --recipe .goose/recipes/keystone-verify.yaml`.

## Sub-agent support

None. Goose runs a single session per invocation; the **review** action runs each review concern sequentially.

Goose's `sub-agent` extension (experimental) can spawn focused child agents, but the children execute sequentially in the session — not in parallel.

## Modes

Goose has no autonomy levels in the harness sense. There is:

- **Interactive mode** (default for `goose session`) — pauses for user input between turns.
- **Non-interactive mode** (`goose run --recipe`) — runs the recipe to completion without prompts.

The harness's `paired`/`solo`/`autopilot` collapse approximately:

| Harness mode | Goose invocation |
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
