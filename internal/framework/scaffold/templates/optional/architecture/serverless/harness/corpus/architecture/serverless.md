# Serverless Architecture

Code runs in short-lived, managed compute units (AWS Lambda, GCP Cloud Functions, Azure Functions, Cloudflare Workers) provisioned and torn down by the platform. The application author writes functions; the platform handles servers, scaling, patching, and idle scaling-to-zero. The umbrella term covers both **Functions-as-a-Service (FaaS)** and **managed services** (DynamoDB, S3, SQS, EventBridge) that compose into serverless applications.

Articulated as a distinct pattern by Mike Roberts ("Serverless Architectures", martinfowler.com, 2016, updated 2018) and operationalized by Sam Newman, Yan Cui, and the AWS architecture team. The promise: pay only for what you use, scale instantly, eliminate operational toil. The catch: every promise lands inside the constraints in [[distributed-systems-fallacies]].

> **Rules extracted:** [`guides/serverless.md`](../guides/serverless.md). This file holds the full reasoning, anti-patterns, and references.

## Defining properties

- **No server management.** No host OS, no kernel patches, no autoscaling rules. The platform's contract is: send the function an event, get a response.
- **Scale-to-zero, scale-to-n.** Cold start when traffic arrives; provision and tear down by request load. Idle cost approaches zero.
- **Stateless compute.** Each invocation is independent. Persistent state lives in managed services (databases, queues, object stores), never in the function instance.
- **Event-driven by default.** Functions are triggered by events — HTTP requests, queue messages, schedule, file uploads — not by long-running processes accepting connections.
- **Pay-per-use pricing.** Cost is request-count × duration × memory, not provisioned capacity.

## The cost the brochure does not mention

Serverless does not eliminate complexity; it relocates it. The complexity moves from infrastructure to:

- **Cold starts.** First invocation after idle pays the platform's bootstrapping cost — milliseconds to seconds, language- and runtime-dependent. Often the long-tail of latency. See [[distributed-systems-fallacies]] (latency is zero, and serverless's tail latency is not).
- **Local development.** No local environment fully reproduces the cloud platform. Test cycles run against a deployed environment; the feedback loop is slower than for a monolith. See [[modern-software-engineering]] (short feedback loops).
- **Observability.** Distributed by default. Every function is a separate process, every invocation a separate context. Tracing is essential, not optional. See [[observability]].
- **Vendor coupling.** The event shapes, the API gateway behavior, the IAM model, the secret manager — each is platform-specific. Migrating between providers is a rewrite.
- **Concurrency limits.** Each cloud limits in-flight invocations per account / per function. Hitting the limit produces throttling errors that look like outages. See [[fail-fast]] and [[error-handling]].

## What it asks of you

- When you write a function, design it to be **invoked many times concurrently** with no shared state between invocations. Globals are anti-patterns; in-memory caches are best-effort, not authoritative.
- When you handle an event, assume **at-least-once delivery.** Queues redeliver; HTTP retries happen; the platform may retry on its own. Idempotency is mandatory. See [[idempotency]].
- When you log, log structured events with the invocation ID, the trace ID, and the originating event ID. The local stack trace ends at the function boundary. See [[observability]].
- When you fail, **fail loudly and fast.** A hung function consumes its full timeout and bills you for it; a fast failure releases the slot. See [[fail-fast]].
- When you provision a function, give it the **minimum IAM permissions** it needs. The blast radius of a compromise is the union of every function's permissions. See [[security]] (least privilege).
- When you compose multiple functions into a workflow, prefer **orchestration via managed step functions / state machines** to function-to-function chains. Chained functions multiply the failure modes; a state machine puts the orchestration in one place. See [[event-driven]].

## Anti-patterns

- A function with a 15-minute timeout doing batch work. The platform charges for the whole timeout when the function hangs; long-running batch belongs on a different compute model.
- Functions that share an in-memory cache "because they're the same warm instance." Sometimes they are; sometimes they aren't. The cache is wrong on a fraction of invocations.
- Chains of synchronous function-to-function HTTP calls. Tail latency multiplies; one slow link slows them all. See [[distributed-systems-fallacies]].
- A function with `iam:*` permissions because narrower scopes were "annoying to set up." See [[security-threats]] (cloud IAM misconfiguration).
- Logging the entire event payload, including secrets, into CloudWatch. See [[secrets-management]].
- A "serverless" app that costs more than the equivalent always-on VM. The economics inverted; serverless is no longer the answer.
- Functions that share a relational database with a connection per invocation. Connection storms at scale; the database becomes the bottleneck the function-tier was meant to escape. Use connection pooling proxies (RDS Proxy) or move state to scale-friendly stores.

## References

- Roberts, M. (2016, updated 2018). "Serverless Architectures." martinfowler.com/articles/serverless.html.
- Cui, Y. (2019). *Production-Ready Serverless*. Manning.
- Sbarski, P. (2017). *Serverless Architectures on AWS*. Manning.
- Wittig, A., & Wittig, M. (2019). *Amazon Web Services in Action*, 2nd ed. Manning. (Operational context for AWS-native serverless apps.)
- Newman, S. (2021). *Building Microservices*, 2nd ed. O'Reilly. (Chapters on serverless as a microservice substrate.)
