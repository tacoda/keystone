---
kind: rule
id: legacy/tests-first-per-region
description: 'Regions the kit touches gain tests by expansion (targeted, passing, landed) — never by rewrite.'
paths:
  - "**"
---
# Tests first, per region

Regions gain tests by expansion, not by the rewrite. The test-expansion
documents track coverage per region: targeted (named, with the kind of
test), passing (run from a clean clone), landed (the map cites them).
The expansion rides the region's migration when one exists; otherwise it
rides nothing — and runs after characterization, never before.
