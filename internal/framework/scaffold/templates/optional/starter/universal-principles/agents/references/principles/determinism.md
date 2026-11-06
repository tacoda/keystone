# Determinism

A deterministic test gives the same answer every time given the same input. A non-deterministic test gives different answers under different ambient conditions — and is therefore not a test of the code, but a test of *the code plus the conditions*. Flaky tests are the most common operational cost of a non-deterministic suite, but the deeper cost is loss of trust: once one test is known to flake, every red is suspect.

> **Rules extracted:** [`guides/principles/determinism.md`](../../guides/principles/determinism.md).

## What it asks of you

- Treat time, randomness, ordering, the network, the filesystem, and the environment as *inputs*, not as ambient state. Inputs are passed; ambient state is read. Tests pass inputs.
- Make every source of variation injectable. The production code uses a real clock; the test uses a frozen one. Same code, different input.
- Quarantine the rest. Tests that genuinely need the real world live in a separate, longer-running, flake-tolerant tier — typically called integration or end-to-end. They are not the main suite, and the main suite does not depend on them.

## Why it holds

Google's 2016 flaky-test study (Memon, Gao, Nguyen, et al.) measured the cost: roughly 1 in 7 test failures across their internal codebase were flakes, costing engineering hours, masking real regressions, and gradually training engineers to retry-and-merge. The fix was systematic — detect flakes, quarantine them, repair the determinism issue, not the symptom.

The clock argument is older. *Working Effectively with Legacy Code* (Feathers, 2004) treats time as the canonical example of an "unmanaged dependency" — code that uses `now()` directly is harder to test than code that takes a clock parameter. The same applies to randomness, IDs, and the network.

The "test pyramid" (Cohn, 2009) is a determinism argument as much as a speed argument: unit tests sit at the base because they are *isolated*, which is what makes them fast *and* reliable. Integration tests are slower partly because they cross more boundaries, but mostly because they admit more sources of variation.

## Anti-patterns

- `await new Promise(r => setTimeout(r, 100))` in a test, hoping that's long enough.
- `assert response.created_at == datetime.now()` — racing the assertion against the system clock.
- `Math.random() > 0.5` inside the code under test, with no seed control.
- A test that passes locally and fails in CI, "fixed" by adding `pytest.mark.flaky`.
- Retries in the runner as a substitute for fixing the cause. Retries hide the bug, they do not solve it.
- A test that depends on another test having run first.

## References

- Memon, A. et al. (2017). *Taming Google-Scale Continuous Testing* — ICSE.
- Feathers, M. *Working Effectively with Legacy Code* (2004).
- Cohn, M. *Succeeding with Agile* (2009) — the test pyramid.
- Fowler, M. [*Eradicating Non-Determinism in Tests*](https://martinfowler.com/articles/nonDeterminism.html).
- Beck, K. *Test-Driven Development: By Example* (2002) — F.I.R.S.T., specifically the *Repeatable* property.

---

Forward link: [`guides/principles/determinism.md`](../../guides/principles/determinism.md). See also: [`testing-patterns.md`](testing-patterns.md), [`tdd.md`](tdd.md).
