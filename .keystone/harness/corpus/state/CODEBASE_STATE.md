---
kind: corpus
id: corpus/state/CODEBASE_STATE
description: 'Empirical map of this codebase.'
last_reconciled: 2026-06-18
---
# Codebase State

Empirical map of this codebase. Updated by the **verify**, **learn**, and **audit** actions.

## Tool commands

The commands each sensor invokes. `(none)` means the project doesn't currently wire that tool.

| Tool | Command |
|---|---|
| lint | `go vet ./...` |
| type_check | `go build ./...` |
| test | `go test ./...` |
| build | `go build ./...` |
| coverage | `go test -cover ./...` |
| secret_scan | `(none)` |
| vuln_scan | `govulncheck ./...` |
| sast | `(none)` |

Severity thresholds (used by vuln-scan and sast):

| Tool | Fail at or above |
|---|---|
| vuln_scan | `high` (default) |
| sast | `error` (default) |

## Sensors

Inventory of sensors wired up for this project. A sensor's row is omitted when it does not apply.

| Sensor | Kind | Status |
|---|---|---|
| lint | computational | wired |
| type-check | computational | wired |
| test | computational | wired |
| build | computational | wired |
| coverage | computational | wired |
| drift | computational | wired |
| commit-message | computational | wired |
| state-region | computational | wired |
| risk-fingerprint | computational | `(none)` |
| traffic-topology | computational | `(none)` |
| tracker-card-fetcher | computational | `(none)` |
| quality-radar | computational | `(none)` |
| code-debt | computational | `(none)` |
| harness-debt | computational | wired |
| stack-drift | computational | wired |
| secret-scan | computational | `(none)` |
| vuln-scan | computational | wired |
| sast | computational | `(none)` |
| spec-adherence | inferential | wired |
| review-functional | inferential | wired |
| review-security | inferential | wired |
| review-risk | inferential | wired |
| review-deployment | inferential | wired |

## Guides

Inferential guides (markdown rules) are activated by directory — see `harness/guides/`. The table below tracks **computational guides** the bootstrap action detected.

| Tool | Kind | What it covers | Activation |
|---|---|---|---|
| `gofmt` | computational | formatting (canonical Go layout, imports) | `go fmt ./...` / editor on-save |
| `go vet` | computational | suspicious constructs (shadowing, printf args, unreachable code) | `go vet ./...` / pre-commit |
| `gopls` | computational | types, completion, unused imports, refactor hints | LSP in editor |

No `.editorconfig`, `.golangci.yml`, or `.pre-commit-config.yaml` is configured.

## Stacks

This repo is the keystone framework dogfooding itself. Two stacks coexist: the Go implementation, and the markdown harness content the framework defines.

| Stack | Idiom folder | Region(s) |
|---|---|---|
| harness-content | `harness/corpus/idioms/harness-content/` (+ paired `harness/guides/idioms/harness-content/`) | `.keystone/harness/`, `internal/framework/templates/` (if present) |
| go | `harness/corpus/idioms/go/` (+ paired `harness/guides/idioms/go/`) | `cmd/`, `internal/` |

**Primary stack:** `harness-content`. The Go code exists to read, write, and validate harness content; the content's shape is the load-bearing surface. Go idioms matter, but document-structure rules matter more.

## Frameworks & libraries

Notable dependencies shaping how code is written.

| Name | Version | Role | Region(s) |
|---|---|---|---|
| `github.com/spf13/cobra` | v1.10.2 | CLI command framework | `cmd/`, `internal/framework/` |
| `github.com/mark3labs/mcp-go` | v0.55.0 | MCP server runtime | `internal/framework/` |
| `github.com/fsnotify/fsnotify` | v1.10.1 | filesystem watcher | `internal/framework/` |
| `gopkg.in/yaml.v3` | v3.0.1 | YAML parsing (frontmatter, configs) | global |
| `golang.org/x/term` | v0.43.0 | terminal I/O (prompts, TUI) | `internal/framework/` |

## Regions

A "region" is a directory or set of directories with shared conventions.

### `cmd/keystone/`
- **Idioms:** `harness/corpus/idioms/go/`
- **Coverage:** n/a (CLI wiring; behavior covered by `internal/` tests)
- **Last reconciled:** 2026-06-18
- **Active migrations:** Keystone 2.0 cleanup

### `internal/framework/`
- **Idioms:** `harness/corpus/idioms/go/`
- **Coverage:** n/a (no coverage tooling wired; `go test -cover ./...` available)
- **Last reconciled:** 2026-06-18
- **Active migrations:** Keystone 2.0 cleanup

### `docs/`
- **Idioms:** none (prose, markdown)
- **Coverage:** n/a
- **Last reconciled:** 2026-06-18
- **Active migrations:** none

### `.keystone/harness/`
- **Idioms:** `harness/corpus/idioms/harness-content/` (primary)
- **Coverage:** n/a (validated structurally by `keystone verify` + the drift / harness-debt sensors)
- **Last reconciled:** 2026-06-18
- **Active migrations:** Keystone 2.0 cleanup

## CI

GitHub Actions. Single workflow `.github/workflows/release.yml` — release-only, fires on `v*` tag push, runs goreleaser. **No PR/test CI** wired today (release-via-tag-push policy; see user memory).

## Methodology

- **Pacing mode:** paired (active — see `harness/guides/process/modes.md`).
- **Testing posture:** TDD — write failing test, get review, smallest impl that turns at least one new test green.
- **Code review:** sensor-driven on diff (`review` action runs the four review sensors).
- **Compliance scope:** none. No SOX/HIPAA/SOC 2 starter rules apply.

## Aspirational notes

Match current Go idioms: stdlib first, small interfaces, cobra commands, table-driven tests. No aspirational pattern set beyond what the code already does today.

## Code-state assessment

Active areas:
- `internal/framework/` — primary implementation surface; high churn during 2.0 cleanup.
- `.keystone/harness/` — the harness primitives this project both ships and consumes (dogfooded).

Do-not-touch: none flagged. Deprecated subsystems: none flagged.
