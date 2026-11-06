# Event-Driven Architecture

Components communicate by publishing and consuming **events** — immutable facts about something that happened — rather than by direct calls. Articulated as an enterprise pattern in Hohpe & Woolf's *Enterprise Integration Patterns* (2003) and clarified by Martin Fowler in "What do you mean by 'Event-Driven'?" (2017), which distinguishes four very different things people call event-driven.

The shared shape: a producer says *X happened*; one or more consumers react. The producer does not know who is listening. The consumer does not know who emitted.

> **Rules extracted:** [`guides/event-driven.md`](../guides/event-driven.md). This file holds the full reasoning, anti-patterns, and references.

## Fowler's four event-driven patterns

These are often confused but solve different problems. Pick deliberately.

### Event Notification
A component publishes "this thing happened" to signal others. Consumers may need to call back to the producer for details. **Loosest coupling**, but produces request fan-out and the network-call cascade that comes with it.

### Event-Carried State Transfer
The event carries enough state that consumers don't need to call back. Consumers may keep a local replica of the data they care about. Higher coupling on the event's *schema*; lower coupling on the producer's *availability*.

### Event Sourcing
The system's state is **derived** from an append-only log of events. The current state is a fold over the log. Powerful for audit, time-travel, debugging; expensive in operational complexity and schema migration. See [[hyrums-law]] — every event format becomes a contract forever.

### CQRS (Command Query Responsibility Segregation)
Writes and reads use separate models. Writes produce events; read models are projections built from those events. Often paired with Event Sourcing; sometimes used alone.

The first two are about **communication style.** The second two are about **state management.** Conflating them produces architectures that nobody can navigate.

## What it asks of you

- When you choose event-driven, choose **which** event-driven (Fowler's four). State which pattern you are using; ban informal use of "events" for everything.
- When you publish an event, name it as a **past-tense fact**: `OrderPlaced`, `EmailVerified`, `PaymentRefunded`. Imperative names (`SendEmail`, `RefundPayment`) describe a *command*, not an event — and commands deserve a different transport.
- When you consume an event, treat the consumer as **at-least-once.** The event broker will redeliver. Make the handler idempotent. See [[idempotency]].
- When you design an event schema, plan its evolution from day one. Hyrum's Law applies in full force — see [[hyrums-law]]. Use additive-only changes where possible; version explicitly when not.
- When tracing a flow, observability is now essential. See [[observability]] — events sever the local stack trace.

## Anti-patterns

- An "event" that the producer waits for a response to. That's a request/reply call wearing the wrong label.
- Event handlers that mutate shared state without idempotency. Redelivery produces duplicate effects.
- A canonical event format owned by no team, evolving by accretion as each consumer adds the field they need. The format is now a permanent contract with every consumer.
- A debugging session that starts with "let me trace where this event came from" and ends two hours later. Observability was not built in.
- Event sourcing chosen because it is fashionable, not because audit/replay was required. The operational cost arrives long before the value.
- A "command" that is broadcast like an event, with multiple receivers each thinking they are *the* receiver. Two refunds get issued; one would have been correct.

## References

- Hohpe, G., & Woolf, B. (2003). *Enterprise Integration Patterns: Designing, Building, and Deploying Messaging Solutions*. Addison-Wesley.
- Fowler, M. (2017). "What do you mean by 'Event-Driven'?" martinfowler.com/articles/201701-event-driven.html.
- Young, G. (2010). *CQRS Documents*. (The canonical articulation of CQRS as distinct from event sourcing.)
- Vernon, V. (2013). *Implementing Domain-Driven Design*. Addison-Wesley. (Domain events; ES patterns in DDD context.)
- Helland, P. (2007). "Life Beyond Distributed Transactions: an Apostate's Opinion." CIDR. (The grand backdrop — eventual consistency as the inescapable rule once messaging is involved.)
- Kleppmann, M. (2017). *Designing Data-Intensive Applications*. O'Reilly. (Chapters 11–12 — log-structured data, event-driven architectures, change data capture.)
