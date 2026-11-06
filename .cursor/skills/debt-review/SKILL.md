---
name: debt-review
description: Triage the code-debt ledger.
disable-model-invocation: true
---

# debt-review

**Triage the code-debt ledger.** Read [`.agents/agents/code-debt.md`](../../../.agents/agents/code-debt.md) and [`.agents/references/state/code-debt.md`](../../../.agents/references/state/code-debt.md).

For **charter debt** (stale rules, placeholder bootstrap, empty idiom dirs, uncited policies), use the [`audit`](../../../.agents/commands/audit.md) command. Its Pruning flywheel writes to `.agents/references/state/charter-debt.md`.

## Activities

1. **Run the code-debt subagent.** Walk the codebase for markers and complexity hotspots; produce the candidate list.
2. **Triage `discovery` items.** For each, propose: a category (deliberate / drift / shortcut), a severity, an owner, and a revisit trigger. If you can't propose all four, leave the item as `discovery` with a note explaining why.
3. **Re-score existing items.** Compare ledger entries against the code. Items whose region is no longer present, or whose severity has dropped to nothing the diff or plan would notice, become `stale`.
4. **Prune `stale` items.** Delete them from the ledger. Do not archive.
5. **Propose the diff** to `.agents/references/state/code-debt.md` and let the user accept or edit. No silent writes.

## When to invoke

- Periodically (suggested: monthly or per audit pass).
- Before any large refactor of a region with multiple ledger entries — to make sure the plan accounts for them.
- When the ledger has accumulated more than ~10 `discovery` items.

## Output

One report:

- Items triaged (with proposed category/severity/owner/trigger).
- Items pruned as `stale`.
- Items re-scored.
- Open questions where triage stalled.
