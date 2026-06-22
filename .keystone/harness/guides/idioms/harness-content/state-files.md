---
kind: guide
id: guides/idioms/harness-content/state-files
description: 'Authoring rules for harness/corpus/state/.'
globs:
  - ".keystone/harness/corpus/state/**/*.md"
  - "internal/framework/templates/**/corpus/state/**/*.md"
---
# State files — rules

The rules from [`corpus/idioms/harness-content/state-files.md`](../../../corpus/idioms/harness-content/state-files.md).

## IRON LAW

- **No silent overwrites.** Every state-file write goes through the agent's edit primitive so the user sees the diff. Sensors that update state must surface their diff for review.
- **Do not hand-edit derived sections.** The `## Index` table in `GLOBS_INDEX.md` is regenerated. `quality-radar.md`, `risk-fingerprints.md`, `traffic-topology.md` are sensor snapshots — replace whole-file via the sensor, never patch.

## GOLDEN RULE

- `CODEBASE_STATE.md`: hand-edits OK, but keep section anchors (`## Tool commands`, `## Sensors`, `## Stacks`, `## Frameworks & libraries`, `## Regions`, `## CI`). `last_reconciled` in frontmatter bumps on every meaningful change.
- `code-debt.md` / `harness-debt.md`: edit during `debt-review` / `audit`. Each entry names the trigger, the impact, and the next step.
- `INSTALL_PROFILE.md`: append-only. `keystone init` re-runs are additive; do not strip prior selections.

## Anti-patterns

- Mutating `CODEBASE_STATE.md` from a sensor without a visible diff.
- Removing section headers when editing state files — downstream sensors and the agent's section-anchored reads break.
- Hand-editing `INDEX.json` or `GLOBS_INDEX.md`'s `## Index` table.
