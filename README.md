# keystone

A project harness for coding agents. Drops a self-updating set of engineering knowledge, rules, and sensors into your repo, wired to whichever agent you use.

> **Status:** working temp name; the product itself is the harness, not the installer.

## What it is

Keystone is a markdown-only **project harness** — no language runtime, no daemon. The `keystone` CLI is a single Go binary with the entire harness embedded; `keystone init` writes two things into your repo:

1. **A harness** (`harness/`) — four components:
   - `guides/` — **rules**. Always loaded. What the agent must do and not do.
   - `corpus/` — **informational reference**. On-demand. The reasoning behind the rules.
   - `sensors/` — **automated checks**. Lint, type-check, test, build, drift, coverage.
   - `learning/` + `archive/` — **flywheels** that keep the harness current.
   Plus per-agent bindings under `harness/adapters/<agent>/`.
2. **An activation file** (the "menu") — `CLAUDE.md`, `AGENTS.md`, `.cursor/rules/*.mdc`, `CONVENTIONS.md`, etc. — depending on your agent. The menu tells the agent to read the harness.

After `init`, the binary is no longer required — the harness and menu are plain markdown files you own.

After install, your agent drives a six-phase workflow (spec → planning → implementation → verification → review → release) and two flywheels (Learning adds rules and reasoning; Pruning removes stale guides regularly and stale corpus rarely).

## Install

Keystone ships as a single binary. Three ways to get it:

### Homebrew (macOS / Linux)

```bash
brew install tacoda/tap/keystone
```

### Curl bootstrap (macOS / Linux)

Downloads the binary into `~/.local/bin/keystone`, then prompts for the agent and runs `keystone init`:

```bash
curl -fsSL https://raw.githubusercontent.com/tacoda/keystone/main/install.sh | sh
```

Inspect before running (recommended):

```bash
curl -fsSL https://raw.githubusercontent.com/tacoda/keystone/main/install.sh > install.sh
less install.sh
sh install.sh
```

Set `KEYSTONE_NO_INIT=1` to install the binary only.

### PowerShell (Windows)

```powershell
iwr -useb https://raw.githubusercontent.com/tacoda/keystone/main/install.ps1 | iex
```

### Manual

Download a prebuilt archive from the [releases page](https://github.com/tacoda/keystone/releases), extract `keystone` (or `keystone.exe`), and place it on your `PATH`.

## Usage

```
keystone init [<dir>] [--agent <name>] [--force]
keystone version
keystone help
```

`init` is non-interactive — it copies `harness/` and the agent's menu file(s) into `<dir>` (default `.`) and exits. If `--agent` is omitted it detects from existing marker files in `<dir>`; if detection fails it errors out. Re-run with `--force` to overwrite an existing `harness/`. Existing target files (e.g. `CLAUDE.md`) are always skipped — review and merge by hand.

## Supported agents

| Agent | Status | Menu file installed |
|---|---|---|
| Claude Code | real adapter | `CLAUDE.md` |
| Codex CLI | real adapter | `AGENTS.md` |
| [pi.dev](https://pi.dev) | real adapter | `AGENTS.md` + `.pi/prompts/` |
| Cursor | stub adapter | `.cursor/rules/000-harness.mdc` |
| Aider | stub adapter | `CONVENTIONS.md` |
| GitHub Copilot CLI | stub adapter | `.github/copilot-instructions.md` |
| Continue | stub adapter | `.continuerules` |
| Cline / Roo Code | stub adapter | `cline-instructions.md` (paste into Cline settings) |
| Goose | stub adapter | `.goosehints` |
| (any other) | generic fallback | `AGENTS.md` |

Stub adapters get a minimal lifecycle file and a working menu — enough to start; fill in the rest as you go.

## Prerequisites

The harness assumes (soft — install runs regardless):

- **A way to track work** — a tracker card (Jira / Linear / GitHub Issues / Asana), a `TODO.md`, or a conversation. The spec phase needs a unit of work to anchor; it doesn't care what tool you use.
- **Sensor commands** — lint, type-check, test, build, optionally coverage. Recorded in `harness/corpus/state/CODEBASE_STATE.md`.
- **PR workflow** — the review phase spawns review agents on a diff; the release phase opens the PR.
- **CI pipeline** (ideally CD) — release assumes CI runs on PRs and CD triggers on merge.

Missing one degrades the corresponding phase but does not break the harness.

## After install

1. Read `harness/README.md` — four-component orientation (corpus, guides, sensors, flywheels).
2. Ask your agent to run the **bootstrap** action — it seeds `harness/corpus/state/CODEBASE_STATE.md`, `harness/corpus/idioms/<your-stack>/`, the paired `harness/guides/idioms/<your-stack>/`, and confirms the sensor commands.
3. Commit `harness/` and any agent files the installer created.

From then on, every task flows through the six phases, and the Learning flywheel grows the harness as your project teaches you new patterns (rules into `guides/`, supplemental reasoning into `corpus/`).

## Layout

```
keystone/
├── README.md                # this file
├── main.go                  # CLI entrypoint + //go:embed all:harness all:targets all:optional
├── init.go                  # `keystone init` command
├── scaffold.go              # walk embedded FS, write files to disk
├── detect.go                # agent detection from marker files
├── go.mod
├── install.sh               # curl bootstrap (downloads binary, prompts, calls init)
├── install.ps1              # PowerShell bootstrap
├── .goreleaser.yml          # cross-compile + Homebrew tap publish
├── .github/workflows/release.yml
├── harness/                 # the harness dropped into consumer projects
│   ├── README.md
│   ├── corpus/              # informational reference (on-demand)
│   │   ├── principles/      # universal engineering knowledge
│   │   ├── idioms/          # stack-specific patterns (per-project)
│   │   ├── domain/          # project-specific business knowledge (template)
│   │   └── state/           # empirical map of the codebase (template)
│   ├── guides/              # rules (always loaded)
│   │   ├── principles/      # rule extracts from corpus/principles/
│   │   ├── idioms/          # rule extracts from corpus/idioms/
│   │   ├── domain/          # business-rule constraints
│   │   └── process/         # six workflow phases + modes
│   ├── sensors/             # automated checks (lint, type-check, test, etc.)
│   ├── adapters/            # per-agent bindings
│   ├── learning/            # Learning flywheel staging
│   └── archive/             # Pruning flywheel destination
├── optional/                # opt-in content seeded by --architecture / --compliance / etc.
│   ├── architecture/<label>/harness/{corpus,guides}/...
│   └── compliance/<label>/harness/{corpus,guides}/...
└── targets/                 # per-agent menu files installed into consumer projects
    ├── _generic/            # universal AGENTS.md fallback
    ├── claude-code/         # CLAUDE.md
    ├── codex/               # AGENTS.md
    ├── pi/                  # AGENTS.md + .pi/prompts/*.md slash templates
    ├── cursor/              # .cursor/rules/000-harness.mdc
    ├── aider/               # CONVENTIONS.md
    ├── github-copilot/      # .github/copilot-instructions.md
    ├── continue/            # .continuerules
    ├── cline/               # paste-into-settings text
    └── goose/               # .goosehints
```

## Contributing

If you write a real adapter for an agent currently listed as a stub, contribute it back. The shape is documented in [`harness/adapters/README.md`](harness/adapters/README.md): a `lifecycle.md`, `sensors.md`, and `activation.md` per agent, plus a target directory under `targets/<agent>/`.

## License

MIT — see [LICENSE](LICENSE).
