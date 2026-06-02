# Changelog

All notable changes to keystone are documented here. The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/); the project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html) and is pre-1.0 (minor versions may include breaking changes).

## [0.3.1] — 2026-06-02

A small install-flow polish. Adds support for projects that use more than one coding agent at a time, smooths over the success message, and introduces a way to add an agent to an existing install without re-running `init`.

### Added

- **`agent` is now multi-select.** Teams using multiple agents (e.g. Claude Code alongside Cursor) can install every target bundle in one pass — either via the interactive prompt or `--agent claude-code,cursor` on the CLI. Each agent's menu file and target bundle are installed; capability-gap warnings print per agent.
- **`monorepo` option for `--app-type`.** Assumes backend + frontend; the **bootstrap** action can refine if the actual structure differs.
- **`keystone add-target <agent>[,<agent>...] [<dir>]` subcommand.** Installs an additional agent's target bundle into an existing harness and merges the new agent(s) into `harness/corpus/state/INSTALL_PROFILE.md`. Errors out if any requested agent is already recorded — remove it first to re-add.

### Fixed

- **Post-install success message** now reads `✓ harness installed for ...` (was `keystone installed`). The binary-install line printed by `install.sh` is unchanged — that one is correctly about the binary itself.

## [0.3.0] — 2026-06-02

A model overhaul. The harness now has **four components** instead of "the corpus plus three roles":

- **Corpus** — informational reference. **Loaded on-demand.**
- **Guides** — rules. **Always loaded.** Surfaced into each agent's rules surface (`.cursor/rules`, `CLAUDE.md` directives, etc.).
- **Sensors** — automated checks. Promoted to a top-level directory.
- **Flywheels** — Learning and Pruning, asymmetric: Pruning churns guides regularly, corpus rarely.

The split is the point: rules are short and high-value-per-token; corpus is long-form. Always-loading guides keeps the agent honest without crowding context with reasoning the situation may not need.

### Added

- **Full adapter implementations for Continue, Cline / Roo Code, and Goose.** Previously stubs. Each now ships `lifecycle.md`, `activation.md`, and `sensors.md` matching the depth of the Claude Code / Codex / pi adapters. Continue gets a documented `config.yaml` with prompts and context providers; Cline gets workflow guidance and auto-approve recommendations; Goose gets recipe templates and developer-extension wiring.
- **Per-agent install-time warnings.** `keystone init` now prints a `⚠ <agent> adapter — capability gaps to address` section before the success message for adapters that do not natively cover every harness feature. Each gap names a configuration remedy and/or a harness file to add (e.g., `harness/adapters/aider/review-strategy.md`). Fully-supported adapters (claude-code, codex, pi, cursor) print no warning.
- **`harness/corpus/`** — informational layer. Houses `principles/`, `idioms/`, `domain/`, `state/`. Read on-demand via forward-links from paired guides, or when process explicitly references a file.
- **`harness/guides/`** — rule layer. Houses `principles/`, `idioms/`, `domain/`, `process/`. Always loaded. Enforced by the drift sensor.
- **`harness/sensors/`** — promoted from `harness/process/sensors.md` (one file) into one file per sensor: `lint`, `type-check`, `test`, `build`, `drift`, `coverage`, `risk-fingerprint`, `traffic-topology`, `state-region`, `commit-message`, `tracker-card-fetcher`, `spec-adherence`, plus a README index.
- **Paired guide files for every principle** that previously carried `## IRON LAW` / `## GOLDEN RULES` sections. The rule sections moved into `harness/guides/principles/<name>.md`; the original corpus file keeps the reasoning, anti-patterns, and references, plus a forward-link.
- **Concern-specific MVC idioms** seeded when `--architecture mvc` is selected: `corpus/idioms/mvc/{models,controllers,views}.md` with paired `guides/idioms/mvc/{models,controllers,views}.md` covering "the model is not a row," "controllers translate, they do not decide," and "views render, they do not compute."
- **Learning flywheel classification.** The **synthesize** action now explicitly routes each inbox candidate as **rule** (lands in `guides/`) or **information** (lands in `corpus/`). The inbox frontmatter carries a `candidate_kind` hint; synthesize confirms or overrides.
- **Pruning flywheel asymmetry.** **audit** runs in two passes — a regular pass over guides (rules churn with the codebase) and a rare pass over corpus (only when design / strategy / ideals have moved on).
- **`harness/guides/idioms/`** and **`harness/guides/domain/`** READMEs documenting the rule-extraction format and the bootstrap/learning population path.

