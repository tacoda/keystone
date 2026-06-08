# Simplicity

Simplicity is the value behind nearly every other principle in this layer. Stated most directly by Edsger Dijkstra: *simplicity is a prerequisite for reliability.* You cannot reason about, test, secure, or safely change what you cannot hold in your head. This file is about simplicity as a **value**; [[simple-design]] is about Kent Beck's mechanics for achieving it.

> **Rules extracted:** [`guides/principles/simplicity.md`](../../guides/principles/simplicity.md). This file holds the full reasoning, anti-patterns, and references.

## The canonical statements

> Simplicity is prerequisite for reliability. — Edsger W. Dijkstra (EWD 498, 1975)

> There are two ways of constructing a software design: One way is to make it so simple that there are obviously no deficiencies, and the other way is to make it so complicated that there are no obvious deficiencies. The first method is far more difficult. — C. A. R. Hoare (Turing Award lecture, 1980)

> The greatest limitation in writing software is our ability to understand the systems we are creating. — John Ousterhout (*A Philosophy of Software Design*, 2018)

Ousterhout defines complexity operationally: *anything related to the structure of a software system that makes it hard to understand and modify the system.* The symptoms are **change amplification** (one logical change requires many code changes), **cognitive load** (how much a developer must know to make a change), and **unknown unknowns** (it is not even obvious what one needs to know).

## KISS

The engineering-folk version, originating with Kelly Johnson at Lockheed's Skunk Works in the late 1950s: **Keep It Simple, Stupid** — the comma is original, and "stupid" addresses the reader, not the design. Johnson's working principle was that a jet aircraft had to be repairable by an average mechanic in a combat zone with basic tools; any design that demanded more was the wrong design, regardless of how clever it was on paper.

The same constraint applies to software: a system has to be debuggable on a Friday night by an on-call engineer who did not write it. Designs that require expertise the team does not have on duty are designs that will fail when it matters. KISS is the same value as Dijkstra's, Hoare's, and Ousterhout's — stated in fewer syllables, but no less serious. See [[simple-design]] for the mechanics of getting there.

## What it asks of you

- When you are about to add a configuration knob, a generic parameter, a plugin point, or a layer of indirection, ask whether the problem actually exists yet. Speculative complexity is the most common kind. See [[simple-design]] (YAGNI).
- When you find that a change requires editing five files, the design is amplifying the change. Push back on the design, not on the change. See [[pragmatic-principles]] (orthogonality).
- When a reviewer cannot follow the diff without a tour from the author, the code is not simple — regardless of how elegant it looks to the author. See [[simple-design]]'s second rule (reveals intention).
- When you cannot explain what a module does in one sentence without "and," cohesion is low and complexity is high. See [[coupling-cohesion]].

## Anti-patterns

- "We might need this later." Every speculative feature, abstraction, hook, and config flag carries this label until paid. Most are never collected. See [[simple-design]].
- A class hierarchy whose levels are named *abstract*, *generic*, *base*, *core*, *common*, *default* — the names confess that the levels are not modeling anything real.
- Naming that requires a glossary to read the code.
- A "framework" inside the application that no one outside the team will ever use.
- Defending a design by its cleverness rather than its clarity.

## References

- Dijkstra, E. W. (1975). *How do we tell truths that might hurt?* EWD 498. (Source of "Simplicity is prerequisite for reliability.")
- Hoare, C. A. R. (1981). "The Emperor's Old Clothes." *Communications of the ACM*, 24(2), 75–83. (Hoare's 1980 Turing Award lecture, where the two-ways quote appears.)
- Ousterhout, J. (2018). *A Philosophy of Software Design*. Yaknyam Press. (The most thorough modern treatment of complexity as the central engineering concern.)
- Brooks, F. P. (1986). "No Silver Bullet — Essence and Accident in Software Engineering." *IEEE Computer*. (The classic distinction between essential and accidental complexity. The simplicity discipline is about minimizing accidental complexity; essential complexity is irreducible.)
