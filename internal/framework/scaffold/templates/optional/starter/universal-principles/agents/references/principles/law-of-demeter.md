# Law of Demeter

Each unit should have only limited knowledge about other units: only units "closely related" to the current one. Articulated by Karl Lieberherr, Ian Holland, and Arthur Riel at Northeastern University in 1988, while working on the Demeter Project — a tool for adaptive programming that needed a formal coupling rule. Also called the **Principle of Least Knowledge**, or informally, *"don't talk to strangers."*

> **Rules extracted:** [`guides/principles/law-of-demeter.md`](../../guides/principles/law-of-demeter.md). This file holds the full reasoning, anti-patterns, and references.

## The formal rule

A method *M* of an object *O* may only invoke methods of:

1. *O* itself.
2. *M*'s parameters.
3. Objects *M* creates / instantiates.
4. *O*'s direct component objects (its fields).
5. Global objects accessible by *O* in the relevant scope.

The rule forbids *M* from calling methods on objects returned by methods on any of the above. The classic violation is the **train wreck**:

```
a.getB().getC().doSomething()
```

The caller of `a` now knows about `B`, `C`, and the path through them. Any change to `B`'s shape — adding a layer, renaming `getC` — forces a change at every call site that traverses the chain.

## What it really says

Demeter is a rule about *coupling*, not about syntax. The deeper claim: a method should only know the **immediate collaborators** it was given or owns. Reaching through them couples you to their internals.

A common misreading is "no chained calls ever." Fluent interfaces (`query.where(...).orderBy(...).limit(...)`) and Optional/Maybe pipelines do not violate Demeter — each call returns the same conceptual object, not a different stranger. The test is *what objects am I now coupled to,* not *how many dots are in the line.*

## What it asks of you

- When you find yourself writing `x.getY().getZ()`, ask whether *x* should expose the operation you actually want. Tell *x* what to do; let it walk its own structure. (Closely related: tell-don't-ask in [[object-oriented-design]].)
- When a caller knows three layers of object structure to do its job, the middle layers are leaking. Either move the operation inward, or introduce a method that hides the traversal.
- When a refactoring renames an inner field and breaks unrelated callers, those callers were violating Demeter — they should not have known about the inner field.

## Anti-patterns

- `order.getCustomer().getAddress().getZipCode()` — caller now coupled to `Order`, `Customer`, `Address`.
- A controller that pulls fields off a domain object, runs business logic on them, and writes the result back — see [[separation-of-concerns]].
- "Helper" code that takes a deeply-nested object and digs through it. Push the helper's logic into the object that owns the data.
- Returning internal collections so callers can mutate them — see [[information-hiding]].

## When the rule bends

Pure-data structures (value objects, DTOs, configuration trees) are exempt by intent — their *whole purpose* is to be navigated. Demeter applies to objects with **behavior**. A `Money` or `Address` value object can be reached through without smell; a `Service` or `Aggregate` cannot.

## References

- Lieberherr, K., Holland, I., & Riel, A. (1988). "Object-Oriented Programming: An Objective Sense of Style." *OOPSLA '88 Conference Proceedings*. (The original formulation.)
- Lieberherr, K., & Holland, I. (1989). "Assuring Good Style for Object-Oriented Programs." *IEEE Software*, 6(5), 38–48.
- Hunt, A., & Thomas, D. (1999). *The Pragmatic Programmer*. Addison-Wesley. (Chapter on Demeter; couples it with tell-don't-ask.)
- Fowler, M. (2003). "GetterEradicator." martinfowler.com/bliki/GetterEradicator.html. (Argues the principle is about behavior, not getters per se.)
