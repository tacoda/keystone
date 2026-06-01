# Observability

A system is **observable** when you can answer new questions about its behavior in production *without redeploying it.* The term is borrowed from control theory (Kalman, 1960) — a system is observable if its internal state can be inferred from its outputs — and re-applied to software by Cindy Sridharan in *Distributed Systems Observability* (2018) and by Charity Majors and colleagues in *Observability Engineering* (2022).

Observability is not a tool, a vendor, or a "three pillars" architecture. It is a **property** of the system: how readily the system reveals what it is doing.

## Monitoring vs. observability

The vocabulary in this space is contested. The clearest line:

- **Monitoring** answers *known unknowns*. You know what could go wrong; you set up alarms for those things. Disk filling up. Error rate climbing. Latency exceeding threshold. The questions are pre-stated.
- **Observability** answers *unknown unknowns*. Something is wrong that nobody anticipated; can the system tell you, after the fact, what happened? The questions are asked during incident response, against data already collected.

A monitored system tells you the things its authors *expected* could break. An observable system tells you the things its authors did not. Real systems need both — monitoring catches the routine; observability handles the novel.

## The signals

There are three traditional "pillars" — **logs**, **metrics**, **traces** — but the more useful frame is the one Charity Majors articulated: a system is observable in proportion to its capacity for **high-cardinality, high-dimensionality, structured events.**

- **Logs** carry the most information per event but are the slowest to query at scale. *Structured* logs (key-value, machine-parseable) are usable; unstructured text logs are write-only.
- **Metrics** are cheap to aggregate but lose all per-event detail. They tell you *that* the error rate jumped; they cannot tell you *which users* or *what request shape*. Pre-aggregation is what makes them cheap; pre-aggregation is also their limitation.
- **Traces** show the path of a single request across services. Indispensable for distributed systems — see [[distributed-systems-fallacies]].

The modern critique: dashboards full of pre-aggregated metrics produce systems that are *monitored* (you can see what you expected) but not *observable* (you cannot ask new questions when an incident departs from expectations). The recovery is to invest more in **wide events** — single log lines per request, carrying as many fields as can be afforded, queryable later by any dimension.

## What it asks of you

- When you write a code path that can fail, ensure the failure has enough context to be diagnosed by someone who did not write the code. See [[error-handling]] on error context, [[fail-fast]] on loud failures.
- When you choose log fields, prefer **high cardinality** (`user_id`, `request_id`, `tenant`, `feature_flag_value`) over **low cardinality** (`level`, `service`). The high-cardinality fields are what let you slice during an incident.
- When you instrument a request path, propagate a correlation ID (trace ID) from the edge through every downstream call. Without it, multi-service incidents cannot be reconstructed.
- When you sample, sample *traces* together — keep all spans of a kept trace; drop all spans of a dropped one. Half-sampled traces are worse than no traces.
- When you log a secret, you have created an incident. See [[secrets-management]] — *rotate, don't redact*.
- When you build a feature, ask: *if this misbehaves at 3am, what would the on-call engineer need to see?* If the answer is "I'd have to redeploy with more logging," the feature is not done.

## Observability and design

Observability is a design property; it cannot be bolted on. The decisions that make a system observable are made at the same time as the ones that make it correct:

- A function whose effects are channeled through one observable seam is debuggable; a function with side effects scattered across modules is not. See [[separation-of-concerns]].
- A request that has a stable identity from edge to storage is traceable; a request that loses its identity at each layer is a needle in a haystack at every layer.
- A system whose components emit structured events about their decisions can be replayed and explained; one that emits only outcomes can only be guessed at.

The implication: instrumentation is not a layer you add on top of the design. It is a constraint on the design. See [[modern-software-engineering]] — short feedback loops require observable systems.

## IRON LAW

**If you cannot debug it from the data, you cannot debug it.** During an incident at 3am, the only information available is whatever the system was already emitting. Adding more logging requires a deploy; a deploy requires the system to be working enough to ship; if it weren't working, you wouldn't be debugging. *Pre-position* the data; that is the whole game.

## GOLDEN RULES

- **Aim for one wide structured event per unit of work.** A request, a job, a message — one event that carries everything you might want to ask about it later. Fan-out per-event from there if needed.
- **Aim for high-cardinality fields.** User IDs, request IDs, version hashes, feature flag states. The questions worth asking during an incident are nearly always sliced by some high-cardinality dimension.
- **Aim for context that crosses service boundaries.** A trace ID propagated everywhere; logs that include it; metrics tagged with it where cardinality allows.
- **Aim for "could I diagnose this from the data?" as a review question.** Same standing as "is this correct?" and "is this tested?"

## Anti-patterns

- A dashboard with thirty graphs, none of which would have caught the last incident.
- Logs that consist of `"started"`, `"working"`, `"done"` — no IDs, no context, no structured fields.
- A metrics-only stack used to debug per-user issues. Aggregates cannot answer per-entity questions.
- Sampling that drops fields independently of each other; the surviving event has no trace ID and no user ID.
- An incident postmortem whose "what we'd improve" includes "add logging to find out what happened next time." This is the IRON LAW saying *you weren't observable.*
- A vendor change framed as an "observability upgrade." The vendor is operational; the property is in the code.

## References

- Sridharan, C. (2018). *Distributed Systems Observability: A Guide to Building Robust Systems*. O'Reilly. (The modern reframing of the term; the canonical short text.)
- Majors, C., Fong-Jones, L., & Miranda, G. (2022). *Observability Engineering: Achieving Production Excellence*. O'Reilly. (The wide-events position, the critique of metrics-only stacks.)
- Beyer, B., Jones, C., Petoff, J., & Murphy, N. R. (2016). *Site Reliability Engineering: How Google Runs Production Systems*. O'Reilly. (The four golden signals — latency, traffic, errors, saturation — and the monitoring/alerting discipline.)
- Kalman, R. E. (1960). "On the General Theory of Control Systems." *Proceedings of the First International Congress of IFAC*. (The original control-theoretic notion of observability.)
- Cantrill, B. (2006). "Hidden in Plain Sight." *ACM Queue*. (Why pre-positioning instrumentation matters more than adding it later.)
- Allspaw, J. (2016). *Trade-Offs Under Pressure: Heuristics Used by Incident Commanders*. Master's thesis, Lund University. (Why what data is available *at the moment of incident* is the constraint that dominates.)
