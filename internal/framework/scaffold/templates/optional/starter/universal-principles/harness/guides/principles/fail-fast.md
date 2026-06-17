# Fail Fast — rules

The non-negotiable and strongly-preferred rules from [`corpus/principles/fail-fast.md`](../../corpus/principles/fail-fast.md).
Loaded ambient; enforced by the [drift sensor](../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**Never continue past a violated invariant.** If your code detects that something it required to be true is false, it has two correct responses: refuse the operation, or terminate the process. Continuing is not one of them, regardless of how disruptive the alternative seems. The disruption is real; the silent corruption is worse.

## GOLDEN PATH

- **Aim to crash at startup, not at runtime.** Configuration errors, missing dependencies, unreachable services discovered at boot should prevent the process from accepting traffic. Failing under load is failing at the worst possible time.
- **Aim for diagnostics at the source.** A failure message should answer *what failed, where, and what was expected* — not just "operation failed."
- **Aim to distinguish expected from unexpected.** Expected errors (user input invalid, network blip) have handlers. Unexpected errors (invariant violated, "impossible" branch taken) have alarms.

---

Traces to: [`corpus/principles/fail-fast.md`](../../corpus/principles/fail-fast.md).
