# Test-Driven Development

Write a failing test, make it pass, then improve the code — repeat. Articulated by Kent Beck in *Test-Driven Development: By Example* (2002), drawing on practice from the Smalltalk and Extreme Programming communities. Beck's central, often-missed claim: **TDD is primarily a design discipline, not a testing one.** The tests are a byproduct; the value is the pressure the act of writing tests first puts on the design.

> **Rules extracted:** [`guides/principles/tdd.md`](../../guides/principles/tdd.md). This file holds the full reasoning, anti-patterns, and references.

## The loop

**Red → Green → Refactor.** Three phases, in order, every cycle:

1. **Red.** Write a small test for behavior that does not yet exist. Run it. Watch it fail. The failure proves the test is connected to the code — a test that has never failed is a test that does not test.
2. **Green.** Write the smallest change that makes the test pass. Not the *right* change — the *passing* change. Hardcoded returns, copy-pasted blocks, obvious fakery — all are legal in the green step. The discipline is to *get to green*, not to design en route.
3. **Refactor.** Now that the test is green, clean up. The test pins the behavior; the refactor changes the shape. See [[refactoring]] — this is the third step's natural home. Stay in refactor until the design tells you the next test, then start the next loop.

The cycles are **small**. Beck's working pace in the book is one cycle in two-to-five minutes. Cycles that take an hour are not TDD; they are "writing tests first, occasionally."

## Test-first as design feedback

The deeper claim: the act of writing a test *before* the code reveals design problems that would be invisible if the code came first. If the test is hard to write, the code under test is probably:

- Hard to construct (too many dependencies — see [[SOLID]] DIP, [[coupling-cohesion]])
- Hard to set up (hidden state, untestable singletons)
- Hard to observe (no return values; side effects through global channels)
- Hard to isolate (does too many things — see [[SOLID]] SRP, [[object-oriented-design]])

Listening to *test pain* as feedback on the *design* is the TDD discipline. The wrong response to a hard test is to write a complicated test; the right response is to reshape the code until the test is easy.

## The test pyramid

Articulated by Mike Cohn in *Succeeding with Agile* (2009) and refined by Fowler: a healthy test suite is shaped like a pyramid.

- **Base — many unit tests.** Fast (milliseconds), isolated, hit one unit of behavior. The bulk of the suite. Failures point precisely at the cause.
- **Middle — fewer integration tests.** Exercise units in combination, with real (or close-to-real) collaborators where the seams matter. Slower but still tractable.
- **Top — very few end-to-end tests.** Exercise the system through its outermost interface. Slow, flaky-prone, expensive to write and maintain. Reserved for the handful of journeys that absolutely must work.

The inverse — many end-to-end, few unit — is the **ice cream cone** anti-pattern: slow feedback, brittle suite, failures that point at a symptom three layers from the cause. See [[modern-software-engineering]] on short feedback loops; the pyramid is what keeps the loop short.

## F.I.R.S.T. — properties of good tests

Robert C. Martin's articulation (*Clean Code*, 2008). A good test is:

- **Fast** — runs in milliseconds, so the suite can run on every change.
- **Independent** — does not depend on other tests' state or ordering.
- **Repeatable** — produces the same result every time, on every machine, regardless of clock or network.
- **Self-validating** — passes or fails with no human interpretation. No "look at the output and decide."
- **Timely** — written *just before* the code that makes them pass, not weeks later.

A test that fails any of these will be skipped, ignored, or deleted within a release cycle.

## Outside-in vs. inside-out

Two schools, both valid, neither universal.

**Outside-in** (Freeman & Pryce, *Growing Object-Oriented Software, Guided by Tests*, 2009): start the cycle from the outermost behavior the user cares about; let the tests for inner collaborators emerge as you discover, in the green step, that you need them. Sometimes called the *London school* of TDD. Tends to use test doubles heavily at boundaries; well-suited to systems where the architecture is the design problem.

**Inside-out** (the *Detroit school*, after Beck's original Smalltalk practice): start from the smallest computation you understand and grow outward. Fewer test doubles, more real objects, tests closer to the units of behavior. Well-suited when the domain logic is the design problem. See [[testing-patterns]] for the practical discipline that descended from this lineage — *Chicago-style* classicist testing, with mocking only at real boundaries.

Most real practice is a blend. The argument is less important than the *rhythm*: small steps, fast feedback, design pressure from the tests.

## What it asks of you

- When you are about to write code, stop. Write the test first. If you cannot write the test, you do not yet understand what you are building. See [[modern-software-engineering]] (falsifiability).
- When a test is painful to write, treat the pain as a design signal. Do not "make the test work"; reshape the code until the test is easy.
- When you are tempted to skip the refactor step, that is the step paying for all the others. Skipping it is how a test-first codebase turns into a tangle of just-barely-passing code.
- When a test fails in a way you did not expect, the test has done its job. Read carefully before "fixing" either side.
- When you find a bug, write the failing test that reproduces it *first*. Then fix it. The test is the only thing that prevents the bug from coming back.

## Anti-patterns

- Writing tests *after* the code, then calling it TDD. The design pressure is lost; the tests merely lock in whatever shape happened to emerge.
- Tests that mock the system under test, leaving the test asserting on the mock rather than the behavior.
- Skipping the refactor step "because the test passes." The accumulating mess is the price; refactoring is how the price gets paid.
- A test suite where one test's failure cascades into ten others (violates Independent).
- A test that calls `sleep(...)` to "wait for it to settle" (violates Fast and Repeatable).
- A test whose assertion is "no exception was thrown" — passing whether or not the code did anything (violates Self-validating).
- An "integration test" that exercises the whole stack to verify one branch of one function. The unit test wanted to be written; the integration test is paying for the missing unit.
- Deleting a flaky test rather than diagnosing it. The flake was telling you something; deletion silences the messenger.

## References

- Beck, K. (2002). *Test-Driven Development: By Example*. Addison-Wesley. (The canonical text.)
- Freeman, S., & Pryce, N. (2009). *Growing Object-Oriented Software, Guided by Tests*. Addison-Wesley. (The outside-in / London school; mock objects as a design tool.)
- Cohn, M. (2009). *Succeeding with Agile: Software Development Using Scrum*. Addison-Wesley. (Origin of the test pyramid.)
- Fowler, M. (2012). "TestPyramid." martinfowler.com/bliki/TestPyramid.html.
- Martin, R. C. (2008). *Clean Code: A Handbook of Agile Software Craftsmanship*. Prentice Hall. (Chapter 9 — F.I.R.S.T.)
- Meszaros, G. (2007). *xUnit Test Patterns: Refactoring Test Code*. Addison-Wesley. (The catalog of test smells and remedies.)
- Feathers, M. (2004). *Working Effectively with Legacy Code*. Prentice Hall. (Bringing existing code under test as a precondition for safe change — see [[refactoring]].)
