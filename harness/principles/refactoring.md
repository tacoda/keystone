# Refactoring

A disciplined technique for restructuring an existing body of code, **altering its internal structure without changing its external behavior**. Articulated by Martin Fowler in *Refactoring: Improving the Design of Existing Code* (1999; 2nd ed., 2018), building on practice in the Smalltalk community and Ward Cunningham's work. Refactoring is not "cleanup"; it is a sequence of named, small, behavior-preserving transformations applied for a reason.

## The discipline

A refactoring has three parts:

1. **A name.** *Extract Function*, *Inline Variable*, *Replace Conditional with Polymorphism*. Named transformations are reproducible; ad-hoc rewrites are not.
2. **A mechanics.** A small, ordered sequence of steps. Each step leaves the code working.
3. **A reason.** Refactoring without a smell to chase is movement without progress.

The bet behind refactoring: a large restructuring done as a sequence of small, tested steps is *safer* than the same change done in one leap, even though it appears to take longer. The safety comes from never being more than one undo away from a working state.

## The two-hat rule

> When you use refactoring to develop software, you divide your time between two distinct activities: adding function and refactoring. When you add function, you shouldn't be changing existing code; you're just adding new capabilities. … When you refactor, you make a resolution not to add any function; you only restructure the code. — Fowler

You wear one hat at a time. Switch often; never wear both at once. A commit that adds a feature and reshapes the code around it conflates two failure modes — a regression could be in either, and review cannot tell them apart.

This is the same disposition as Kent Beck's "make the change easy, then make the easy change" — see [[simple-design]].

## The Rule of Three

> Three strikes and you refactor. — Don Roberts, quoted by Fowler in *Refactoring*

The first time you do something, just do it. The second time you do something similar, wince at the duplication but do it anyway. The third time, refactor — extract the abstraction, name it, and replace all three sites.

The rule guards against premature abstraction. Two examples are not enough information to know which parts vary and which stay the same; the abstraction you build from two cases routinely fits the third worse than the duplication did. Three concrete cases give you enough variation to draw the right line. See [[simple-design]] (YAGNI) and [[simplicity]] — abstractions are complexity, paid for in change-cost; the third occurrence is the receipt that says the cost is worth paying.

## Smells as triggers

A *code smell* is a surface indicator — duplicated code, long function, large class, long parameter list, divergent change, shotgun surgery, feature envy, primitive obsession, switch statements, parallel inheritance hierarchies, lazy class, speculative generality, message chains, middle man. Each smell maps to a refactoring (or several) that addresses it.

The smell tells you *something* is off; the catalog tells you *what to do about it*. Refactoring without a smell is fiddling.

## What it asks of you

- When the next feature is hard to add, ask whether a refactoring would make it easy. If yes, do the refactoring first — in its own commit. See [[simple-design]].
- When you notice a smell, name it. "This is shotgun surgery" is a more actionable observation than "this is messy."
- When you refactor, run the tests after every step. If a test breaks, the step was wrong; revert and try a smaller one.
- When the test coverage is too thin to refactor safely, write characterization tests first. They pin the *current* behavior so the refactoring can preserve it.

## The Boy Scout Rule

> Always leave the campground cleaner than you found it. — popularized for code by Robert C. Martin, *Clean Code* (2008)

A maxim about *opportunistic* improvement: when you are already in a file for a real reason, small cleanups along the way are nearly free. A renamed variable, a clarified comment, an extracted local — these compound across a codebase even though each individual change is trivial. The opposite — stepping over the smell because "it isn't what I'm here to fix" — is how a codebase silently degrades, the lesson behind [[pragmatic-principles]]' broken-windows section.

**The caveat that matters.** The Boy Scout Rule, applied without restraint, collides with the two-hat rule above. A "small cleanup" smuggled into a behavior-changing commit is exactly the failure mode the two-hat rule exists to prevent — review can no longer tell whether a regression came from the feature or the tidy. The way to honor both:

- **Yes** — opportunistically improve files you are already in.
- **No** — do not pack the improvement into the *same commit* as a behavior change.
- **Either** — a separate tidying commit before the behavior commit, or a separate cleanup PR after. Pick by the size of the cleanup and the patience of the review queue.

The Boy Scout Rule is a disposition; the two-hat rule is a commit-shape rule. Both are correct. The slogan that puts them on the same side: *tidy first, then change behavior — never together.* See [[simple-design]] ("make the change easy, then make the easy change").

## IRON LAW

**Behavior preservation is not optional.** A "refactoring" that changes behavior is not a refactoring; it is a bug-shaped change wearing the wrong hat. If you discover during refactoring that the existing behavior is wrong, stop, finish the refactoring, commit it, and then fix the behavior in a separate commit.

## GOLDEN RULES

- **Aim for steps small enough that the test suite stays green between each.** If two steps must land together to keep tests passing, the step was too large.
- **Aim for refactorings with names.** The catalog is large enough that nearly every move you want to make has a name. Use it.
- **Aim for separate commits for tidying and behavior.** Reviewers can read either kind quickly; the mixed kind is slow and error-prone.

## Anti-patterns

- "Refactor" used as a synonym for "rewrite" — a multi-week, behavior-changing replacement masquerading as a refactor.
- Mixing rename + reshape + bug fix in one commit.
- Refactoring without tests, on the hope that "it should be equivalent."
- Stopping mid-refactoring on a broken intermediate state and pushing it.
- Refactoring code that nothing is asking you to change. See [[simple-design]] (YAGNI applies to design too).

## References

- Fowler, M. (2018). *Refactoring: Improving the Design of Existing Code* (2nd ed.). Addison-Wesley.
- Fowler, M. (1999). *Refactoring: Improving the Design of Existing Code* (1st ed.). Addison-Wesley. (Original; introduces the catalog and the two-hat rule.)
- Beck, K. (2023). *Tidy First? A Personal Exercise in Empirical Software Design*. O'Reilly. (Beck's complementary framing: tidyings as small, cheap, preparatory restructurings.)
- Feathers, M. (2004). *Working Effectively with Legacy Code*. Prentice Hall. (Refactoring under conditions of low test coverage — characterization tests, seams.)
- Opdyke, W. F. (1992). *Refactoring Object-Oriented Frameworks*. PhD thesis, University of Illinois. (The original academic treatment.)
- Roberts, D. — quoted by Fowler in *Refactoring* (1999) as the source of *Three strikes and you refactor*.
- Martin, R. C. (2008). *Clean Code: A Handbook of Agile Software Craftsmanship*. Prentice Hall. (Source of the Boy Scout Rule as a programming maxim.)
