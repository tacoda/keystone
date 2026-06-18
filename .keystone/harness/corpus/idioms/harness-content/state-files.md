---
kind: corpus
id: corpus/idioms/harness-content/state-files
description: 'How state files are written and what must not be hand-edited.'
---
# State files

State files under `harness/corpus/state/` are the empirical record of the project: what stacks exist, what tools run, what guides claim which globs, what debt is outstanding. Some are agent-authored only; some are hand-edited freely; some are regenerated on every `keystone index` and must not be touched.

> **Rules extracted:** [`guides/idioms/harness-content/state-files.md`](../../../guides/idioms/harness-content/state-files.md).

## The state files

| File | Authored by | Hand-edit? |
|---|---|---|
| `CODEBASE_STATE.md` | bootstrap (initial), verify / learn / audit (updates) | yes, with care — section anchors must survive |
| `GLOBS_INDEX.md` | bootstrap / synthesize / audit (the `## Index` table is regenerated) | **no** on the `## Index` table; rest is template |
| `INSTALL_PROFILE.md` | `keystone init` | yes (re-runs are additive) |
| `code-debt.md` | code-debt sensor + debt-review action | yes during debt-review |
| `harness-debt.md` | harness-debt sensor + audit action | yes during audit |
| `quality-radar.md` | quality-radar sensor | no (snapshot of the most recent reviewed diff) |
| `risk-fingerprints.md` | risk-fingerprint sensor | no |
| `traffic-topology.md` | traffic-topology sensor | no |

## Iron law (no silent overwrites)

The **bootstrap** action propose-edits every state file through the agent's edit primitive — the user accepts or rejects each diff. Narration is not a write. A sensor that updates state must produce a diff the user reviews; in-place mutation without a visible diff is a contract violation.

## Index regeneration

`keystone index` reads every file under `harness/` (project layer) and `harness/plugins/*/` (vendored plugins), then writes `.keystone/INDEX.json`. Manual edits to INDEX.json are lost on the next run. The same rule applies to the `## Index` table in `GLOBS_INDEX.md`: it's a derived view, not a hand-authored doc.
