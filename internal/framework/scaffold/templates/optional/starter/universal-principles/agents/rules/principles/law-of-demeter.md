# Law of Demeter — rules

The non-negotiable and strongly-preferred rules from [`references/principles/law-of-demeter.md`](../../references/principles/law-of-demeter.md).
Loaded ambient; enforced by the [drift hook](../../hooks/drift.md). The reference file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**A method's coupling is the union of every type it names.** Every chained call adds a type to that union. Demeter is the discipline of keeping the union small enough that you can hold it in your head while you reason about change.

## GOLDEN RULE

- **Aim for methods that operate on what they were given.** Parameters, fields, locals — not the transitive object graph reachable from them.
- **Aim for behavior at the right level.** If the caller is computing something *about* an object's internals, the computation probably belongs inside the object.
- **Aim to distinguish a navigation chain from a fluent chain.** The first walks strangers; the second walks the same friend.

---

Traces to: [`references/principles/law-of-demeter.md`](../../references/principles/law-of-demeter.md).
