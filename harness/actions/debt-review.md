# debt-review

**Triage the code-debt ledger.** Read [`harness/sensors/code-debt.md`](../sensors/code-debt.md) and [`harness/corpus/state/code-debt.md`](../corpus/state/code-debt.md).

For **harness debt** (stale rules, placeholder bootstrap, empty idiom dirs, uncited policies), use the [`audit`](audit.md) action — its Pruning flywheel writes to `corpus/state/harness-debt.md`.

## Activities

1. **Run the code-debt sensor.** Walk the codebase for markers and complexity hotspots; produce the candidate list.
2. **Triage `discovery` items.** For each, propose: a category (deliberate / drift / shortcut), a severity, an owner, and a revisit trigger. If you can't propose all four, leave the item as `discovery` with a note explaining why.
3. **Re-score existing items.** Compare ledger entries against the code. Items whose region is no longer present, or whose severity has dropped to nothing the diff or plan would notice, become `stale`.
4. **Prune `stale` items.** Delete them from the ledger. Do not archive.
5. **Propose the diff** to `corpus/state/code-debt.md` and let the user accept or edit. No silent writes.

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
