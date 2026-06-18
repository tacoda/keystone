---
kind: corpus
id: corpus/idioms/harness-content/README
description: 'Harness content idioms — markdown primitive shape, paths, pairing.'
---
# Harness content idioms — reasoning

This repo authors and dogfoods the keystone harness. The primary stack is **markdown primitives** under `.keystone/harness/`. Same content also ships as the install-time template inside `internal/framework/templates/` (when present), so changes to one must be reflected in the other.

## Primitive kinds

The twelve-kind taxonomy enforced by the keystone harness:

- **action** — atomic unit of work (`harness/actions/<id>.md`).
- **playbook** — ordered sequence of actions (`harness/playbooks/<id>.md`).
- **guide** — prescriptive rule file with optional `globs:` (`harness/guides/<topic>/<id>.md`).
- **corpus** — informational *why* paired to a guide or to state (`harness/corpus/<topic>/<id>.md`).
- **sensor** — automated check, computational or inferential (`harness/sensors/<id>.md`).
- **persona** — subagent system-prompt body (`harness/personas/<id>.md`).
- **skill** — host-native auto-activation manifest (`harness/skills/<slug>/SKILL.md`).
- **subagent** — host-native agent definition (`harness/agents/<id>.md`).
- **command** — host slash command (`harness/commands/<id>.md`).
- **rule** — host-native rule directive (`harness/rules/<id>.md`).
- **computational** — deterministic ambient enforcement (`harness/guides/computational/<tool>.md`).
- **adapter** — per-agent host wiring (`harness/adapters/<agent>/`).

## Layout

- `primitive-shape.md` — frontmatter, canonical paths, the corpus/guide pair convention, the narrow-only `globs:` contract.
- `state-files.md` — how state files (CODEBASE_STATE, GLOBS_INDEX, INSTALL_PROFILE, code-debt, harness-debt) are written, by whom, and what must not be hand-edited.
- (more idioms accumulate via the **learn** → **synthesize** flywheel.)

## What makes this stack different

- Files are **read by an agent at runtime**, not compiled. Naming, frontmatter, and paths *are* the API.
- Edits propagate through `keystone index` → `.keystone/INDEX.json`. Manual edits to INDEX.json drift instantly.
- Project-layer files always win over installed-plugin files unless a policy is marked `strict`. This is the cascade contract — guides must not assume they are absolute.
