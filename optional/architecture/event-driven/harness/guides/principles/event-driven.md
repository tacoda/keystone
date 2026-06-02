# Event-Driven Architecture — rules

The rules from [`corpus/principles/event-driven.md`](../../corpus/principles/event-driven.md).
Loaded ambient; enforced by the [drift sensor](../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**Every event is delivered at least once, never exactly once.** Consumers must be idempotent. A consumer that double-processes a redelivered event has a defect; "exactly once" promised by a broker is at-least-once + idempotent consumers under the hood. There is no third option. See [[idempotency]] and [[distributed-systems-fallacies]].

## GOLDEN RULES

- **Aim for events as facts, commands as requests.** Two different transports if the system mixes both; never one channel called "events" carrying both kinds.
- **Aim for events that carry enough state to be useful without a callback** when the consumer is in a different team / service. Event-carried-state-transfer reduces cascading failure modes.
- **Aim for schema evolution as a first-class design concern.** Versioned events, registries, contract tests. The cost of a careless schema change compounds across every consumer.
- **Aim for tracing across event boundaries.** A correlation ID on every event, propagated by every consumer, surfaced in every log. Without it, multi-service incidents are unreconstructible.

---

Traces to: [`corpus/principles/event-driven.md`](../../corpus/principles/event-driven.md).
