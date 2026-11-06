---
description: Run the drift sensor on the current diff.
---
# check-drift

**Run the drift sensor on the current diff.** Compare the in-progress changes against loaded charter rules. Fast pre-verify check. Read [`.agents/hooks/drift.md`](../hooks/drift.md) and [`.agents/rules/process/implementation.md`](../rules/process/implementation.md).

## Activities

1. **Identify the diff** — `git diff`, the staged changes, or the in-progress edits. The list of changed files is the **touched-files set** for this action.
2. **List which guides apply** — load `.agents/rules/idioms/<stack>/*.md` for each touched stack, plus `.agents/rules/process/implementation.md` and any guides from installed policys (`.keystone/policies/*/guides/`). For each candidate, if it declares `globs:` in frontmatter, keep it only when at least one touched file matches; guides without `globs:` keep today's loading.
3. **Compare diff to guides** — line by line, surface violations or hot drift signals. A guide's findings are reported only against files matching its `globs:` (or all touched files when no `globs:` is set).
4. **Report findings** — one bullet per drift. Cite the guide that was violated.

## When to invoke

Between **implementation** and **verify**. Faster and cheaper than the full verify pass; meant to catch obvious drifts before sensors run.

## Index freshness precondition

check-drift reads `.keystone/INDEX.json` to know which guides apply to the touched files. If the index is stale (any guide file modified more recently than `INDEX.json`), run `keystone index` first — otherwise the drift check uses an outdated set of rules.
