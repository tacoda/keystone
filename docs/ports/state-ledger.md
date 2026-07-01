# Port: State ledger

**Activation:** Read at session start (state ledgers are loaded ambient alongside guides); updated by specific actions (`bootstrap`, `audit`, `learn`).
**Purpose:** Project-specific state the agent reads to ground itself and writes to keep current. Mutable corpus.

State ledgers are how the charter remembers what it knows about the project — language, frameworks, sensor commands, debt items, risk fingerprints, traffic topology. Guides and corpus describe rules and reasoning; state ledgers describe **this project, right now**.

## Path convention

```
<charter-root>/corpus/state/<name>.md
```

State lives under `corpus/state/` (a reserved subdirectory of the corpus port). Policys **do not** ship state ledgers — state is per-project by definition.

## Required shape

State ledgers are mostly free-form markdown with optional frontmatter. The framework recognizes specific ledgers by filename:

| Filename               | Purpose                                                                                   | Written by         |
| ---------------------- | ----------------------------------------------------------------------------------------- | ------------------ |
| `CODEBASE_STATE.md`    | Empirical map of the codebase: language, frameworks, libs, db, CI, sensor commands.       | `bootstrap`        |
| `INSTALL_PROFILE.md`   | Selections captured by `keystone init` + link to the lockfile.                            | `keystone init`    |
| `code-debt.md`         | Known code-quality debt items: shape, hotspot files, planned fixes.                       | `audit`            |
| `harness-debt.md`      | Harness-quality debt: stale rules, missing pairs, drift candidates.                       | `audit`, `learn`   |
| `quality-radar.md`     | Per-area quality signal (test coverage, error rate, time-to-recover, etc.).               | `audit`            |
| `risk-fingerprints.md` | Files/areas that are high-risk to change — change blast radius notes.                     | `audit`            |
| `traffic-topology.md`  | How traffic flows through the system (services, queues, sync vs async, fan-out, fan-in).  | `audit`, optional. |

Project-specific ledgers (e.g. `corpus/state/glossary.md`) follow the same shape; the agent reads whatever is present.

## Update semantics

- **Idempotent.** Each action that writes a ledger reads the current state, merges or updates in place, and rewrites. Conflicts that can't be auto-merged are surfaced to the user.
- **Section-bounded.** Updates are scoped to declared headings — a `bootstrap` re-run replaces the `## Sensors` section but leaves the `## Glossary` section untouched.
- **No drift sensor required.** State ledgers are expected to drift from any "ideal" content; they reflect reality. The `drift` sensor doesn't flag them; the `harness-debt` sensor does, but only for ledgers the charter itself owns (CODEBASE_STATE, harness-debt, etc.).

## Cascade behavior

State ledgers are not part of the strict-cascade resolution. They live at the project layer only — policies never ship state. The agent reads them by exact filename.

## Example

```markdown
# Codebase State

Empirical map of this project. Updated by the **bootstrap** action; safe
to edit by hand. Re-run bootstrap to refresh from the codebase.

## Stack

- Language: Go 1.25
- Framework: stdlib HTTP + chi router
- Test framework: stdlib testing + testify
- CI: GitHub Actions (`.github/workflows/release.yml`)

## Sensors

| Sensor      | Command           | Notes                              |
| ----------- | ----------------- | ---------------------------------- |
| lint        | `go vet ./...`    |                                    |
| type-check  | `go build ./...`  | Compiler is the type checker.      |
| test        | `go test ./...`   | Plus `-race` in CI.                |
| sast        | `gosec ./...`     | Opt-in; not run on every commit.   |

## Idioms

- Workhorses in `internal/framework/...`, CLI in `cmd/keystone/`.
- Embed templates via `go:embed`; never duplicate them in tests.
```

## Authoring

State ledgers are written by actions, not by users directly. To add a
new state ledger:

1. Write the markdown by hand under `corpus/state/<name>.md`.
2. Update the relevant action (or write a new one) to read + update it.
3. Document the ledger's shape in `CODEBASE_STATE.md` so the agent
   knows it exists.

There is no `keystone new state-ledger` generator — state shape is
project-specific by design.

## Read by

- The agent at session start (ambient context).
- `bootstrap`, `audit`, `learn` actions (read + update).
- `keystone doctor` — `harness-debt.md` is one of its inputs.
