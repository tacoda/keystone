---
kind: computational
---

# Computational guides

Deterministic ambient enforcement. Anything that shapes the agent's behavior in real time without requiring the agent to reason over natural-language rules — language servers, formatters-on-save, editor lint-as-you-type, pre-commit hooks scoped to formatting, type-checked autocomplete.

This directory ships empty. The **bootstrap** action populates it by inventorying what the project's stack actually supports (an LSP for the primary language, a formatter wired into the editor, the lint configuration the editor reads). The inventory lands here as one file per tool, plus a row in `corpus/state/CODEBASE_STATE.md`.

## File shape

Each file declares what the tool is, how it activates, and how the agent should treat its signal.

```markdown
---
kind: computational
---

# <tool name>

**What it covers** — <e.g. types, formatting, lint, unused imports>.
**Activation** — <e.g. LSP in editor, on-save formatter, pre-commit hook>.
**Authority** — <treat findings as: blocking | advisory | informational>.
**Configured by** — <path to config file in the repo>.
```

## Anti-patterns

- A computational guide whose result the agent ignores. If a deterministic check is wired up, its signal must be honored (or its authority downgraded explicitly).
- A computational guide that overlaps an existing sensor (e.g., an editor lint and a CI lint that disagree). One source of truth per concern; reference the same config.
- Adding entries by hand that the toolchain does not actually run. Bootstrap inventories what exists; if a tool isn't running locally, it doesn't belong here.

## Install-time options

If a computational guide requires shipping a config file, a binary, or an agent setting, the project surfaces it as a flag on `keystone init` rather than scaffolding it unconditionally.
