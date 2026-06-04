# Hexagonal Architecture — rules

The rules from [`corpus/hexagonal.md`](../corpus/hexagonal.md).
Loaded ambient; enforced by the [drift sensor](../../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**Dependencies cross the hexagon boundary in one direction only — inward.** A package, file, or class inside the hexagon must not name any class, type, or function defined by an adapter. If you cannot ship the domain as a library that compiles without any adapter present, the architecture is not hexagonal.

## GOLDEN RULES

- **Aim for ports named in domain language.** They are the contract; they should read like the business describing itself.
- **Aim for one adapter per port per technology.** Adapters are how you swap technology; they are not where domain logic lives.
- **Aim for the domain to be testable with no adapters but in-memory fakes.** If domain tests need a real database to pass, the dependency direction is wrong.
- **Aim for the application's wiring to live at the outermost edge.** Composition root assembles adapters into ports at startup; nothing else constructs adapters.

---

Traces to: [`corpus/hexagonal.md`](../corpus/hexagonal.md).
