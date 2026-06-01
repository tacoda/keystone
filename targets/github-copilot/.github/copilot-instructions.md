# Copilot instructions — Keystone harness

This project uses a **project harness**. The corpus at [`harness/`](../harness/) defines the engineering knowledge and the workflow phases you will operate within.

Read these first when starting any task:

- [`harness/README.md`](../harness/README.md) — the five layers, the lifecycle actions, the flywheels.
- [`harness/adapters/github-copilot/`](../harness/adapters/github-copilot/) — how each lifecycle action binds to GitHub Copilot.
- [`harness/domain/`](../harness/domain/) — the business rules for this project.

## Five layers of the corpus

- `harness/principles/` — universal engineering rules.
- `harness/idioms/` — stack-specific patterns.
- `harness/domain/` — business rules for this project.
- `harness/state/` — empirical map of the codebase right now.
- `harness/process/` — six workflow phases (spec → planning → implementation → verification → review → release).

## Lifecycle actions

Invoke by asking in plain language. Each maps to a phase doc you should read before acting.

| Action | Trigger | Reads |
|---|---|---|
| spec | "Start the spec phase for `<task>`." | `harness/process/spec.md` |
| orient | "Orient for work in `<region>`." | `harness/process/planning.md`, `state/CODEBASE_STATE.md`, matching idioms |
| check-drift | "Check the diff for drift." | `harness/process/implementation.md`, loaded corpus rules |
| verify | "Run the verify action." | `harness/process/verification.md`, sensor commands from `state/CODEBASE_STATE.md` |
| review | "Run the review action." | `harness/process/review.md`, spec, diff |
| learn | "Capture the learnings." | `harness/process/release.md` |
| bootstrap | "Bootstrap the harness." | One-time; populates `idioms/` and `state/` |

## Iron laws

Non-negotiable across every phase:

- **No proceeding without explicit acceptance criteria** in the spec.
- **No completion claims without fresh verification evidence** — sensors must have run this turn.
- **No commits with failing sensors.** Never `--no-verify`.
- **No AI attribution** in commits, PRs, or tracker comments — no `Co-Authored-By: Copilot`, no auto-generated footers.
- **No silent overwrites** of state files — propose a diff, confirm before applying.

## GitHub-native primitives

This adapter uses Copilot's native GitHub integration where it exists:

- **Issue/PR context** — read directly from GitHub for the **spec** action when the tracker is GitHub Issues.
- **`gh` CLI** — used for tracker-card-fetcher (`gh issue view`), CI status (`gh run list`, `gh run view`), and PR operations (`gh pr create`).
- **Code search** — `gh search code` and `gh search issues` for cross-repo lookups.

For Jira, Linear, or Asana trackers, paste card content into the chat — there is no native MCP equivalent for non-GitHub trackers in this adapter.

## Prerequisites

- A way to track work — a tracker card (GitHub Issues preferred for this adapter), a `TODO.md`, or a conversation.
- Lint / type-check / test / build commands in `harness/state/CODEBASE_STATE.md`.
- PR workflow on GitHub.
- CI (and ideally CD) via GitHub Actions or a connected runner.

## Operating mode

Copilot's only autonomy lever is per-command approval. Treat every session as **paired** — confirm each shell command before it runs.
