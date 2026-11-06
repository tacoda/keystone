---
applyTo: .agents/references/state/**/*.md,internal/framework/scaffold/templates/agents/references/state/**/*.md
description: Authoring rules for .agents/references/state/.
---

# State files — rules

Full rule: `.agents/rules/idioms/charter-content/state-files.md` (read on demand).

## IRON LAW

- **No silent overwrites.** Every state-file write goes through the agent's edit primitive so the user sees the diff. Hooks and subagents that update state must surface their diff for review.
- **Do not hand-edit derived sections.** The `## Index` table in `GLOBS_INDEX.md` is regenerated. `quality-radar.md`, `risk-fingerprints.md`, `traffic-topology.md` are snapshots — replace the whole file through the hook or subagent that writes it, never patch.


## GOLDEN RULE

- `CODEBASE_STATE.md`: hand-edits OK, but keep section anchors (`## Tool commands`, `## Hooks`, `## Stacks`, `## Frameworks & libraries`, `## Regions`, `## CI`). `last_reconciled` in frontmatter bumps on every meaningful change.
- `code-debt.md` / `charter-debt.md`: edit during `debt-review` / `audit`. Each entry names the trigger, the impact, and the next step.
- `INSTALL_PROFILE.md`: append-only. `keystone init` re-runs are additive; do not strip prior selections.


## Anti-patterns

- Mutating `CODEBASE_STATE.md` from a hook without a visible diff.
- Removing section headers when editing state files — downstream hooks and the agent's section-anchored reads break.
- Hand-editing `INDEX.json` or `GLOBS_INDEX.md`'s `## Index` table.
