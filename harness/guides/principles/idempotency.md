# Idempotency — rules

The non-negotiable and strongly-preferred rules from [`corpus/principles/idempotency.md`](../../corpus/principles/idempotency.md).
Loaded ambient; enforced by the [drift sensor](../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**Every mutation across a network must be idempotent — by design, by key, or by reconciliation.** The network is unreliable; retries are unavoidable; non-idempotent retries are a defect. If an operation cannot be made idempotent, the architecture has to absorb the cost — exactly-once protocols, two-phase commits, manual reconciliation — and pay it explicitly, not silently.

## GOLDEN RULES

- **Aim for idempotent operations by construction.** "Set X to value V" is idempotent; "increment X by 1" is not. Where the design admits a choice, prefer the idempotent shape.
- **Aim for client-generated idempotency keys at every retry boundary.** Server-generated keys cannot deduplicate retries — by the time the server generates one, it has already done the work.
- **Aim for the deduplication record and the operation to commit together.** Anything less is a race waiting to be observed.
- **Aim to bound retry behavior.** Exponential backoff, jitter, circuit breakers — without them, a flaky downstream becomes a self-inflicted outage. See [[distributed-systems-fallacies]].

---

Traces to: [`corpus/principles/idempotency.md`](../../corpus/principles/idempotency.md).
