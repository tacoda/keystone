# Testing Patterns

How to build a test suite that stays useful for years. [[tdd]] states the loop; [[bdd]] states the vocabulary; this file states the **patterns** that distinguish a suite which is an asset from one which is a maintenance tax.

The central commitment of this file is **Chicago-style** (classicist) testing: prefer real objects over test doubles wherever possible, and **mock only at the actual boundaries** of your system. The arguments for this position were made by Beck in the original TDD practice and articulated in detail by Martin Fowler in "Mocks Aren't Stubs" (2007); the contrasting London school is described in [[tdd]].

> **Rules extracted:** [`guides/principles/testing-patterns.md`](../../guides/principles/testing-patterns.md). This file holds the full reasoning, anti-patterns, and references.

## Mock only at actual boundaries

The rule, stated plainly:

> If you did not write it, and it crosses the boundary of your process, mock it. Otherwise, use the real thing.

"Actual boundaries" are the seams between your code and the outside world — places where determinism, speed, cost, or availability would otherwise compromise the test:

- **Network calls** to services you do not own (third-party APIs, payment providers, email/SMS).
- **The filesystem,** when it would slow tests or require fixtures the suite cannot manage.
- **The database,** *sometimes* — a fast in-process database (SQLite, an in-memory Postgres, testcontainers) is often closer to "real" than a mock and is preferred when feasible.
- **Time and randomness** — wall-clock reads, `now()`, UUID generation, RNGs. Hidden non-determinism is the largest single source of flaky tests.
- **The operating system** — process spawning, signals, environment.

Everything *inside* the boundary — your own services, your own classes, your own functions — should be tested with the **real implementations** wired together. The test you write is then a test of the **interaction** between collaborators, not a test of "did I call my mock correctly?"

Why this matters:

- **Mocks lie.** A mock specifies the return value the test author expected the collaborator to produce — not necessarily the value it actually produces in production. Refactor the collaborator, and the test that mocks it stays green even when the system is broken.
- **Mocks couple tests to structure.** A test that asserts *which methods were called in what order* breaks every time the design changes, even when the behavior does not. The test fights [[refactoring]] instead of supporting it.
- **Mocks hide design problems.** When a class needs five mocks to be testable, the design is telling you it has five things wrong with it. Mocking around the problem silences the messenger.

The mock-everything style produces a suite where every test passes individually but the system does not work. The mock-only-at-boundaries style produces a suite where the unit tests are tests of real behavior and the integration tests serve their actual purpose.

## The test-double taxonomy (Meszaros / Fowler)

Vocabulary matters. The word "mock" is overloaded; the categories are not interchangeable.

- **Dummy** — passed but never used. A placeholder to satisfy a parameter list.
- **Fake** — a working implementation, simpler than production. The classic example: an in-memory implementation of a repository interface.
- **Stub** — returns canned answers to calls. No verification.
- **Spy** — a stub that also records how it was called, for later inspection.
- **Mock** — pre-programmed with expectations: which calls it must receive, in what order, with what arguments. Fails the test if expectations are not met.

**Prefer fakes to mocks.** A fake exercises the real shape of the collaborator's behavior; a mock asserts on calls. When you must use a double, a fake is almost always the right one. Mocks (in the strict sense — pre-programmed expectations) are the heaviest tool and the one with the most leverage to make tests brittle.

## Arrange, Act, Assert

The structural counterpart to Given-When-Then (see [[bdd]]) for plain unit tests:

- **Arrange** — set up the situation under test.
- **Act** — perform exactly one operation.
- **Assert** — verify the observable outcome.

A test with more than one Act is more than one test. A test with no Arrange is a test of a singleton — usually a smell. A test whose Arrange is fifty lines is a test of code that is too hard to construct, which is the test giving you design feedback — see [[tdd]].

## Test naming as specification

Test names form the readable spec of the system. The shape that holds up:

```
<behavior>_<situation>
```

- *calculates_tax_for_orders_over_a_hundred_dollars*
- *rejects_signup_when_email_is_already_registered*
- *retries_three_times_then_returns_an_error_when_the_upstream_is_unreachable*

Avoid: *test1*, *test_method*, *test_happy_path*, *test_X_works*, *test_setUserPaymentMethod_succeeds*. None of these specify a behavior.

When the test names are listed alphabetically, they should read like a table of contents for the system's behavior. If they do not, the system does not have a spec — it has a suite.

## One logical assertion per test

A test should fail for one reason. Asserting on five unrelated things in a single test produces a failure message that requires investigation to interpret. The constraint is on **logical** assertions, not the count of `assert` statements — asserting that an object has a name, an id, and a created timestamp is one logical assertion ("the object was constructed correctly") even if it takes three statements.

