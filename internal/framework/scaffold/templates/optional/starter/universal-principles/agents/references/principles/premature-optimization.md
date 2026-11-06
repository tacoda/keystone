# Premature Optimization

> Programmers waste enormous amounts of time thinking about, or worrying about, the speed of noncritical parts of their programs, and these attempts at efficiency actually have a strong negative impact when debugging and maintenance are considered. We *should* forget about small efficiencies, say about 97% of the time: **premature optimization is the root of all evil.** Yet we should not pass up our opportunities in that critical 3%. — Donald Knuth (*Computing Surveys*, 1974)

The quotation is famous; the full sentence is rarely cited. Knuth is **not** saying "do not optimize." He is saying: do not optimize *before measurement*, do not optimize *the wrong code*, and do not pay simplicity costs for performance you have no evidence you need.

> **Rules extracted:** [`guides/principles/premature-optimization.md`](../../guides/principles/premature-optimization.md). This file holds the full reasoning, anti-patterns, and references.

## The principle

Optimization is a tradeoff: you spend complexity, readability, change-tolerance — sometimes correctness — to buy speed or memory. The trade is **only worth it** when you have evidence that the speed/memory matters in the part of the system being changed. Without evidence, the trade is a one-way debit.

Three regimes:

1. **Premature** — optimization before profiling. The code is faster *somewhere*; you do not know if it is the somewhere that mattered.
2. **Targeted** — optimization in a hot path identified by measurement, with a measured improvement and a measured cost. This is engineering.
3. **Architectural** — design-time decisions about algorithmic complexity, data structures, IO patterns. These *are* worth thinking about up front, because they are expensive to retrofit. Choosing O(n log n) over O(n²) is not premature; choosing hand-unrolled assembly over a loop is.

## What it asks of you

- When you are about to make code "faster," ask: *measured against what baseline, on what workload?* If you cannot answer, the optimization is premature. See [[modern-software-engineering]] (empirical).
- When the existing code is slow, profile before refactoring. The hottest line is almost never where you expected. Working from intuition is how teams optimize the 97% and leave the 3% slow.
- When a teammate proposes an optimization, ask whether it changes the architectural class (O-complexity, IO pattern, batch vs. stream) or only the constant factor. The first is usually worth the cost; the second usually is not.
- When you cannot avoid the optimization, *isolate* it. A hot inner loop wrapped in a clear interface contains its complexity; the same code smeared across the module taxes every reader. See [[information-hiding]].

## Architectural choices that are *not* premature

Knuth's caveat — *the critical 3%* — covers more than people remember. The following decisions are made up front *because they are hard to undo*, not because they are speculative:

- Algorithmic complexity class for data that will grow.
- IO patterns (sync vs. async, batch vs. stream, push vs. pull).
- Storage model (normalized vs. denormalized, row vs. column).
- Cache placement at architectural boundaries.

Reaching for the *right* algorithm or data structure on day one is not premature optimization. Reaching for hand-tuned bit manipulation on day one is. The line is: *is this changing the curve, or only the constant?*

## Anti-patterns

- Manually inlining a function "because the compiler might not."
- Switching from a clear collection type to a custom one "for performance" without a benchmark.
- A code review comment that says "this might be slow" with no measurement attached.
- A "fast path" that adds complexity to every read because some unmeasured workload might benefit.
- Pre-computing and caching values that are never re-read.

## References

- Knuth, D. E. (1974). "Structured Programming with go to Statements." *Computing Surveys*, 6(4), 261–301. (The source of the quotation, in context.)
- Hoare, C. A. R., quoted by Knuth in the same paper, originating the "root of all evil" phrasing.
- Bentley, J. (1982). *Writing Efficient Programs*. Prentice Hall. (Disciplined approach to performance: profile, change, measure.)
- Gregg, B. (2020). *Systems Performance: Enterprise and the Cloud*, 2nd ed. Pearson. (Modern systems-level treatment of the measurement discipline.)
