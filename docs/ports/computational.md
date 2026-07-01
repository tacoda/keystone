# Port: Computational guide

**Activation:** Ambient. The underlying tool (LSP, formatter, pre-commit hook) runs continuously; the charter file documents its existence for the **stack-drift** sensor.
**Purpose:** Inventory of deterministic enforcement the project's stack already provides — language servers, formatters-on-save, editor lint, pre-commit hooks scoped to formatting. The agent doesn't reason over these files at runtime; the **stack-drift** sensor uses them to detect when documented enforcement diverges from what's actually wired.

## Path convention

```
.charter/guides/computational/<tool>.md                      # project-owned
.charter/policies/<policy>/guides/computational/<tool>.md     # policy-owned (read-only)
```

One file per tool. Flat under `computational/` — no further sub-directories.

## Required shape

```markdown
---
kind: computational
id: guides/computational/<tool>
description: 'One sentence — what the tool covers.'
globs:                # optional, documentation-only
  - "<pattern>"
---

# <tool name>

**What it covers** — <e.g. types, formatting, lint, unused imports>.
**Activation** — <e.g. LSP in editor, on-save formatter, pre-commit hook>.
**Authority** — <treat findings as: blocking | advisory | informational>.
**Configured by** — <path to config file, or "none">.
```

- **`kind: computational`** — required.
- **`globs:` here are documentation, not enforcement.** The tool reads its own config; recording paths here lets **stack-drift** compare documented vs. effective.

## Cascade behavior

Project wins. Pruning flywheel removes entries when the tool stops running locally (the **charter-debt** sensor flags `failing-sensor` and `empty-shell`).

## Example

See `.charter/guides/computational/gofmt.md`.

## Authoring

Inventoried by the **bootstrap** action from the project's stack. Manual authoring is rare — add a file only when the tool was wired up after bootstrap ran and re-running bootstrap is too heavy. Then re-run `keystone index`.
