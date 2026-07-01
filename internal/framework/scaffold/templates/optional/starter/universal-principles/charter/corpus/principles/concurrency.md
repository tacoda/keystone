# Concurrency and Shared State

Two units of work that may run at the same time, on the same data, are a source of bugs that no test suite can reliably find. Articulated as a discipline by Tony Hoare in *Communicating Sequential Processes* (CACM, 1978), formalized for the JVM in Doug Lea's and Brian Goetz's work, and given its working slogan by Rob Pike: *don't communicate by sharing memory; share memory by communicating.*

The deeper truth: most concurrency bugs are not bugs *in* concurrent code. They are bugs of **design** — a shared mutable thing nobody noticed was shared, or noticed but did not protect. Concurrency principles are mostly about removing shared mutable state, not about synchronizing it.

> **Rules extracted:** [`guides/principles/concurrency.md`](../../guides/principles/concurrency.md). This file holds the full reasoning, anti-patterns, and references.

## The forces

- **Shared mutable state.** Two threads, two processes, two coroutines, two callers — any two units of execution — that can both read and write the same memory at the same time. Without explicit synchronization, the result is undefined: torn reads, lost writes, reads from cache that never see the other thread's write, branches taken on stale values.
- **The memory model.** Modern CPUs and compilers reorder reads and writes. What looks like sequential code in one thread is not what another thread sees. The language's memory model defines what *is* visible, when. The **happens-before** relation (Lamport, 1978) is the formal vocabulary; every concurrency primitive in every modern language is in the business of establishing happens-before edges.
- **Liveness.** Deadlock (everyone waiting for someone), livelock (everyone working, no one progressing), starvation (one party never gets a turn). The dual of safety: correctness *eventually*, not just correctness *now*.

## The strategies, in preference order

1. **No shared state.** The strongest defense. Each unit of work owns its data; communication is by message passing. Actors (Erlang), goroutines + channels (Go), processes + queues, web requests sharing nothing but the database. If two threads cannot reach the same memory, they cannot race on it.
2. **Immutable shared state.** If a value never changes, all threads see the same value. Reading is safe; there is nothing to coordinate. The simplest concurrency primitive is *no mutation*.
3. **Confinement.** A piece of state is owned by exactly one thread; other threads access it only by asking that thread. Single-writer disciplines, event loops, the JavaScript model.
4. **Synchronized mutation.** The fallback. Locks, mutexes, atomics, transactions, compare-and-swap. Powerful and necessary; also where most bugs live. Used last, after the first three have been exhausted.

The order is not aesthetic — it reflects the cost of reasoning. Lock-based code is much harder to reason about correctly than message-passing or immutable-value code. Reach for it last.

## What it asks of you

- When two units of work could touch the same data, ask first whether they need to. Most "shared" state is shared because nobody asked the question. See [[simplicity]] — eliminating the shared thing eliminates the class of bug.
- When you write a class that may be touched by more than one thread, document its **thread-safety contract** plainly: immutable, thread-safe, thread-confined, or unsafe. Ambiguity is how races find their way in. See [[design-by-contract]].
- When you take a lock, write down what it protects, in a comment next to its declaration. A lock with no documented invariant is a lock that will be acquired wrong.
- When you reach for "lock-free," ask why. Lock-free algorithms are correct only with the exact memory-model primitives the literature specifies; *almost* lock-free is incorrect. See [[premature-optimization]].
- When you debug a flaky test, suspect concurrency before suspecting flakiness. A non-deterministic test in a concurrent codebase is usually telling the truth quietly.

## Anti-patterns

- A "thread-safe" collection passed across boundaries, with callers iterating it concurrently with another thread's mutation. The collection's individual operations were safe; the *composite* operation never was.
- A `volatile` field used as a synchronization primitive for anything more than a single-variable visibility flag.
- Double-checked locking written from memory instead of from the language's published idiom.
- A "global lock" that protects everything and therefore protects nothing (deadlocks compose; one global lock is one lock graph).
- "I'll just retry on conflict" — without idempotency, retries change behavior. See [[idempotency]].
- A flaky test marked `@Retry(3)`. The race is still there; the charter is hiding it. See [[fail-fast]].
- Spawning more threads to "make it faster" without a measurement. See [[premature-optimization]].

## References

- Hoare, C. A. R. (1978). "Communicating Sequential Processes." *Communications of the ACM*, 21(8), 666–677.
- Lamport, L. (1978). "Time, Clocks, and the Ordering of Events in a Distributed System." *Communications of the ACM*, 21(7). (The happens-before relation.)
- Lea, D. (1999). *Concurrent Programming in Java: Design Principles and Patterns*, 2nd ed. Addison-Wesley.
- Goetz, B., Peierls, T., Bloch, J., Bowbeer, J., Holmes, D., & Lea, D. (2006). *Java Concurrency in Practice*. Addison-Wesley. (The canonical modern reference; the memory-model discussion is reusable across languages.)
- Pike, R. (2012). "Concurrency is not Parallelism." Heroku Waza talk. (Source of "share memory by communicating.")
- Armstrong, J. (2003). *Making reliable distributed systems in the presence of software errors*. PhD thesis, KTH. (Erlang's actor model; concurrency without shared state.)
- Dijkstra, E. W. (1965). "Solution of a Problem in Concurrent Programming Control." *Communications of the ACM*, 8(9). (The first treatment of mutual exclusion as a problem in its own right.)
