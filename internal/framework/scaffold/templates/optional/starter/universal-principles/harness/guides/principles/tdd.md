# Test-Driven Development — rules

The non-negotiable and strongly-preferred rules from [`corpus/principles/tdd.md`](../../corpus/principles/tdd.md).
Loaded ambient; enforced by the [drift sensor](../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**Never write production code without a failing test for it.** The exceptions are rare and identifiable: prototypes you will throw away, exploratory spikes meant to learn (which are then thrown away or rewritten under tests), and trivial declarative changes (renames, comments, formatting). Everything else moves through Red-Green-Refactor or it is not under the discipline.

## GOLDEN PATH

- **Aim for cycles measured in minutes, not hours.** A long red phase is a hint that the step is too big — split it.
- **Aim to watch the test fail before making it pass.** The watching is what proves the test is wired to the code. Skip it and you will eventually ship a test that passes whether the code works or not.
- **Aim for tests that pin behavior, not structure.** A test that breaks every time the implementation changes is a test that fights refactoring. Tests should change when behavior changes, not when shape changes. See [[refactoring]].
- **Aim for the pyramid, not the ice cream cone.** When end-to-end tests dominate, the cure is more unit tests under the high-value paths, not more end-to-end tests.

---

Traces to: [`corpus/principles/tdd.md`](../../corpus/principles/tdd.md).
