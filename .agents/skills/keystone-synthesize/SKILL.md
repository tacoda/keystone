---
name: keystone-synthesize
description: 'Promote accepted learning-inbox candidates into the right reference / rule / hook layer, then refresh INDEX.json. Triggers: keystone synthesize; keystone:synthesize; /keystone:synthesize; run synthesize; promote learning candidates; promote inbox to rules.'
allowed-tools: Read Write Edit Grep Glob Bash
metadata:
  model: opus
---

# keystone:synthesize — promote inbox candidates

The Learning flywheel's promotion step. Walks
`.keystone/learning/inbox/`, promotes `status: accepted`
candidates to the right layer (rule / reference / hook), regenerates
the globs index, and refreshes Cursor projections if present.

## Run

**Promote learning-inbox candidates into the right reference or rule layer.** Read [`.keystone/learning/README.md`](../../../.keystone/learning/README.md).

## Activities

For each file in `.keystone/learning/inbox/`:

1. **Read the candidate.** Note its `proposed-layer` and (if present) `proposed-globs:` frontmatter.
2. **Decide the destination:**
   - Universal principle → `.agents/references/principles/<name>.md` and `.agents/rules/principles/<name>.md`
   - Stack idiom → `.agents/references/idioms/<stack>/<name>.md` and `.agents/rules/idioms/<stack>/<name>.md`
   - Process rule → `.agents/rules/process/<name>.md`
   - Hook → `.agents/hooks/<name>.md`
3. **Author the reference and rule pair.** The reference explains the *why*; the rule states the directive.
4. **Propose `paths:` for the rule.** If the candidate carries `proposed-globs:`, use those as the starting point. Otherwise, infer from the evidence: regional surprise → narrow globs that match the touched region (consult the region map in `.agents/references/state/CODEBASE_STATE.md`); cross-cutting surprise → no `paths:`. Always show the proposed `paths:` to the user as part of the rule-write diff. It is part of the rule, not a side detail.
5. **Choose the rule tier.** Default: regular rule. Iron law / golden rule only when deviation is genuinely non-negotiable.
6. **Move the candidate** from `inbox/` to `promoted/` (kept), or `rejected/` (with a one-line reason in the moved file).
7. **Regenerate `.agents/references/state/GLOBS_INDEX.md`** if any rule was written with `paths:`. Walk every rule, invert the `paths:`, replace the `## Index` table (same procedure as **bootstrap** step 6).
8. **Regenerate host projections.** Run `keystone index && keystone project`. The project step rewrites each host's rule files, such as `.cursor/rules/*.mdc`, and removes the ones whose rule is gone.

## Iron law

**No silent overwrites.** If a destination already exists, propose a diff and let the user merge.

## Gate

Synthesize is where the charter changes shape. Show the user every promotion and removal before applying. The `paths:` value on every promoted rule is part of that diff. Never silently default it. After synthesize writes to `rules/` or `GLOBS_INDEX.md`, the active session has stale rules in context. Recommend a context reset.

## Index refresh

After every promotion / removal, run `keystone index` so `.keystone/INDEX.json` reflects the new layout. If skills, subagents, or commands changed (rare for synthesize, possible for cross-cutting promotions), also run `keystone project` so host projections track.

## When to trigger

- After accepting candidates from `/keystone:learn`.
- Periodically (weekly) as part of charter hygiene.

## Iron law

No silent overwrites. Propose every diff before applying.
