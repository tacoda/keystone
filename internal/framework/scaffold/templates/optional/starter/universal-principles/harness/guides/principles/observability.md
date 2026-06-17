# Observability — rules

The non-negotiable and strongly-preferred rules from [`corpus/principles/observability.md`](../../corpus/principles/observability.md).
Loaded ambient; enforced by the [drift sensor](../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**If you cannot debug it from the data, you cannot debug it.** During an incident at 3am, the only information available is whatever the system was already emitting. Adding more logging requires a deploy; a deploy requires the system to be working enough to ship; if it weren't working, you wouldn't be debugging. *Pre-position* the data; that is the whole game.

## GOLDEN PATH

- **Aim for one wide structured event per unit of work.** A request, a job, a message — one event that carries everything you might want to ask about it later. Fan-out per-event from there if needed.
- **Aim for high-cardinality fields.** User IDs, request IDs, version hashes, feature flag states. The questions worth asking during an incident are nearly always sliced by some high-cardinality dimension.
- **Aim for context that crosses service boundaries.** A trace ID propagated everywhere; logs that include it; metrics tagged with it where cardinality allows.
- **Aim for "could I diagnose this from the data?" as a review question.** Same standing as "is this correct?" and "is this tested?"

---

Traces to: [`corpus/principles/observability.md`](../../corpus/principles/observability.md).
