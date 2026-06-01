# Behavior-Driven Development

Test names are sentences; tests are examples of behavior; the words you choose are shared with the people who decide what the system should do. Articulated by Dan North in *Introducing BDD* (*Better Software*, 2006), drawing on years of practice and conversation with Chris Matts. BDD is **TDD with the vocabulary fixed.** Where TDD's terminology ("test", "assert", "unit") obscures the intent, BDD's vocabulary ("behavior", "should", "example", "scenario") names it.

This file states the BDD principles; mechanical practice belongs in `../idioms/<stack>/`. Companion to [[tdd]] — most BDD practitioners are doing TDD; the difference is what the tests *say*, not when they get written.

## Why the renaming matters

North's central observation: when developers say "I am writing a *test* for this *unit*," they reach for the testing mindset — coverage, edge cases, can-this-break. When they say "I am writing an *example* of how this *should* *behave*," they reach for the specification mindset — what should happen, in what situation, for what reason. The shift in vocabulary shifts the design pressure. See [[tdd]] on test-first as design feedback.

A few of North's renamings, and what each does:

- *test* → **example** / **scenario.** A test sounds like verification; an example sounds like specification. Specifications are read by people who do not yet know how the system works.
- *assert* → **should.** "It should return the total" reads as a sentence to a non-engineer. "Assert that the result equals the total" reads as code.
- *unit* → **behavior.** A unit is a structural division; a behavior is something the system does. The system's users care about the second; the build system cares about the first.
- *test name* → **the first sentence of the spec.** If your test names do not form a readable specification when listed, the system does not have a spec, only a suite.

## Given-When-Then

The canonical structure for a BDD example, refined by Aslak Hellesøy and others into the Gherkin syntax used by Cucumber, SpecFlow, and similar tools:

- **Given** — the situation that holds before the behavior occurs. Setup; context; preconditions.
- **When** — the event whose behavior is being specified. The action the user, system, or upstream caller takes.
- **Then** — the observable outcome. What the system should do, in terms a stakeholder can verify.

Given-When-Then is a structured form of [[design-by-contract]]: *Given* is the precondition, *Then* is the postcondition, *When* is the operation under contract. The structure is also useful in plain unit tests with no BDD tooling at all — see [[testing-patterns]] on Arrange-Act-Assert, the same structure under a different name.

A scenario that does not fit Given-When-Then is usually a scenario that is doing too many things. The cure is to split, not to soften the structure.

## Ubiquitous language

BDD's vocabulary discipline extends beyond test names into the test bodies themselves: the **same words** that the business uses for an entity, an action, or a rule should appear in the test, in the code, and in the conversation. The idea originates in Eric Evans's *Domain-Driven Design* (2003) and is BDD's most important inheritance from DDD.

When the business says "order" and the test says "order" and the class is `Order`, a non-engineer can read the test. When the business says "order" and the test says "Cart.checkout(payload)", the seam between domain and code has already started to leak. See [[separation-of-concerns]].

## Outside-in, with examples first

BDD inherits TDD's outside-in flavor — see [[tdd]] on the London school — but adds a step **before** the first failing test: a conversation, often called *Three Amigos* (engineer, product, QA), that produces concrete examples of the behavior under discussion. The examples are the input to TDD; the tests that flow from them are the output.

Gojko Adzic's *Specification by Example* (2011) is the most thorough articulation of this practice: the examples are the spec, the spec is executable, the executable spec is the regression suite, and the regression suite is read by the people who originally wrote the examples.

## ATDD — the lineage BDD descended from

BDD did not invent the idea of writing tests from the user's perspective before the code; it inherited it. **Acceptance Test-Driven Development (ATDD)** had already been in practice for several years in the Extreme Programming community by the time North coined "BDD" — driven by Ward Cunningham's *Framework for Integrated Test* (FIT, 2002) and codified in Lisa Crispin and Janet Gregory's *Agile Testing* (2009).

The differences are real but small:

- **ATDD** centers on the **acceptance criteria** of a story: the user-visible behavior the team has agreed the system must exhibit before the story is *done*. The acceptance tests are the contract between developers and the customer. FIT/FitNesse tables are the archetypal artifact.
- **BDD** keeps everything ATDD has and adds the vocabulary shift: behavior-focused naming, Given-When-Then structure, ubiquitous language. The example-driven Three-Amigos workshop is shared with ATDD; the *words* in the resulting tests are tightened.
- Both sit *outside* the unit-test loop. The TDD loop is engineer-facing and runs in minutes; the ATDD/BDD acceptance loop is stakeholder-facing and runs at story or feature granularity.

