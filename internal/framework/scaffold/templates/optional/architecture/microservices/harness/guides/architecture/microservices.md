# Microservices — rules

The rules from [`corpus/microservices.md`](../corpus/microservices.md).
Loaded ambient; enforced by the [drift sensor](../../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**No shared database.** Two services that read or write each other's tables are not two services; they are one service with two deployment artifacts. The single biggest predictor of microservice-architecture failure is shared state in a shared schema. Each service owns its data; cross-service data exchange goes through APIs or events.

## GOLDEN PATH

- **Aim for service boundaries that match team boundaries.** Conway's law is descriptive — the architecture *will* mirror the org chart eventually. Pick boundaries the org can actually staff and own.
- **Aim for synchronous calls minimized; asynchronous events preferred** at service boundaries. Synchronous fan-out cascades tail latency; events isolate. See [[event-driven]].
- **Aim for one deployment pipeline per service, fully owned by the service team.** Shared pipelines reintroduce the coupling microservices were meant to eliminate.
- **Aim for observability built in from day one.** Distributed tracing, structured logging with correlation IDs, SLOs per service. Without these, multi-service incidents are unsolvable. See [[observability]].
- **Aim to start with a monolith.** Most successful microservice architectures came from a monolith that needed to scale. Premature microservices are how teams ship one feature in a quarter that a monolith would have shipped in a week. See [[premature-optimization]].

---

Traces to: [`corpus/microservices.md`](../corpus/microservices.md).
