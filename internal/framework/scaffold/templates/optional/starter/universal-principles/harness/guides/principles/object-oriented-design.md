# Object-Oriented Design Heuristics — rules

The non-negotiable and strongly-preferred rules from [`corpus/principles/object-oriented-design.md`](../../corpus/principles/object-oriented-design.md).
Loaded ambient; enforced by the [drift sensor](../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**Behavior lives with the data it depends on.** A rule about an object's state — what makes it valid, what transitions are allowed, what invariants must hold — belongs on that object. Spreading it across callers is not OO design; it is procedural code in a class costume.

## GOLDEN PATH

- **Aim for objects with verbs, not just nouns.** A bag of getters and setters is a data structure, not an object. Data structures are fine; just don't pretend they encapsulate behavior.
- **Aim for dependencies on contracts, not classes.** The concrete class should be visible only at the seam where it is wired up.
- **Aim for shallow hierarchies.** A class hierarchy more than two levels deep is a hypothesis about taxonomy that you are probably wrong about.

---

Traces to: [`corpus/principles/object-oriented-design.md`](../../corpus/principles/object-oriented-design.md).
