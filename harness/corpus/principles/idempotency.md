# Idempotency

An operation is **idempotent** when applying it twice has the same effect as applying it once. Articulated mathematically long before software (Benjamin Peirce, 1870), elevated into a working principle of distributed systems by Pat Helland's *Idempotence Is Not a Medical Condition* (ACM Queue, 2012). In a world where the network is unreliable — see [[distributed-systems-fallacies]] — idempotency is what lets retries be safe.

> **Rules extracted:** [`guides/principles/idempotency.md`](../../guides/principles/idempotency.md). This file holds the full reasoning, anti-patterns, and references.

## Why it matters

A request crosses a network. The response does not arrive. Did the request succeed and the response get lost, or did the request itself fail? You cannot tell. The caller has three choices:

1. **Give up.** Sometimes correct, often unacceptable (the user just paid for something).
2. **Retry.** Safe only if the operation is idempotent.
3. **Ask the server which it was.** Requires a protocol that supports it — and that protocol has to rely on… idempotency at some level.

The honest summary: **"exactly once" delivery is a fiction.** What real systems implement is *at-least-once delivery* of *idempotent operations*. The combination is the working substitute. Where the operation is not naturally idempotent, the engineering job is to make it idempotent — usually by attaching a client-generated identifier the server uses to deduplicate.

## What it asks of you

- When you design an operation that mutates state, ask: *can it be safely retried?* If yes, document the property. If no, make it so — typically by adding an idempotency key. See [[design-by-contract]].
- When you call across a network, the call has three possible outcomes — success, failure, *unknown*. The "unknown" case is the one idempotency exists for. See [[distributed-systems-fallacies]].
- When you write a retry loop, the operation it retries must be idempotent. A non-idempotent retry is a bug latent in the network's behavior. See [[error-handling]].
- When you build a queue or message-bus consumer, assume **every message will be delivered more than once**. The consumer's idempotency is the only thing that prevents duplicate processing.
- When you expose an HTTP API, follow RFC 9110 on which methods *are* idempotent (`GET`, `PUT`, `DELETE`, `HEAD`, `OPTIONS`) and which are not (`POST`, `PATCH`). For non-idempotent operations, expose an idempotency-key header. See [[least-astonishment]] — diverging from the standard semantics is expensive.

## The idempotency-key pattern

The canonical mechanism for making a non-idempotent operation safely retryable:

1. The **client generates a unique key** for the operation (UUID, hash of inputs, or sequence number) and sends it with the request.
2. The **server records the key and the result** in storage atomically with the operation itself.
3. If the **same key arrives again**, the server returns the stored result without re-executing.
4. The key has a **lifetime** — long enough to cover realistic retry windows (minutes to hours), short enough to bound storage growth.

The crux is step 2: the deduplication record and the operation must be **committed together**, or the server can lie. The standard implementation puts both in the same database transaction. Without atomicity, a client retry can land between the operation and the dedup record — and the server believes it hasn't seen this key, processes it again, and the user is charged twice.

## What "idempotent" does — and doesn't — promise

Idempotent does not mean *side-effect-free* (a GET that hits a cache and increments a counter is not pure but still idempotent — repeated calls leave the same end state). It does not mean *commutative* (operations A then B and B then A may produce different results even when each is idempotent). And it does not mean *concurrency-safe* (two clients sending the same idempotent operation at the same time can still race — see [[concurrency]]).

The single guarantee: **the second-and-later applications of the operation, with the same inputs, leave the system in the state the first application would have produced.** Useful, narrow, and load-bearing.

## Anti-patterns

- A POST endpoint that creates a record, with retries by clients, with no idempotency key. Every flaky network produces duplicate records.
- A queue worker that processes a message, then acks. If the worker crashes between processing and ack, the next worker processes the same message. The processing has to be idempotent for this design to be correct.
- A retry policy at three layers of the stack — client, gateway, service — none of them coordinating, all of them multiplying. One real failure produces dozens of attempts.
- Storing the dedup key after the operation commits, in a separate transaction. There is a window in which the operation succeeded but the dedup key is not yet recorded; a retry in that window double-processes.
- A "compensating action" framed as the fix for a non-idempotent operation, with no design for what happens when the compensating action itself fails.
- An HTTP API where `POST /resource` is idempotent and `PUT /resource` is not. The conventions exist; respect them — see [[least-astonishment]].

## References

- Helland, P. (2012). "Idempotence Is Not a Medical Condition." *ACM Queue*, 10(4). (The canonical software-engineering treatment.)
- Fielding, R., & Reschke, J. (Eds.) (2022). RFC 9110: *HTTP Semantics*. (The current normative source for HTTP method idempotency and safe methods.)
- Vogels, W. (2008). "Eventually Consistent." *ACM Queue*, 6(6). (Distributed-systems consistency models — the surrounding context within which idempotency operates.)
- Hohpe, G., & Woolf, B. (2003). *Enterprise Integration Patterns*. Addison-Wesley. (The *Idempotent Receiver* pattern, ch. 10 — the canonical messaging-system articulation.)
- Kleppmann, M. (2017). *Designing Data-Intensive Applications*. O'Reilly. (Chapter 8 — the "exactly-once is at-least-once plus idempotent" framing.)
