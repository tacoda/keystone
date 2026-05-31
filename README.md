# keystone

A project harness for coding agents. Drops a self-updating corpus of engineering knowledge into your repo, wired to whichever agent you use.

> **Status:** working temp name; the product itself is the harness, not the installer.

## What it is

Keystone is a markdown-only **project harness** — no language runtime, no daemon. The installer copies two things into your repo:

1. **A corpus** (`harness/`) — five layers of engineering knowledge: principles, idioms, domain, state, process. Plus per-agent bindings under `harness/adapters/<agent>/`.
2. **An activation file** (the "menu") — `CLAUDE.md`, `AGENTS.md`, `.cursor/rules/*.mdc`, `CONVENTIONS.md`, etc. — depending on your agent. The menu tells the agent to read the cookbook.

After install, your agent drives a six-phase workflow (spec → planning → implementation → verification → review → release) and two flywheels (Learning adds rules, Pruning removes stale ones).

## Install

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/tacoda/keystone/main/install.sh | sh
```

To inspect before running (recommended):

```bash
curl -fsSL https://raw.githubusercontent.com/tacoda/keystone/main/install.sh > install.sh
less install.sh
sh install.sh
```

To specify the agent explicitly:

```bash
sh install.sh claude-code
sh install.sh codex
sh install.sh pi
# etc.
```

### Windows

```powershell
iwr -useb https://raw.githubusercontent.com/tacoda/keystone/main/install.ps1 | iex
```

Or download and inspect:

```powershell
iwr https://raw.githubusercontent.com/tacoda/keystone/main/install.ps1 -OutFile install.ps1
Get-Content install.ps1 | more
.\install.ps1 -Agent claude-code
```

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
- **Sensors** — lint, type-check, test, build, optionally coverage. Their commands live in `harness/state/CODEBASE_STATE.md`.
- **PR workflow** — the review phase spawns review agents on a diff; the release phase opens the PR.
- **CI pipeline** (ideally CD) — release assumes CI runs on PRs and CD triggers on merge.

Missing one degrades the corresponding phase but does not break the harness.

## After install

1. Read `harness/README.md` — five-minute orientation to the corpus.
2. Ask your agent to run the **bootstrap** action — it populates `harness/state/CODEBASE_STATE.md` and `harness/idioms/<your-stack>/` from your project.
3. Commit `harness/` and any agent files the installer created.

From then on, every task flows through the six phases, and the Learning flywheel grows the corpus as your project teaches you new patterns.

## Updating

Keystone has no update CLI. Either:

- **Re-run the installer** — fetches the latest, asks before overwriting any file you've edited.
- **Pull from this repo** — clone keystone alongside your project and `cp -R keystone/harness/principles/. your-project/harness/principles/` for the layers you want updated. Principles rarely change; process and adapters change more often.

The corpus is yours after install — keystone is not a runtime dependency.

## Layout

```
keystone/
├── README.md                # this file
├── install.sh               # bash installer (macOS / Linux)
├── install.ps1              # PowerShell installer (Windows)
├── harness/                 # the corpus skeleton dropped into consumer projects
│   ├── README.md
│   ├── principles/          # universal engineering rules
│   ├── idioms/              # stack-specific patterns (per-project)
│   ├── domain/              # project-specific business rules (template)
│   ├── state/               # empirical map of the codebase (template)
│   ├── process/             # six workflow phases + sensors + modes
│   ├── adapters/            # per-agent bindings
│   ├── learning/            # Learning flywheel staging
│   └── archive/             # Pruning flywheel destination
└── targets/                 # per-agent menu files installed into consumer projects
    ├── _generic/            # universal AGENTS.md fallback
    ├── claude-code/         # CLAUDE.md
    ├── codex/               # AGENTS.md
    ├── pi/                  # AGENTS.md + .pi/prompts/*.md slash templates
    ├── cursor/              # .cursor/rules/000-harness.mdc
    ├── aider/               # CONVENTIONS.md
    ├── github-copilot-cli/  # .github/copilot-instructions.md
    ├── continue/            # .continuerules
    ├── cline/               # paste-into-settings text
    └── goose/               # .goosehints
```

## Contributing

If you write a real adapter for an agent currently listed as a stub, contribute it back. The shape is documented in [`harness/adapters/README.md`](harness/adapters/README.md): a `lifecycle.md`, `sensors.md`, and `activation.md` per agent, plus a target directory under `targets/<agent>/`.

## License

> _TODO: pick a license. MIT or Apache-2.0 are sensible defaults for a corpus + installer._
