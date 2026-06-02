# Object-Oriented Design Heuristics

Three guidelines that, taken together, describe what *good* object-oriented design looks like at the level of how objects relate to each other. Drawn from the Gang of Four (*Design Patterns*, 1994), Alec Sharp's work on object collaboration, and Hunt & Thomas's *The Pragmatic Programmer*.

These are the operational counterparts to [[SOLID]]: SOLID names the principles; these name the moves you actually make at the keyboard.

> **Rules extracted:** [`guides/principles/object-oriented-design.md`](../../guides/principles/object-oriented-design.md). This file holds the full reasoning, anti-patterns, and references.

## Tell, don't ask

> Procedural code gets information, then makes decisions. Object-oriented code tells objects to do things. — Alec Sharp

When you find yourself pulling data out of an object to make a decision about that object, the decision belongs *inside* the object. Asking exposes internals; telling preserves the seam.

```
// Asking
if (account.getBalance() >= amount) {
    account.setBalance(account.getBalance() - amount);
}

// Telling
account.withdraw(amount);
```

The "telling" form hides the rule that balance must remain non-negative. The "asking" form leaks it to every caller, where it can drift, be forgotten, or be enforced differently. See [[information-hiding]] for the deeper principle.

Tell-don't-ask is closely related to — but weaker than — the [[law-of-demeter]]. Demeter is the formal coupling rule; tell-don't-ask is the everyday disposition that tends to satisfy it.

## Program to an interface, not an implementation

> Program to an interface, not an implementation. — Gamma, Helm, Johnson, Vlissides, *Design Patterns*

A caller should depend on the **abstract contract** of what it uses, not on the concrete class behind the contract. The contract is what is stable; the implementation is what changes. This is the same force named in [[SOLID]] (Dependency Inversion) and in [[information-hiding]] (a module's interface is its contract; its implementation is its secret).

The mechanical test: search your code for the concrete type name. If callers reference `PostgresUserRepository` instead of `UserRepository`, the abstraction is not load-bearing — swapping the database means editing every caller.

"Interface" here is broader than the language-level `interface` keyword. It is whatever abstract contract the type system, the duck-typing convention, or the team's discipline can enforce.

## Favor composition over inheritance

> Favor object composition over class inheritance. — Gamma, Helm, Johnson, Vlissides, *Design Patterns*

Inheritance is the strongest form of coupling that mainstream OO offers. A subclass depends on the **substance** of its parent — fields, method bodies, invocation order, protected methods — not just the parent's interface. Change the parent and the subclasses break in ways the compiler cannot find.

Composition couples *much* more weakly: a class that *has-a* `Logger` depends only on `Logger`'s contract, not on its substance. Behavior is assembled by wiring small pieces together, not by inheriting a position in a tree.

The standard guidance:

- Use **interface inheritance** (subtyping) to declare a contract.
- Use **composition** to share implementation.
- Reach for **implementation inheritance** only when (a) the *is-a* relationship is truly stable, (b) LSP holds — see [[SOLID]] — and (c) the subclass would otherwise repeat substantial code that has no other reasonable home.

Deep inheritance hierarchies are a well-documented smell — see [[refactoring]].

## What it asks of you

- When you write `if (x.getY() ...) { x.setY(...) }`, replace it with a method on `x` that does both. The method names the rule; the rule has a home.
- When a class references a concrete type that has — or should have — an interface, route the dependency through the interface. Construction is the *only* place concrete types should appear.
- When you reach for `extends`, pause. Ask whether composition (a field of that type) would do. Most of the time, it would.
- When a subclass overrides a method to throw `UnsupportedOperationException`, the inheritance is wrong — see [[SOLID]] (LSP).

## Anti-patterns

- "Anemic domain model" — entities are bags of fields; business logic lives in service classes that operate on them.
- A caller that fetches a value, computes on it, and writes it back: the object was treated as a struct, not as an object.
- `extends AbstractFooBase` where the base class has eight protected methods, three abstract, and a template-method workflow nobody can hold in their head.
- A constructor that takes `PostgresUserRepository` directly instead of `UserRepository`.
- Inheritance used purely for code reuse, with no *is-a* relationship that survives a moment's scrutiny.

## References

- Gamma, E., Helm, R., Johnson, R., & Vlissides, J. (1994). *Design Patterns: Elements of Reusable Object-Oriented Software*. Addison-Wesley. (Introduces both "program to an interface" and "favor composition over inheritance" as the book's two opening principles.)
- Sharp, A. (1997). *Smalltalk by Example*. McGraw-Hill. (Earliest written articulation of "tell, don't ask"; the phrase is widely attributed to Sharp's teaching from this period.)
- Hunt, A., & Thomas, D. (1999). *The Pragmatic Programmer*. Addison-Wesley. (Popularizes tell-don't-ask alongside the Law of Demeter.)
- Fowler, M. (2003). "TellDontAsk." martinfowler.com/bliki/TellDontAsk.html.
- Riel, A. (1996). *Object-Oriented Design Heuristics*. Addison-Wesley. (A catalog of OO design rules; companion to the GoF style.)
