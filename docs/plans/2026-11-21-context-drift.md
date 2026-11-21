---
status: draft
captured: 2026-11-21
topic: context drift (plan phase 5)
---

# Plan — context drift (`keystone drift`)

Drift is the gap between what the charter says and what the code is. The
charter is living context; keystone notices when it rots, names the stale
primitives, and lets the team refresh them. Keystone reports; agents and
people judge (decision 16: no runtime, no verdicts — warnings only).

## Three checks

1. **Dead links.** A relative link in a primitive body whose target
   doesn't exist. The rebase resolver (projection/links.go) already
   resolves targets from a linking file's folder; drift resolves and
   stats. URLs, anchors, and absolute paths are out of scope.
2. **Named paths gone.** A rule's `paths:` globs matching zero files:
   the region the rule governs vanished. Warning-level — a rule may run
   ahead of a rename in flight.
3. **Staleness.** A `verified:` date in frontmatter (author-set; drift
   never stamps); surfaced in `show`, INDEX, and the dashboard. Older
   than the staleness window (keystone.json `drift.days`, default 90) =
   warn. Bodies and windows never auto-bump; the author re-verifies.

## Where it surfaces

- `keystone drift [files…|--diff]` — the report, human + `--json`.
- `keystone verify` — stale-but-in-scope primitives as warnings.
- The dashboard — a drift pane fed by the same data.
- MCP — `keystone_drift_report`, descriptors only.

## Out of scope (decisions 12 and 16)

Contradiction-with-diff judgment (the existing drift hook keeps that),
symbol-level semantic checking, auto-stamping, auto-refresh.

## Acceptance criteria

- A synthetic fixture with one dead link, one vanished path, and one
  stale date reports exactly three warnings, named by file and reason.
- A clean tree reports zero. `verify` surfaces the stale one; `show`
  prints `verified:`; the pane and MCP read the same data.