### Changed

- **Bootstrap action** now seeds three things: corpus (`idioms/<stack>/`, `state/`), paired guides (`idioms/<stack>/`), and sensor commands. Adapter lifecycle docs updated across every supported agent.
- **`optional/<cat>/<label>/` bundles** now ship corpus and guides separately. Selecting an architecture or compliance label installs both the explanatory corpus file and the rule-bearing guide file.
- **Activation model.** Corpus is **on-demand only** — the agent reads a corpus file when it follows a forward-link from a guide, when process explicitly names one, or when researching a topic. Guides remain ambient.
- **Adapter framing.** Every adapter's `activation.md` now distinguishes "project this guide into the agent's rules surface" (ambient) from "reach this corpus file when needed" (on-demand).
- **Menu files** (CLAUDE.md, AGENTS.md, .continuerules, .goosehints, CONVENTIONS.md, copilot-instructions.md, etc.) reframed to point at the four components and call out the always-loaded vs. on-demand split.
- **`harness/state/INSTALL_PROFILE.md`** now lives at `harness/corpus/state/INSTALL_PROFILE.md`. `profile.go` updated.

### Removed

- **The "Discipline" role.** It was always an audit action, never a peer of guides/sensors/flywheels. Folded into the audit/synthesize lifecycle.
- **The "corpus = the whole thing" framing.** `corpus` now names a specific component (informational reference). What used to be called "the corpus" is now called "the harness."

### Migration from 0.2.0

Path moves for hand-references inside any project that has installed an earlier version:

| Old path | New path |
|---|---|
| `harness/principles/` | `harness/corpus/principles/` |
| `harness/idioms/` | `harness/corpus/idioms/` |
| `harness/domain/` | `harness/corpus/domain/` |
| `harness/state/` | `harness/corpus/state/` |
| `harness/process/<phase>.md` | `harness/guides/process/<phase>.md` |
| `harness/process/sensors.md` | `harness/sensors/<sensor-name>.md` (one file per sensor) + `harness/sensors/README.md` |
| `harness/state/INSTALL_PROFILE.md` | `harness/corpus/state/INSTALL_PROFILE.md` |

Each principle file previously containing `## IRON LAW` / `## GOLDEN RULES` sections has had those sections moved into a paired `harness/guides/principles/<name>.md`. The corpus file now ends with a forward-link to the guide. If a project has extended a principle file in-place with custom rule sections, hand-port those sections to the matching guide file.

The internal classification convention is: **rules go in `guides/`, reasoning goes in `corpus/`.** When in doubt during Learning flywheel reviews, default to corpus — adding a guide narrows the agent's behavior across the whole project, so the bar should be higher than adding context.

## [0.2.0] — 2026-06-01

A second pass focused on three things: deepening the corpus, broadening the install-time intent surface, and making installs safe to re-run on existing projects.

### Added

