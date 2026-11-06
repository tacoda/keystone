---
kind: corpus
id: corpus/idioms/charter-content/README
description: 'Charter content idioms — markdown primitive shape, paths, pairing.'
---
# Charter content idioms — reasoning

This repo authors and dogfoods the keystone charter. The primary stack is **markdown primitives** under `.charter/`. Same content also ships as the install-time template inside `internal/framework/scaffold/templates/` (when present), so changes to one must be reflected in the other.

## Primitive kinds

The twelve-kind taxonomy enforced by the keystone charter:

- **action** — atomic unit of work (`.charter/actions/<id>.md`).
- **playbook** — ordered sequence of actions (`.charter/playbooks/<id>.md`).
- **guide** — prescriptive rule file with optional `globs:` (`.charter/guides/<topic>/<id>.md`).
- **corpus** — informational *why* paired to a guide or to state (`.charter/corpus/<topic>/<id>.md`).
- **sensor** — automated check, computational or inferential (`.charter/sensors/<id>.md`).
- **persona** — subagent system-prompt body (`.charter/personas/<id>.md`).
- **skill** — host-native auto-activation manifest (`.charter/skills/<slug>/SKILL.md`).
- **subagent** — host-native agent definition (`.charter/agents/<id>.md`).
- **command** — host slash command (`.charter/commands/<id>.md`).
- **rule** — host-native rule directive (`.charter/rules/<id>.md`).
- **computational** — deterministic ambient enforcement (`.charter/guides/computational/<tool>.md`).
- **adapter** — per-agent host wiring (`.charter/adapters/<agent>/`).

## Layout

- `primitive-shape.md` — frontmatter, canonical paths, the corpus/guide pair convention, the narrow-only `globs:` contract.
- `state-files.md` — how state files (CODEBASE_STATE, GLOBS_INDEX, INSTALL_PROFILE, code-debt, charter-debt) are written, by whom, and what must not be hand-edited.
- (more idioms accumulate via the **learn** → **synthesize** flywheel.)

## What makes this stack different

- Files are **read by an agent at runtime**, not compiled. Naming, frontmatter, and paths *are* the API.
- Edits propagate through `keystone index` → `.keystone/INDEX.json`. Manual edits to INDEX.json drift instantly.
- Project-layer files always win over installed-policy files unless a policy is marked `strict`. This is the cascade contract — guides must not assume they are absolute.
