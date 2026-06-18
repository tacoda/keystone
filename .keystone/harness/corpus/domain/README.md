# Domain

What the project *is* — its product shape, its business rules, the invariants the codebase exists to enforce. Distinct from idioms (how to write code) and state (what the code looks like right now).

## What lives here

- `product-shape.md` — what the project ships, what it doesn't, and what survives a release. Add it on first use.

Teams often add more:

- `glossary.md` — domain vocabulary that the agent should use consistently.
- `invariants.md` — business rules that must always hold (e.g., "an order cannot be both shipped and refunded").
- `personas.md` — user types the system serves.
- `<feature>.md` — domain notes for a feature area.

## Activation

Ambient, always loaded. Domain rules constrain agent behavior across every action.

## Empty by default

This directory ships with a `product-shape.md` template populated by the **bootstrap** action. Until then, the agent has only principles and process — no project-specific business context.

## Authorship

Domain expert + lead engineer. Agents may propose additions through the Learning flywheel, but domain claims must be confirmed by a human who owns the business knowledge.

## Changes when

The product changes. New invariant accepted. Business rule shifts. Vocabulary settles.
