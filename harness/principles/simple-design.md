# Simple Design

A design is "simple" not when it is small, but when it carries no weight it does not need. Articulated by Kent Beck in *Extreme Programming Explained* (1999) as four rules, applied in priority order, and extended in *Tidy First?* (2023) with the disposition of making the next change cheap before making it.

## The four rules of simple design

In priority order — earlier rules dominate later ones:

1. **Passes the tests.** A design that does not work is not a design; it is a wish.
2. **Reveals intention.** A reader should be able to tell what the code is for. Names, structure, and shape do the explaining; comments are a fallback.
3. **No duplication.** Every piece of behavior should have a single home. See [[pragmatic-principles]] (DRY) for the deeper formulation.
4. **Fewest elements.** Given the first three are satisfied, prefer the design with fewer classes, methods, branches, parameters, and concepts.

The order matters: a design with no duplication that fails tests is broken. A design that passes tests but lies about its intent is a debt that compounds.

## Make the change easy, then make the easy change

> For each desired change, make the change easy (warning: this may be hard), then make the easy change. — Kent Beck

When the next change is hard, the right first step is usually not "do the hard change carefully," but "rearrange the code so the change becomes easy." Beck calls the rearrangement a *tidying*; Fowler calls it *preparatory refactoring* (see [[refactoring]]).

## YAGNI — You Aren't Gonna Need It

A principle from Extreme Programming: do not add functionality until it is actually needed. Speculative generality is itself complexity. The bet is that the cost of adding the feature later, when you understand the real requirement, is lower than the cost of carrying the wrong feature now.

## What it asks of you

- When you finish a change, ask: *does this code reveal what it is for?* If not, rename, reshape, or extract — not later, now.
- When you are about to write a "configurable" or "extensible" component for a use case that does not yet exist, stop. Build the concrete thing. See [[modern-software-engineering]] on falsifiability.
- When the next change is hard, separate the *tidying* commit from the *behavior* commit. Two small commits beat one large one.
- When you have a choice between two correct designs, pick the one with fewer concepts.

## IRON LAW

**Simplicity is a property of the *next* reader, not the author.** The author already understands the code; their judgment that "this is simple" is uncalibrated. Simplicity is whether someone unfamiliar with the change can follow it.

## GOLDEN RULES

- **Aim for code that explains itself.** A comment that translates code into English is a sign the code was not written for the reader.
- **Aim for two commits, not one.** Tidying then behavior. Or behavior then tidying. Never both.
- **Aim for the smallest design that solves today's problem.** Tomorrow's problem will arrive with information you do not have yet.

## Anti-patterns

- A `FooFactoryStrategyManager` indirection layer for one concrete implementation.
- "We might need this later" — config flags, hook points, and extension seams with no current caller.
- Cleverness for its own sake — a one-line expression that takes ten minutes to read.
- A commit that mixes a rename, a refactor, and a behavior change.
- Tests that pass but do not pin the behavior the next reader cares about.

## References

- Beck, K. (1999, 2004). *Extreme Programming Explained: Embrace Change* (1st and 2nd eds.). Addison-Wesley.
- Beck, K. (2023). *Tidy First? A Personal Exercise in Empirical Software Design*. O'Reilly.
- Fowler, M. (2015). "BeckDesignRules." martinfowler.com/bliki/BeckDesignRules.html. (Fowler's articulation of the four rules and their ordering.)
- Jeffries, R., Anderson, A., & Hendrickson, C. (2000). *Extreme Programming Installed*. Addison-Wesley. (Early articulation of YAGNI as a practice.)
