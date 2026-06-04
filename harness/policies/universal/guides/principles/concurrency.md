# Concurrency and Shared State — rules

The non-negotiable and strongly-preferred rules from [`corpus/principles/concurrency.md`](../../corpus/principles/concurrency.md).
Loaded ambient; enforced by the [drift sensor](../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**No shared mutable state without synchronization.** *Every* read of a mutable value visible to another thread, and *every* write, must go through a synchronization primitive that establishes happens-before with respect to the other thread. There is no "the variable is just a boolean" exception. There is no "but the operation is atomic on this CPU" exception. The language's memory model is the contract; respect it or the program is undefined.

## GOLDEN RULES

- **Aim to eliminate sharing before synchronizing it.** A bug that cannot exist is the cheapest bug.
- **Aim for immutable values across thread boundaries.** Pass copies, pass snapshots, pass frozen structures — anything but a live mutable reference.
- **Aim for one writer.** Reads from many threads are cheap when there is exactly one writer; the moment there are two writers, the cost compounds.
- **Aim for synchronization with names.** A `mutex` called `lock` protects nothing in particular; a `mutex` called `cacheEntriesByKey` declares its job. The name is half the discipline.

---

Traces to: [`corpus/principles/concurrency.md`](../../corpus/principles/concurrency.md).
