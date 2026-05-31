# Coupling and Cohesion

Two complementary forces that determine whether a system is changeable. Introduced by Larry Constantine and Edward Yourdon in *Structured Design* (1979); the vocabulary predates object-orientation and applies regardless of paradigm.

## The forces

**Coupling** is the degree to which two modules depend on each other. Low coupling means a change in one module rarely forces a change in another. Constantine ranked coupling on a scale, from worst to best: content, common, control, stamp, data, message. Modern code rarely exhibits the worst forms in static analysis but reproduces them dynamically — shared mutable state, implicit global config, side-effects across module boundaries.

**Cohesion** is the degree to which the elements inside a module belong together. High cohesion means every part of a module exists to serve a single, well-defined responsibility. Constantine ranked cohesion, from worst to best: coincidental, logical, temporal, procedural, communicational, sequential, functional.

The goal: **loose coupling, high cohesion.** Modules that change for one reason and connect through minimal, explicit interfaces.

## What it asks of you

- When two modules pass large structs back and forth, ask whether the right module owns the operation (stamp coupling).
- When a flag parameter changes a function's behavior, ask whether you have two functions pretending to be one (low cohesion, control coupling).
- When you cannot reorder two modules' development without one depending on the other, look for hidden coupling.
- When a module's name covers many topics ("UtilHelper", "Manager", "Common"), cohesion is probably low.

## GOLDEN RULES

- **Aim for the smallest seam.** When two modules must communicate, pass only what is needed — not the whole object.
- **Aim for single-purpose modules.** The cohesion test: can you describe what this module does in one sentence without using "and"?
- **Aim for dependencies that point in one direction.** Cycles between modules guarantee that any change propagates.

## Anti-patterns

- A "Common" or "Shared" module everything imports.
- Global mutable state.
- Modules that depend on each other's private state.
- Boolean flag parameters that change a function's flow.
- "Utility" modules whose members have nothing to do with each other.

## References

- Yourdon, E., & Constantine, L. (1979). *Structured Design: Fundamentals of a Discipline of Computer Program and Systems Design*. Prentice Hall.
- Stevens, W., Myers, G., & Constantine, L. (1974). "Structured Design." *IBM Systems Journal*, 13(2).
- McConnell, S. (2004). *Code Complete*, 2nd ed. Microsoft Press. (Chapter 5 on design fundamentals.)
