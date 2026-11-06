---
name: keystone-debt-review
description: 'Triage the code-debt ledger — review the code-debt findings and update .agents/references/state/code-debt.md. Triggers: keystone debt-review; keystone:debt-review; /keystone:debt-review; review code debt; triage debt ledger; debt review.'
allowed-tools: Read Grep Glob Bash Write Edit
metadata:
  model: opus
---

# keystone:debt-review — triage the code-debt ledger

Reads the code-debt subagent's latest findings and reconciles them against
`.agents/references/state/code-debt.md`. Each finding gets one of:
keep (still real), close (resolved), defer (not now), or escalate
(promote to charter-debt or a rule).

Subagent
contract at `.agents/agents/code-debt.md`.

## Run

**Triage the code-debt ledger.** Read [`.agents/agents/code-debt.md`](../../agents/code-debt.md) and [`.agents/references/state/code-debt.md`](../../references/state/code-debt.md).

For **charter debt** (stale rules, placeholder bootstrap, empty idiom dirs, uncited policies), use the [`audit`](../keystone-audit/SKILL.md) command. Its Pruning flywheel writes to `.agents/references/state/charter-debt.md`.

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

## When to trigger

- Weekly or biweekly cadence.
- Before a major refactor — clears the slate.
- After incident response — fresh findings need triage before they rot.
