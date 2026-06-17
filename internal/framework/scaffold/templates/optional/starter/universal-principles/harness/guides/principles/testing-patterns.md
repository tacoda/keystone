# Testing Patterns — rules

The non-negotiable and strongly-preferred rules from [`corpus/principles/testing-patterns.md`](../../corpus/principles/testing-patterns.md).
Loaded ambient; enforced by the [drift sensor](../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAWS

**Real objects in, real boundaries out.** Use real instances of every collaborator inside the system under test. Replace only the things that cross out of the process. A suite that violates this rule is testing its mocks, not its code, and will eventually pass a build that does not work.

**Flaky tests are not allowed.** A test whose result depends on timing, ordering, environment, or any other source of hidden non-determinism is not a green test — it is a future false signal. Fix the non-determinism (control the clock, control RNGs, control test ordering, isolate fixtures) or delete the test. Marking a test "flaky" and retrying it is forbidden — the retry hides the failure that the suite exists to surface.

## GOLDEN PATH

- **Test quality is the ideal — not coverage, not type passage.** A good test names a real use case or behavior, exercises it Chicago-style with real collaborators (see the IRON LAW), and fails meaningfully when that behavior breaks. Coverage percentage and a green type-checker are byproducts, not goals: a 100%-covered suite of tautology tests is worse than a 60%-covered suite of behavior tests. When evaluating tests, ask what user-visible or use-case-level fact this test would protect — not how many lines it touches.
- **Aim for tests that survive refactoring.** A test that breaks when the implementation changes but the behavior does not is fighting you — see [[refactoring]].
- **Aim for tests that fail loudly and locally.** The failure message should name the behavior that broke and point at the cause; this is what unit-level tests buy you. See [[fail-fast]].
- **Aim for a fast suite.** Slow tests get skipped; skipped tests rot; rotted tests get deleted. The pyramid in [[tdd]] is what keeps the suite fast — many small, fast tests at the base.
- **Aim for tests as readable as the code they test.** A reader should be able to learn the system from its tests.

---

Traces to: [`corpus/principles/testing-patterns.md`](../../corpus/principles/testing-patterns.md).
