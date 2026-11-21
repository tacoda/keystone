---
kind: document
id: test-expansion
description: Coverage expansion per region — targeted, passing, landed.
type: test-expansion
produced_by: plan
gates:
  - targeted
  - passing
  - landed
---

# test-expansion

Adding tests to a region that has none — or too few to trust. Work rides
the migration documents: test expansion is the smallest safe refactoring,
and characterization runs first (Feathers: characterization tests pin the
behavior as it is; seams make the untestable testable; sprout and wrap
keep the diff behavioral). The kit's rule set plants the discipline;
this document tracks the state.

## Gates

- **targeted** — the regions are named, with the coverage they have now
  and the kind of tests they need (characterization, seam, unit).
- **passing** — the tests run, from a clean clone, and pin the behavior
  they set out to pin.
- **landed** — the empirical map and the migration documents cite them.

## The sequence

1. **Characterize** the region (the characterization document runs first).
2. **Find the seams** — where the region is testable from the outside.
3. **Pin the behavior** with characterization tests; sprout new behavior
   out where the seams are absent.
4. **Unit-test the sprouts**; leave the pins standing until the seams
   hold the real behavior.
5. **Land**, and record the coverage delta in the ledger.

## Done when

- The region's coverage claim matches what the test machinery reports.
- No test changes production behavior: pins pin, sprouts sprout, and the
  characterization suite stays as-is while the code moves under it.