The right mental model: ATDD and BDD are the **outer** loop (acceptance-level, stakeholder-readable, slow-but-stable); TDD is the **inner** loop (developer-level, fast, fine-grained). Most healthy teams run both — see [[testing-patterns]] on the test pyramid for how they nest, and [[tdd]] for the inner loop's mechanics.

A team that does ATDD without TDD ships features that satisfy the acceptance criteria with internals that resist change. A team that does TDD without ATDD builds well-tested units of something nobody asked for. The two loops are complements, not alternatives.

## What it asks of you

- When you name a test, write the name as a sentence about behavior. *"calculates tax for orders over $100"* beats *"test_tax_2"*.
- When you structure an example, reach for Given-When-Then (or AAA — see [[testing-patterns]]). If you cannot, the example is probably specifying more than one behavior.
- When you find that the test and the business use different words for the same thing, fix the test, the code, or the business vocabulary — whichever is wrong. Diverging vocabulary is technical debt with a long compounding interest rate.
- When a scenario reads as a sequence of system calls rather than a description of behavior, you are writing an integration test in BDD costume. Lift the language.
- When you cannot describe a scenario without reaching for the implementation, the implementation is the wrong abstraction. See [[information-hiding]].

## IRON LAW

**The test name is a promise to the next reader.** A test named for a behavior commits to specifying that behavior, and only that behavior. A test that asserts five unrelated things under one behavior name has broken the promise — split it. A test named *test_method_2* has made no promise at all, which is worse.

## GOLDEN RULES

- **Aim for tests that read aloud as English.** If the test cannot be read to a non-engineer, it is not yet a specification — it is still a test.
- **Aim for examples produced *with* the people who care about the behavior, not for them.** A spec written alone is a spec rewritten under pressure.
- **Aim for the system's vocabulary to match the business vocabulary.** Each translation between them is a place for the truth to drift.
- **Aim for examples that are the regression suite.** When the same artifacts serve discovery, verification, and prevention-of-regression, the cost of maintenance drops to one suite from three.

## Anti-patterns

- Cucumber / Gherkin used as a slow, awkward replacement for unit tests — Given/When/Then steps that mirror code calls, with no business reader in mind.
- Test names like *test_1*, *test_method*, *testOrderHappyPath* — names that specify nothing.
- A Given that requires three paragraphs of setup. The behavior is probably tangled with something else.
- Scenarios that read as UI scripts ("click the button, then click the next button") rather than behavior ("the order is placed").
- A glossary mismatch — code says *Customer*, business says *Account*, tests say *User*. Pick one. Change the others. Today.
- BDD adopted as a tool stack without the conversation. Cucumber without Three Amigos is a slower unit-test framework.

## References

- North, D. (2006). "Introducing BDD." *Better Software*. dannorth.net/introducing-bdd/. (The canonical statement.)
- Adzic, G. (2011). *Specification by Example: How Successful Teams Deliver the Right Software*. Manning. (The practice as a software-delivery discipline.)
- Wynne, M., Hellesøy, A., & Tooke, S. (2017). *The Cucumber Book: Behaviour-Driven Development for Testers and Developers*, 2nd ed. Pragmatic Bookshelf.
- Smart, J. F. (2014). *BDD in Action: Behavior-Driven Development for the Whole Software Lifecycle*. Manning.
- Evans, E. (2003). *Domain-Driven Design: Tackling Complexity in the Heart of Software*. Addison-Wesley. (Origin of ubiquitous language, which BDD operationalizes in tests.)
- Matts, C., & North, D. (2009). "Feature Injection." (The conversation-first half of the BDD practice.)
- Crispin, L., & Gregory, J. (2009). *Agile Testing: A Practical Guide for Testers and Agile Teams*. Addison-Wesley. (The canonical ATDD reference; the "Agile Testing Quadrants" model.)
- Crispin, L., & Gregory, J. (2014). *More Agile Testing: Learning Journeys for the Whole Team*. Addison-Wesley.
- Cunningham, W. (2002). *Framework for Integrated Test (FIT)*. fit.c2.com. (The original acceptance-testing framework; FitNesse is its surviving descendant.)
- Pugh, K. (2010). *Lean-Agile Acceptance Test-Driven Development*. Addison-Wesley. (ATDD as a delivery discipline.)
