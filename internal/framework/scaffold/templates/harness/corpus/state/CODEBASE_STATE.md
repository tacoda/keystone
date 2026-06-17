---
kind: corpus
id: corpus/state/CODEBASE_STATE
description: 'Empirical map of this codebase.'
---
# Codebase State

> **Template.** The **bootstrap** action will populate this from your project. Until then, fill in by hand or run bootstrap.

Empirical map of this codebase. Updated by the **verify**, **learn**, and **audit** actions.

## Tool commands

The commands each sensor invokes. Filled in by **bootstrap** from your manifest / config files. `(none)` is a valid value when a project doesn't have that tool.

| Tool | Command |
|---|---|
| lint | `<your lint command, e.g. eslint .>` |
| type_check | `<your type-check command, e.g. tsc --noEmit>` |
| test | `<your test command, e.g. pytest>` |
| build | `<your build command, e.g. npm run build>` |
| coverage | `<your coverage command, e.g. pytest --cov>` |
| secret_scan | `<e.g. gitleaks detect>` |
| vuln_scan | `<e.g. trivy fs . / npm audit / pip-audit>` |
| sast | `<e.g. semgrep ci / bandit / gosec>` |

Severity thresholds (used by vuln-scan and sast):

| Tool | Fail at or above |
|---|---|
| vuln_scan | `<low \| medium \| high \| critical>` (default: high) |
| sast | `<info \| warning \| error>` (default: error) |

## Sensors

Inventory of sensors wired up for this project. Filled in by **bootstrap**. A sensor's row is omitted when it does not apply (no tracker, no spec workflow, adapter cannot run sub-agents, etc.).

| Sensor | Kind | Status |
|---|---|---|
| lint | computational | `<wired \| (none)>` |
| type-check | computational | `<wired \| (none)>` |
| test | computational | `<wired \| (none)>` |
| build | computational | `<wired \| (none)>` |
| coverage | computational | `<wired \| (none)>` |
| drift | computational | wired |
| commit-message | computational | wired |
| state-region | computational | wired |
| risk-fingerprint | computational | `<wired \| (none)>` |
| traffic-topology | computational | `<wired \| (none)>` |
| tracker-card-fetcher | computational | `<wired \| (none)>` |
| quality-radar | computational | `<wired \| (none)>` |
| code-debt | computational | `<wired \| (none)>` |
| harness-debt | computational | wired |
| stack-drift | computational | wired |
| secret-scan | computational | `<wired \| (none)>` |
| vuln-scan | computational | `<wired \| (none)>` |
| sast | computational | `<wired \| (none)>` |
| spec-adherence | inferential | wired |
| review-functional | inferential | `<wired \| (none)>` |
| review-security | inferential | `<wired \| (none)>` |
| review-risk | inferential | `<wired \| (none)>` |
| review-deployment | inferential | `<wired \| (none)>` |

## Guides

Inferential guides (markdown rules) are activated by directory — see `harness/guides/`. The table below tracks **computational guides** the bootstrap action detected. Filled in by **bootstrap**.

| Tool | Kind | What it covers | Activation |
|---|---|---|---|
| `<e.g. typescript-language-server>` | computational | `<e.g. types, completion, unused imports>` | `<e.g. LSP in editor>` |

## Stacks

| Stack | Idiom folder | Region(s) |
|---|---|---|
| `<stack-name>` | `harness/corpus/idioms/<stack-name>/` (+ paired `harness/guides/idioms/<stack-name>/`) | `<paths>` |

## Frameworks & libraries

Notable frameworks and libraries the codebase depends on, beyond what the stack name implies. Filled in by **bootstrap** from manifests (`package.json`, `composer.json`, `Gemfile`, `pyproject.toml`, `go.mod`, `Cargo.toml`, etc.). Limit to dependencies that shape how code is written — routers, ORMs, validation, HTTP clients, UI kits, test frameworks. Skip transitive noise.

| Name | Version | Role | Region(s) |
|---|---|---|---|
| `<name>` | `<version or range>` | `<one-line role, e.g. ORM, HTTP client, validation>` | `<paths or "global">` |

## Regions

A "region" is a directory or set of directories with shared conventions. Examples: `src/`, `tests/`, `infrastructure/`, `docs/`.

### `<region-path>`
- **Idioms:** `<which idiom folders apply>`
- **Coverage:** `<percentage or n/a>`
- **Last reconciled:** `<YYYY-MM-DD>`
- **Active migrations:** `<names or none>`
