# Determinism — rules

The non-negotiable and strongly-preferred rules from [`corpus/principles/determinism.md`](../../corpus/principles/determinism.md). Loaded ambient; enforced at implementation and verification.

## GOLDEN RULE

- **Aim to treat a non-deterministic test as not-a-test.** A test whose outcome depends on time, randomness, ordering, the network, the filesystem, or any other ambient state must control that state explicitly. If it cannot, it does not belong in the test suite — it belongs in a separate, gated, flake-tolerant tier. (Builds on the existing testing IRON LAW *"flaky tests are not allowed."*)
- **Aim to inject the clock.** `now()` is a dependency; let the test pass in a fixed value. Equivalents: `freeze_time`, `MockClock`, `time.tzset`, `sinon.useFakeTimers`.
- **Aim to seed every randomness source.** `random.seed(0)`, `np.random.default_rng(0)`, `Math.random` replaced with a seeded PRNG in tests. Bootstrap records the project's pattern.
- **Aim to control IDs and tokens.** UUIDs, ULIDs, request IDs, JWTs — generated through an injectable source so tests can pin them.
- **Aim to remove the network from unit tests.** If a test makes an outbound call to a server you do not control, it is an integration test (or a flake).

## RULES

- **No `sleep()` in tests.** Use explicit synchronization (event, condition variable, polling with a bounded timeout and a *reason*).
- **No real wall-clock arithmetic in tests.** `expected = now + duration` is fragile; freeze the clock.
- **No reliance on dict / map / set iteration order** unless the language guarantees it (Python 3.7+, JavaScript objects since ES2015 for string keys). Even then, prefer explicit sorting in tests.
- **No reliance on filesystem ordering.** `os.listdir` order is platform-specific.
- **No reliance on locale.** Date formatting, sort order, number formatting — pin or avoid in test assertions.
- **No environment leakage.** A test that sets `os.environ['X']` cleans up. A test that reads it from the surrounding shell is environment-dependent.
- **No reused test data between cases.** Each test seeds its own state.

## Sensors

- The **test sensor** flags tests that have been retried within the same run as flaky candidates (when the runner supports retry annotations).
- **Bootstrap** records the time-mocking library, the seeded-random pattern, and the test-isolation library (`pytest-freezer`, `vitest.useFakeTimers`, `quicktime`, etc.) in `corpus/state/CODEBASE_STATE.md`.

---

Traces to: [`corpus/principles/determinism.md`](../../corpus/principles/determinism.md). See also: [[testing-patterns]], [[tdd]], [[ci-failure]].
