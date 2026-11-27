---
kind: reference
id: idioms/charter-content/README
description: 'Charter content idioms: markdown primitive shape, paths, pairing.'
---
# Charter content idioms, reasoning

This repo authors and dogfoods the keystone charter. The primary stack is **markdown primitives** under `.agents/`. The same content ships as the install-time template under `internal/framework/scaffold/templates/`, so a change to one must reach the other.

## Primitive kinds

Keystone knows ten kinds, no categories. It reads the kind from the first folder under `.agents/`. The triad: **rule** = how, **reference** = why, **corpus** = what.

- **rule**: a directive, scoped by `paths:` (`.agents/rules/<topic>/<id>.md`).
- **reference**: the reasoning behind a rule (`.agents/references/<topic>/<id>.md`).
- **corpus**: standalone information — facts, patterns, product info (`.agents/corpus/<id>.md`), plus repo docs adopted via `charter.roots`.
- **hook**: a check that runs on an event and can block (`.agents/hooks/<id>.md`).
- **permission**: allow, ask, and deny rules for tools (`.agents/permissions/<id>.md`).
- **eval**: a test that the charter changes behavior (`.agents/evals/<id>/EVAL.md`).
- **skill**: an Agent Skills workflow (`.agents/skills/<name>/SKILL.md`).
- **command**: a unit of work the user asks for by name (`.agents/commands/<id>.md`).
- **agent**: a role the agent delegates to (`.agents/agents/<id>.md`).
- **document**: a template for governed output (`.agents/documents/<id>.md`).

The `tool` kind is retired in 6.0: keystone manages context, not callables.

## Layout

- `primitive-shape.md`: frontmatter, canonical paths, the rule and reference pair, and the narrow-only `paths:` contract.
- `state-files.md`: how the state files (CODEBASE_STATE, GLOBS_INDEX, INSTALL_PROFILE, code-debt, charter-debt) are written, by whom, and what must not be hand-edited.
- More idioms accumulate through the **learn** and **synthesize** flywheel.

## What makes this stack different

- An agent reads these files at runtime. Nothing compiles them. Naming, frontmatter, and paths *are* the API.
- Edits propagate through `keystone index` to `.keystone/INDEX.json`. A manual edit to INDEX.json is lost on the next run.
- Project files override installed policy files, unless the policy marks the item `strict`. So a rule must not assume it is absolute.
