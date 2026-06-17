# The Fallacies of Distributed Computing — rules

The non-negotiable and strongly-preferred rules from [`corpus/principles/distributed-systems-fallacies.md`](../../corpus/principles/distributed-systems-fallacies.md).
Loaded ambient; enforced by the [drift sensor](../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**Every network operation is fallible and slow.** "Fallible" means *every* call has three outcomes — success, failure, and *unknown* (the request may or may not have happened). "Slow" means orders of magnitude slower than the equivalent local operation. Code that treats network calls as local calls is wrong, regardless of how the test suite behaves.

## GOLDEN PATH

- **Aim for timeouts on every call.** No exceptions. An untimed call is a vector for cascading failure.
- **Aim for idempotency at every retry seam.** Retries without idempotency convert flaky networks into duplicated state changes. See [[idempotency]].
- **Aim for failure modes that are observable.** When a call fails, the next operator needs to know *which* call, *to which peer*, *with what context*. See [[observability]].
- **Aim for degraded modes, not all-or-nothing.** A system that returns a stale but valid answer when a dependency is down beats one that returns 500.

---

Traces to: [`corpus/principles/distributed-systems-fallacies.md`](../../corpus/principles/distributed-systems-fallacies.md).
