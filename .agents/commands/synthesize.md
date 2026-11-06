---
kind: command
id: synthesize
description: 'Promote learning-inbox candidates into the right corpus or guide layer.'
---
# synthesize

**Promote learning-inbox candidates into the right corpus or guide layer.** Read [`.charter/learning/README.md`](learning/README.md).

## Activities

For each file in `.charter/learning/inbox/`:

1. **Read the candidate.** Note its `proposed-layer` and (if present) `proposed-globs:` frontmatter.
2. **Decide the destination:**
   - Universal principle → `.charter/corpus/principles/<name>.md` and `.charter/guides/principles/<name>.md`
   - Stack idiom → `.charter/corpus/idioms/<stack>/<name>.md` and `.charter/guides/idioms/<stack>/<name>.md`
   - Process rule → `.charter/guides/process/<name>.md`
   - Sensor → `.charter/sensors/<name>.md`
3. **Author the corpus and guide pair.** Corpus explains the *why*; guide states the rule.
4. **Propose `globs:` for the guide.** If the candidate carries `proposed-globs:`, use those as the starting point. Otherwise, infer from the evidence: regional surprise → narrow globs that match the touched region (consult `corpus/state/CODEBASE_STATE.md`'s region map); cross-cutting surprise → no `globs:`. Always show the proposed `globs:` to the user as part of the guide-write diff — it is part of the rule, not a side detail.
5. **Choose the rule tier.** Default: regular rule. Iron law / golden rule only when deviation is genuinely non-negotiable.
6. **Move the candidate** from `inbox/` to `promoted/` (kept), or `rejected/` (with a one-line reason in the moved file).
7. **Regenerate `.charter/corpus/state/GLOBS_INDEX.md`** if any guide was written with `globs:`. Walk every guide, invert the globs, replace the `## Index` table (same procedure as **bootstrap** step 6).
8. **Regenerate Cursor projections** if `.cursor/rules/` exists. For each newly promoted guide that declares `globs:`, write `.cursor/rules/keystone-<topic>-<name>.mdc` mirroring the guide's globs and pointing at the source guide. For any guide whose `globs:` changed during this synthesize, rewrite the corresponding `.mdc`. Delete any `.cursor/rules/keystone-*.mdc` whose guide no longer exists or no longer declares `globs:`.

## Iron law

**No silent overwrites.** If a destination already exists, propose a diff and let the user merge.

## Gate

Synthesize is where the charter changes shape. Show the user every promotion and removal before applying. The `globs:` value on every promoted guide is part of that diff — never silently default it. After synthesize writes to `guides/` or `GLOBS_INDEX.md`, the active session has stale rules in context — recommend a context reset.

## Index refresh

After every promotion / removal, run `keystone index` so `.keystone/INDEX.json` reflects the new layout. If skills, subagents, or commands changed (rare for synthesize, possible for cross-cutting promotions), also run `keystone project` so host projections track.
