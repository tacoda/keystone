# Changelog

All notable changes to keystone are documented here. The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/); the project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html) and is pre-1.0 (minor versions may include breaking changes).

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
