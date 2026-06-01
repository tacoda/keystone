# The Fallacies of Distributed Computing

Eight assumptions that everyone writing distributed code makes implicitly, all of which are false. Articulated by L. Peter Deutsch at Sun Microsystems in 1994 (the first seven), with the eighth added later by James Gosling. Every fallacy is the source of a known incident class; every defense against one is documented engineering practice.

This file states the fallacies; the operational practices for each — retries, timeouts, circuit breakers, multi-region failover, observability — belong in `../idioms/<stack>/` and in [[observability]].

## The eight fallacies

### 1. The network is reliable
Packets are dropped. Connections are reset. Routers reboot. Cables get cut by backhoes. Cloud providers have entire-region outages. Code that does not assume failure will, the first time the assumption is tested, behave as if the failure had not happened — which is the worst possible behavior. See [[fail-fast]] and [[error-handling]].

### 2. Latency is zero
A local function call is nanoseconds; an in-datacenter RPC is sub-millisecond; a cross-region call is tens of milliseconds; a transcontinental call is hundreds. A loop of one hundred calls that is fine locally is a five-second user-visible request across a network. The cost is asymmetric — read-heavy chatty designs that work in development break under modest production load.

### 3. Bandwidth is infinite
Even within a datacenter, every link has a ceiling. A "small" payload multiplied by a "small" amount of traffic is a routine source of saturation. The right-sizing question is not "is this small?" but "is this small *at peak times the population this code will serve*?"

### 4. The network is secure
Every network between your code and its peer is hostile until proven otherwise. There is no "internal network." TLS everywhere, authenticate every call, expect MITM, expect replay. See [[security]] (narrow trust boundaries) and [[security-threats]] (SSRF, supply chain, confused deputy — all assume an attacker on the network path).

### 5. Topology doesn't change
The set of machines, their addresses, which one is the leader, which AZ is healthy — all change continuously and without notice. Code that resolves a name once at startup and caches the IP is correct for an hour and broken for the next nine. Service discovery, health checks, and re-resolution are not optional features; they are the consequence of this fallacy being real.

### 6. There is one administrator
Production has many operators, many policies, and many simultaneous changes. The set of feature flags, IAM policies, network rules, and runtime configs in effect right now is not knowable from any one viewpoint. Designs that assume a single coherent operator do not survive contact with operational reality.

### 7. Transport cost is zero
Serialization, deserialization, TLS handshake, JSON parsing, schema validation — each call pays a cost beyond the wire latency. A microservice architecture where every request fans out to twenty services pays this cost twenty times per user request. The number that matters is not "how fast is one call?" but "what is the cost of a call *in aggregate* under realistic load?"

### 8. The network is homogeneous
Some clients are on fiber, some on 3G; some servers are bare-metal, some are containers under a noisy neighbor; some links are within a rack, some cross an ocean. Code calibrated to one slice of the population breaks for everyone else. See [[postels-law]] — diversity at the edge is the rule, not the exception.

## What it asks of you

- When you write a network call, ask: *what happens if it never returns?* That is the design question. A call without a timeout is a call that can hang the system; a timeout without a retry is a call that will fail under transient blips; a retry without idempotency is a call that double-spends. See [[idempotency]].
- When you fan a request out to N peers, the latency of the request is the latency of the **slowest** peer, not the average. Tail-latency reasoning is the only honest kind for distributed systems.
- When you assume the network is your own, you have already lost. Treat every wire as adversarial. See [[security]].
- When you design caches, expect them to be inconsistent with the source of truth — the network between cache and source obeys all eight fallacies.
- When you operate in production, expect that two parts of the system disagree about the state of the world right now. The argument is which-truth-wins, not whether-there-is-one-truth.

## IRON LAW

**Every network operation is fallible and slow.** "Fallible" means *every* call has three outcomes — success, failure, and *unknown* (the request may or may not have happened). "Slow" means orders of magnitude slower than the equivalent local operation. Code that treats network calls as local calls is wrong, regardless of how the test suite behaves.

## GOLDEN RULES

- **Aim for timeouts on every call.** No exceptions. An untimed call is a vector for cascading failure.
- **Aim for idempotency at every retry seam.** Retries without idempotency convert flaky networks into duplicated state changes. See [[idempotency]].
- **Aim for failure modes that are observable.** When a call fails, the next operator needs to know *which* call, *to which peer*, *with what context*. See [[observability]].
- **Aim for degraded modes, not all-or-nothing.** A system that returns a stale but valid answer when a dependency is down beats one that returns 500.

## Anti-patterns

- A retry loop with no backoff, multiplying load on a struggling peer until it fails entirely (retry storm / thundering herd).
- A health check that reports green based on process liveness, not on the dependencies the process actually needs.
- A timeout configured to match the *average* call latency. The slow path is what will hurt; the average is irrelevant.
- A microservices architecture whose request graph nobody has mapped, because "each service is independent."
- A "we'll add metrics later" deployment to production. See [[observability]].
- DNS resolved once at startup; the application reconnects "automatically" by restarting.
- An assumption that the database, queue, or cache is "always there." Each is a network peer; each obeys all eight fallacies.

## References

- Deutsch, L. P. (1994). *The Eight Fallacies of Distributed Computing*. Sun Microsystems internal memo. (The original seven; the eighth was added by Gosling.)
- Rotem-Gal-Oz, A. (2006). "Fallacies of Distributed Computing Explained." rgoarchitects.com. (The canonical written elaboration.)
- Waldo, J., Wyant, G., Wollrath, A., & Kendall, S. (1994). "A Note on Distributed Computing." Sun Microsystems Laboratories. (Companion paper; argues that the local/remote distinction cannot be hidden by a transparent RPC layer.)
- Helland, P. (2007). "Life Beyond Distributed Transactions: an Apostate's Opinion." CIDR. (How distributed systems actually behave when you stop pretending they are not.)
- Kleppmann, M. (2017). *Designing Data-Intensive Applications*. O'Reilly. (Modern textbook; chapters 8–9 cover the consequences of every fallacy in detail.)
