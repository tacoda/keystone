---
last_reconciled: <YYYY-MM-DD>
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
