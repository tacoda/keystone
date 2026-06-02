# Information Hiding

Modules should be characterized by the design decisions they hide from the rest of the system. Articulated by David L. Parnas in "On the Criteria To Be Used in Decomposing Systems into Modules" (*Communications of the ACM*, 1972) — the foundational paper for modularity in modern software.

> **Rules extracted:** [`guides/principles/information-hiding.md`](../../guides/principles/information-hiding.md). This file holds the full reasoning, anti-patterns, and references.

## The principle

A module's interface is a *contract*; its implementation is a *secret*. The right way to decompose a system is not along the lines of its flowchart, but along the lines of its design decisions — each module hides one decision behind a stable interface, so the decision can change without rippling out.

Parnas's two-systems comparison (the KWIC index) showed that two decompositions producing the same output can have radically different change-tolerance: one couples every module to a shared data structure; the other hides each design decision. Same output, opposite trajectories under change.

## What it asks of you

- When you design a module, ask: *what decision is this module hiding?* If you cannot name one, the module probably has no reason to exist.
- When you expose internal data, ask whether you are committing to its shape forever — or whether a method that does the operation would let the shape change later.
- When you find yourself needing to know an implementation detail to use a module, the interface is wrong.

## Anti-patterns

- A module that exposes a `data` dict / struct of internal state.
- Two modules that read each other's internals to coordinate.
- A "facade" that simply re-exports every method of the underlying implementation.
- Public fields where a method would do.
- An "ORM entity" exposed all the way to the HTTP boundary.

## References

- Parnas, D. L. (1972). "On the Criteria To Be Used in Decomposing Systems into Modules." *Communications of the ACM*, 15(12), 1053–1058.
- Parnas, D. L., Clements, P., & Weiss, D. M. (1985). "The Modular Structure of Complex Systems." *IEEE Transactions on Software Engineering*, 11(3).
- Ousterhout, J. (2018). *A Philosophy of Software Design*. Yaknyam Press. (Chapter 4, "Modules Should Be Deep.")