- **Interactive `keystone init`** powered by [charmbracelet/huh](https://github.com/charmbracelet/huh). When stdin is a TTY and required options are unset, init prompts for each missing answer; when stdin is not a TTY, it falls back to flags-or-error.
- **Five categories of declared intent** at install time: `--agent`, `--app-type`, `--architecture`, `--testing`, `--compliance`. Multi-select categories accept comma-separated values.
- **`keystone options` subcommand** — lists every allowed label for every flag.
- **Install profile** written to `harness/state/INSTALL_PROFILE.md`, recording the user's selections for the bootstrap action to read.
- **Conditional install plumbing** via `optional/<category>/<label>/<...>`. Files install only when the matching label is selected.
- **24 new principle files** under `harness/principles/`, covering OO design (tell-don't-ask, Demeter, design-by-contract), simplicity & evolution (simple-design, refactoring, pragmatic principles, naming, simplicity), engineering discipline (modern-software-engineering, premature-optimization, fail-fast, error-handling, least-astonishment, postels-law, hyrums-law), production & distributed systems (concurrency, distributed-systems-fallacies, observability, idempotency), testing (tdd, bdd, testing-patterns), and security (security-threats, secrets-management). Each cites real foundational sources and cross-links related principles via `[[name]]`.
- **12 architecture seeds** under `optional/architecture/<label>/`: hexagonal, clean-architecture, onion-architecture, layered, mvc, mvvm, event-driven, microservices, monolith, serverless, spa, continuous-delivery.
- **5 compliance seeds** under `optional/compliance/<regime>/`: gdpr, hipaa, pci-dss, soc2, fedramp.
- **Full adapter implementations** (lifecycle / activation / sensors) for `cursor`, `aider`, and `github-copilot`. Previously stubs.
- **7 starter `.cursor/rules/*.mdc` files** for the cursor target (keystone menu + one per common lifecycle action).
- **Additive menu-file merge** with HTML-comment markers (`<!-- keystone:start -->` / `<!-- keystone:end -->`). If a `CLAUDE.md`, `CONVENTIONS.md`, `.continuerules`, `.goosehints`, or other menu file already exists, the harness section is inserted under the existing H1 (or prepended at top if no H1). Re-installing refreshes the section in place — idempotent.
- **Expanded `harness/README.md`** with a per-action lifecycle table (one sentence each) and a consolidated **Iron laws** section. Menu files now defer the long-form detail to the README.

### Changed

- **Agent rename: `github-copilot-cli` → `github-copilot`.** The single adapter covers both Copilot in VS Code and the Copilot CLI; the suffix was misleading.
- **Agent rename: `_generic` → `generic`** (catalog value). The `targets/_generic/` directory keeps its underscore convention via an internal mapping; users now pass `--agent generic`.
- **Menu-file content is now concise** — read-first index, lifecycle action names, iron laws. Detail moved to `harness/README.md` so the agent's instruction file stays small but discoverable.
- **TTY detection** now uses `golang.org/x/term.IsTerminal` instead of `os.ModeCharDevice`, correctly distinguishing `/dev/null` (character device, not TTY) from a real terminal.

### Removed

The following flags were dropped from `keystone init`. The bootstrap action in your agent will infer these from the codebase on first run, where it has accurate context:

- `--language`
- `--database`
- `--ci-platform`
- `--deployment-target`
- `--project-maturity`
- `--team-size`

### Migration from 0.1.0

- `--agent github-copilot-cli` → `--agent github-copilot`.
- `--agent _generic` → `--agent generic`.
- Any script passing `--language`, `--database`, `--ci-platform`, `--deployment-target`, `--project-maturity`, or `--team-size`: remove those flags. The bootstrap action handles them.
- Pre-existing `CLAUDE.md` / `CONVENTIONS.md` / etc. are now preserved on re-install — the harness inserts a `## Keystone harness` section under your existing H1 instead of overwriting the file.

## [0.1.0] — 2026-06-01

Initial release.

- Embedded-FS Go binary replaces the legacy `install.sh` / `install.ps1` scripts.
- `keystone init [<dir>] [--agent <name>] [--force]` scaffolds `harness/` and the agent's menu file.
- Marker-file detection for 10 agents (claude-code, codex, cursor, aider, github-copilot-cli, continue, cline, goose, pi, _generic).
- GoReleaser-driven release workflow with macOS / Linux / Windows binaries and a Homebrew tap.
