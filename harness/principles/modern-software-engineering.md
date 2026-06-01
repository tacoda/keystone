# Modern Software Engineering

Software engineering is the application of an empirical, scientific approach to finding efficient, economic solutions to practical problems in software. Articulated by David Farley in *Modern Software Engineering: Doing What Works to Build Better Software Faster* (2021), the discipline rests on two pillars: being **experts at learning** and **experts at managing complexity**. Every other principle in this layer instantiates one or both.

## The two pillars

### Experts at learning
We do not know what the right answer is when we begin; we discover it. The techniques that make discovery cheap and reliable:

- **Iterative** — repeat the same activity, refining each time. Each pass is an experiment.
- **Incremental** — build the system in small slices that are integrated end-to-end.
- **Empirical** — base decisions on observation, not opinion or seniority.
- **Experimental** — form a hypothesis, design a way to falsify it, run the test.
- **Feedback** — close the loop fast. The longer the loop, the slower the learning and the more expensive the mistakes.

### Experts at managing complexity
Complexity is the chief enemy of software that is correct, cheap to change, and safe to operate. The techniques that keep complexity bounded:

- **Modularity** — see [[information-hiding]]. Decompose along design decisions, not along the flowchart.
- **Cohesion** — see [[coupling-cohesion]]. Each module exists for one reason.
- **Separation of concerns** — see [[separation-of-concerns]]. Different aspects live in different places.
- **Abstraction** — name the *what* in a way that lets the *how* change. See [[SOLID]] (DIP).
- **Coupling** — see [[coupling-cohesion]]. Minimize what a module must know about its neighbors.

## What it asks of you

- When a change is hard, ask whether the *learning loop* is slow (long feedback) or the *system* is complex (high coupling, low cohesion). The remedy is different in each case.
- When you find yourself arguing about which design is better in the abstract, design an experiment instead. Cheap evidence beats expensive opinion.
- When the integration step is large and rare, the increments are too large.
- When a module is hard to test in isolation, complexity is leaking across a boundary that should hold.

## IRON LAW

**Falsifiability before authority.** A design decision that cannot be tested — by a unit, an integration, a measurement in production, or a deliberate experiment — is a guess. Guesses are allowed; calling them engineering is not.

## GOLDEN RULES

- **Aim for short feedback loops.** Compile time, test time, integration time, deploy time, time-to-detect-in-prod. Each one shorter is each one of these loops tighter.
- **Aim for changes that are small enough to be safe and large enough to be useful.** Increment size is a tuning parameter, not a fixed quantity.
- **Aim for designs that admit being wrong.** Reversibility is a property worth paying for.

## Anti-patterns

- "We'll integrate everything at the end" — large, rare merges instead of continuous integration.
- "The senior engineer said so" — authority substituting for evidence.
- A test suite that takes long enough that no one runs it before pushing.
- A "big design up front" with no provision for what the team will learn while building.
- Treating production as the place where bugs are discovered rather than where hypotheses are confirmed.

## References

- Farley, D. (2021). *Modern Software Engineering: Doing What Works to Build Better Software Faster*. Addison-Wesley.
- Humble, J., & Farley, D. (2010). *Continuous Delivery: Reliable Software Releases through Build, Test, and Deployment Automation*. Addison-Wesley.
- Forsgren, N., Humble, J., & Kim, G. (2018). *Accelerate: The Science of Lean Software and DevOps*. IT Revolution Press.
- Beck, K. (1999). *Extreme Programming Explained: Embrace Change*. Addison-Wesley. (The earliest comprehensive articulation of short feedback loops as engineering practice.)
