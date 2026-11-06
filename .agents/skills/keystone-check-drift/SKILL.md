---
name: keystone-check-drift
description: 'Run the drift hook on the current diff — fast pre-verify check that loaded charter rules still match what the code is doing. Triggers: keystone check-drift; keystone:check-drift; /keystone:check-drift; check drift; run drift hook; drift check.'
allowed-tools: Read Bash Grep Glob
metadata:
  model: sonnet
---

# keystone:check-drift — fast pre-verify drift check

Compares the in-progress diff against loaded charter rules. Fires
before the full computational hook suite. It catches the obvious case
where the agent is implementing something the rules explicitly
disallow.

Hook
contract at `.agents/hooks/drift.md`. Full discipline at
`.agents/rules/process/implementation.md`.

## Run

**Run the drift hook on the current diff.** Compare the in-progress changes against loaded charter rules. Fast pre-verify check. Read [`.agents/hooks/drift.md`](../../hooks/drift.md) and [`.agents/rules/process/implementation.md`](../../rules/process/implementation.md).

## Activities

1. **Identify the diff** — `git diff`, the staged changes, or the in-progress edits. The list of changed files is the **touched-files set** for this command.
2. **List which rules apply** — load `.agents/rules/idioms/<stack>/*.md` for each touched stack, plus `.agents/rules/process/implementation.md` and any rules from installed policies (`.keystone/policies/*/rules/`). If a candidate declares `paths:` in frontmatter, keep it only when at least one touched file matches. Rules without `paths:` always apply.
3. **Compare diff to rules** — line by line, surface violations or hot drift signals. Report a rule's findings only against files matching its `paths:`, or against all touched files when it has no `paths:`.
4. **Report findings** — one bullet per drift. Cite the rule that was violated.

## When to invoke

Between **implementation** and **verify**. Faster and cheaper than the full verify pass; meant to catch obvious drifts before hooks run.

## Index freshness precondition

check-drift reads `.keystone/INDEX.json` to know which rules apply to the touched files. If the index is stale (any rule file modified more recently than `INDEX.json`), run `keystone index` first. Otherwise the drift check uses an outdated set of rules.

## When to trigger

- Mid-implementation, before running `/keystone:verify`.
- After a large refactor that touched many regions — make sure the
  loaded idiom rules still apply.
- When `/keystone:verify` keeps failing on the same rule — drift is
  often the upstream cause.
