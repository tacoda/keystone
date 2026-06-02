# Principles

Universal, language-agnostic engineering truths. SOLID, coupling/cohesion, information hiding, separation of concerns. The kind of thing that shows up in foundational literature and stays true across decades.

> **Rules extracted:** [`guides/principles/`](../../guides/principles/) — every principle here has a paired guide carrying the IRON LAW(S) and GOLDEN RULES. This directory is the informational side: reasoning, anti-patterns, references.

## What lives here

Principles grouped by what they address. Files marked **(scaffold)** ship with the harness at install time and are expanded by the **bootstrap** action; the rest are team-curated additions that meet the same bar.

**Modularity & structure**
- `information-hiding.md` *(scaffold)* — Parnas: hide design decisions behind stable interfaces.
- `coupling-cohesion.md` *(scaffold)* — Constantine & Yourdon: loose coupling, high cohesion.
- `separation-of-concerns.md` *(scaffold)* — Dijkstra: address different aspects in different places.
- `SOLID.md` *(scaffold)* — Martin: five class-/module-level design principles.

**Object-oriented design**
- `object-oriented-design.md` — GoF / Sharp: tell-don't-ask, program to interfaces, composition over inheritance.
- `law-of-demeter.md` — Lieberherr: principle of least knowledge; don't talk to strangers.
- `design-by-contract.md` — Meyer: preconditions, postconditions, invariants.

**Simplicity & evolution**
- `simplicity.md` — Dijkstra, Hoare, Ousterhout, KISS: simplicity as a value.
- `simple-design.md` — Beck: four rules of simple design; make the change easy.
- `refactoring.md` — Fowler: behavior-preserving restructuring; two-hat rule; Rule of Three; Boy Scout Rule (with caveat).
- `pragmatic-principles.md` — Hunt & Thomas: DRY, orthogonality, ETC, broken windows.
- `naming.md` — McConnell, Martin, Karlton: names are promises; names are part of the contract.

**Engineering discipline**
- `modern-software-engineering.md` — Farley: experts at learning, experts at managing complexity.
- `premature-optimization.md` — Knuth: measure before optimizing; the full quote.
- `fail-fast.md` — Shore: detect failures at the source; never continue past a violated invariant.
- `error-handling.md` — Liskov, Stroustrup, Sutter, Armstrong: exception safety guarantees; propagation vs. swallowing; expected vs. exceptional.
- `least-astonishment.md` — Cowlishaw / Raymond: behave the way the next reader expects.
- `postels-law.md` — Postel, with Allman/Thomson's modern critique: the robustness principle on a tighter leash.
- `hyrums-law.md` — Wright: the contract is what callers can observe, not what you documented.

**Production & distributed systems**
- `concurrency.md` — Hoare, Pike, Lea, Goetz: shared mutable state, happens-before, message passing over shared memory.
- `distributed-systems-fallacies.md` — Deutsch / Gosling: the eight things everyone wrongly assumes about the network.
- `observability.md` — Sridharan, Majors: ability to answer new questions about production without redeploying.
- `idempotency.md` — Helland; RFC 9110: at-least-once delivery of idempotent operations as the working substitute for "exactly once."

**Testing**
- `tdd.md` — Beck: Red → Green → Refactor; test-first as design feedback; the test pyramid; F.I.R.S.T.
- `bdd.md` — North: behavior over tests; Given-When-Then; ubiquitous language. Includes the ATDD lineage (Cunningham, Crispin & Gregory).
- `testing-patterns.md` — Chicago-style classicist testing; mock only at real boundaries; test doubles, AAA, test data builders, characterization tests.

**Security**
- `security.md` — Saltzer & Schroeder: eight foundational protection principles.
- `security-threats.md` — OWASP Top 10 (2021) + modern threats (supply chain, cloud/IAM, confused deputy, prompt injection).
- `secrets-management.md` — secret lifecycle: generation, distribution, rotation, revocation; the hierarchy from workload identity down to "never in source."

You may add more if your team has principles that meet the bar: universal, language-agnostic, cited. If a "principle" is actually stack-specific, it belongs in `../idioms/<stack>/` (plus `../../guides/idioms/<stack>/` if it has rules). If it's business-specific, it belongs in `../domain/`.

## Activation

Ambient, always loaded. Principles set the floor for every decision the agent makes.

## Authorship

Drawn from literature; refined through discipline. The harness is opinionated about what counts as a principle.

## Format

Each principle file follows the same shape — explanatory only, no rule sections:

```markdown
# <Principle Name>

One-paragraph statement of the principle.

> **Rules extracted:** [`guides/principles/<file>.md`](../../guides/principles/<file>.md).

## What it asks of you

The behavioral implications — what the agent should do or avoid.

## Why it holds

The reasoning. Where the principle was first articulated; the canonical source.

## Anti-patterns

What it looks like when this principle is violated.

## References

Books, papers, talks. Real ones — not made-up citations.
```

The rule sections (IRON LAW(S) / GOLDEN RULES) live in the paired file under `guides/principles/`.

## Conventions

- **IRON LAW** / **GOLDEN RULES** belong in the paired guide, not here.
- No `Traces to:` footer — principles are root nodes; idioms trace to *them*.
- No `## Project-Specific Notes` — principles are universal by definition.

## Changes when

Almost never. A principle file changing is rare. If you find yourself editing one often, you've probably mis-classified an idiom as a principle.