When you find yourself splitting a single test into two because of the rule, you have usually found a missing behavior — name it.

## Build a test data builder

The Test Data Builder pattern (Pryce & Freeman) — small, fluent factories that produce valid domain objects with sensible defaults — is the single highest-leverage tool for making a suite readable. The contrast:

```
// Without
Customer c = new Customer("test@example.com", "Test", "User",
    Address.of("123 Main", "Anytown", "CA", "94000"),
    null, false, Locale.US, ZonedDateTime.now(), ...);

// With
Customer c = aCustomer().build();
Customer locked = aCustomer().locked().build();
Customer canadian = aCustomer().in(Country.CA).build();
```

Every test that needs *just any customer* uses the first form; every test that needs *a particular kind of customer* says so in one line. The suite reads as specifications about the variations that matter, not as construction code.

## Characterization tests for code without tests

When you must change code that has no tests — the Feathers case — the discipline is:

1. Write tests that pin the **current** behavior, whatever it is. Even bug-for-bug. These are *characterization tests* — they characterize what the code does, not what it should do.
2. Refactor under their protection. See [[refactoring]].
3. Now that the code is tested and clean, fix the behavior if needed, with new tests for the new behavior.

Characterization tests are temporary scaffolding. They keep the system honest while you make it changeable.

## Property-based tests at the right seams

Articulated by Claessen & Hughes in "QuickCheck" (2000): rather than (or alongside) example-based tests, state a **property** the code should satisfy for every input and let the framework generate inputs to try. Properties find edge cases example-based tests routinely miss — empty inputs, boundary values, Unicode, integer overflow, ordering invariants.

Most useful at the seams where input space is large and properties are clear — parsers, serializers, sort, idempotent operations, round-trips. Less useful for business workflows where the property *is* a sequence of specific examples.

## What it asks of you

- When you write a test, ask whether the doubles in it are at real boundaries. If the doubles are around your own code, replace them with real instances.
- When a class is hard to test, do not reach for more mocks; reach for a better design. The pain is feedback. See [[tdd]] (test-first as design).
- When a test breaks every time you refactor, the test is asserting on structure, not behavior. Rewrite it to assert on observable outcomes.
- When the same setup appears in five tests, extract a test data builder. The duplication is the spec asking to be named.
- When you discover a bug in untested code, write a characterization test for the bug, *then* fix it. The test pins the regression you do not want repeated.
- When a test is flaky, find the source — hidden time, hidden ordering, hidden network — and remove it. Disabling flaky tests is silencing the smoke alarm.

## Anti-patterns

- Mocking your own classes. The doubles are arguing with your own architecture.
- Asserting on method-call order with a strict mock framework. The test is now an enemy of refactoring.
- Tests that share state through class-level fixtures, so the order of execution matters. They will be order-dependent until removed.
- Tests with branches and loops. The test then needs its own tests; the bug will be in the test, not the code.
- A "unit test" that takes 800ms because it boots a container. The pyramid is collapsing into a cone.
- Sleeps in tests to "wait for things to settle." The non-determinism is still there; you just hid it. See [[fail-fast]].
- Tests that pass when commented out. The assertion is missing, weak, or testing the wrong thing.
- A suite that requires `--retry` on the CI runner to be green. Each retry is a coin flip with the truth.
- Deleting a flaky test instead of diagnosing it. The flake was a clue.

## References

- Beck, K. (2002). *Test-Driven Development: By Example*. Addison-Wesley. (The original classicist practice.)
- Fowler, M. (2007). "Mocks Aren't Stubs." martinfowler.com/articles/mocksArentStubs.html. (The definitive comparison of classicist and mockist schools; the source of the modern test-double taxonomy in practice.)
- Meszaros, G. (2007). *xUnit Test Patterns: Refactoring Test Code*. Addison-Wesley. (The catalog of test smells and the test-double vocabulary.)
- Freeman, S., & Pryce, N. (2009). *Growing Object-Oriented Software, Guided by Tests*. Addison-Wesley. (The London school; the Test Data Builder pattern.)
- Feathers, M. (2004). *Working Effectively with Legacy Code*. Prentice Hall. (Characterization tests; seams.)
- Claessen, K., & Hughes, J. (2000). "QuickCheck: A Lightweight Tool for Random Testing of Haskell Programs." *ICFP '00*. (Property-based testing.)
- Hevery, M. (2009). "Guide: Writing Testable Code." misko.hevery.com. (Practical guidance on how design choices make code testable or untestable — the design-feedback half of [[tdd]].)
