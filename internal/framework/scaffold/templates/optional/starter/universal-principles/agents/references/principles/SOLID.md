# SOLID

Five class- and module-level design principles, articulated by Robert C. Martin in the 1990s and consolidated in *Agile Software Development: Principles, Patterns, and Practices* (2002). Applied together, they keep software changeable.

> **Rules extracted:** [`guides/principles/SOLID.md`](../../guides/principles/SOLID.md). This file holds the full reasoning, anti-patterns, and references.

## The five

### Single Responsibility Principle (SRP)
A module should have one, and only one, reason to change. "Reason to change" is shorthand for "stakeholder whose evolving needs reshape this code." Two unrelated stakeholders touching the same module is the canonical SRP smell.

### Open/Closed Principle (OCP)
Software entities should be open for extension, but closed for modification. New behavior is added by composing or extending existing abstractions, not by editing them in place. The mechanism is abstraction; the test is "does adding feature X require editing module Y?"

### Liskov Substitution Principle (LSP)
Subtypes must be substitutable for their base types without altering the correctness of the program. Barbara Liskov's original 1987 formulation (*Data Abstraction and Hierarchy*) is about behavioral compatibility, not just type compatibility — a subclass that raises new exceptions or violates preconditions breaks LSP even if the compiler accepts it.

### Interface Segregation Principle (ISP)
Clients should not be forced to depend on methods they do not use. Many small, focused interfaces beat one large, general one. Fat interfaces are coupling in disguise.

### Dependency Inversion Principle (DIP)
High-level modules should not depend on low-level modules; both should depend on abstractions. Abstractions should not depend on details; details should depend on abstractions. The direction of the source-code dependency is inverted relative to the direction of the flow of control.

## What it asks of you

- When you reach for inheritance, prove LSP holds before committing.
- When a class accumulates unrelated responsibilities, split it before the next feature.
- When you find a `switch` over types, ask whether OCP is being violated.
- When a client only uses two of fifteen methods on an interface, ISP wants you to split the interface.
- When high-level policy reaches down into low-level details, invert the dependency.

## Anti-patterns

- "God class" — one module with everything in it (SRP).
- Editing every existing case in a `switch` to add a new case (OCP).
- A subclass overriding a method to throw `NotImplementedException` (LSP).
- An interface every client implements but uses only a third of (ISP).
- An application layer that `new`s up a concrete database driver (DIP).

## References

- Martin, R. C. (2002). *Agile Software Development: Principles, Patterns, and Practices*. Prentice Hall.
- Martin, R. C. (2017). *Clean Architecture*. Prentice Hall.
- Liskov, B. (1987). "Data Abstraction and Hierarchy." *OOPSLA '87 Addendum*.
- Meyer, B. (1988). *Object-Oriented Software Construction*. Prentice Hall. (Source of the Open/Closed Principle in its original form.)
