# Principles

Universal, language-agnostic engineering truths. SOLID, coupling/cohesion, information hiding, separation of concerns. The kind of thing that shows up in foundational literature and stays true across decades.

## What lives here

Files scaffolded at install time, expanded by the **bootstrap** action:

- `SOLID.md`
- `coupling-cohesion.md`
- `separation-of-concerns.md`
- `information-hiding.md`

You may add more if your team has principles that meet the bar: universal, language-agnostic, cited. If a "principle" is actually stack-specific, it belongs in `../idioms/<stack>/`. If it's business-specific, it belongs in `../domain/`.

## Activation

Ambient, always loaded. Principles set the floor for every decision the agent makes.

## Authorship

Drawn from literature; refined through discipline. The harness is opinionated about what counts as a principle.

## Format

Each principle file follows the same shape:

```markdown
# <Principle Name>

One-paragraph statement of the principle.

## What it asks of you

The behavioral implications — what the agent should do or avoid.

## Why it holds

The reasoning. Where the principle was first articulated; the canonical source.

## Anti-patterns

What it looks like when this principle is violated.

## References

Books, papers, talks. Real ones — not made-up citations.
```

## Conventions

- **IRON LAW** / **GOLDEN RULES** as appropriate.
- No `Traces to:` footer — principles are root nodes; idioms trace to *them*.
- No `## Project-Specific Notes` — principles are universal by definition.

## Changes when

Almost never. A principle file changing is rare. If you find yourself editing one often, you've probably mis-classified an idiom as a principle.
