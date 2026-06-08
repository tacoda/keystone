# Serverless Architecture — rules

The rules from [`corpus/serverless.md`](../corpus/serverless.md).
Loaded ambient; enforced by the [drift sensor](../../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**Every function is stateless.** Instance-local state is incidental — the platform may reuse an instance for the next invocation, or not. Code that assumes persistence between invocations is wrong by construction; correctness must derive entirely from inputs and external state (database, queue, parameter store).

## GOLDEN RULES

- **Aim for functions that do one thing.** A function that branches on event type to do five different jobs is five functions wearing one trench coat.
- **Aim for synchronous chains kept short.** Each hop adds latency, error modes, and concurrency-limit interactions. Two hops is fine; six is a red flag.
- **Aim for short cold-start paths.** Lazy-load only what an invocation needs. Heavy frameworks at the top of the file cost on every cold start.
- **Aim for end-to-end tracing.** A correlation ID injected at the edge, propagated through every event and every function. Without it, multi-function debugging is forensic archaeology.
- **Aim for managed services over self-managed where they exist.** Queues, schedulers, secret stores — let the platform run them. Resist the urge to spin up a small EC2 instance "just for this one thing."

---

Traces to: [`corpus/serverless.md`](../corpus/serverless.md).
