# Domain rules

Business-rule constraints extracted from `corpus/domain/`. Where corpus describes what the project *is*, this directory holds the rules the agent must obey when writing code that touches the domain.

## What lives here

Teams typically add files such as:

- `invariants.md` — non-negotiable business rules ("an order cannot be both shipped and refunded").
- `vocabulary.md` — terminology rules ("never call a `Customer` a `User` in surfaced text").
- `<feature>.md` — feature-specific rules.

Each file traces to a corpus counterpart that explains *why* the rule exists.

## Empty by default

This directory ships **empty** in a fresh install. The **bootstrap** action seeds `corpus/domain/product-shape.md`; rules emerge through the Learning flywheel as the agent encounters domain constraints in the codebase.

## Activation

**Default** — ambient, every action. Domain rules constrain agent behavior across every action.

**With `globs:`** — narrows that default to a subset of paths. A domain rule with a clear regional boundary (billing, auth, search) keeps context lean by firing only where it applies:

```markdown
---
globs:
  - "src/billing/**"
  - "tests/billing/**"
---
# Billing invariants — rules
```

The above is silent on files outside `src/billing/**` and `tests/billing/**`. Domain rules without `globs:` keep today's behavior (always loaded). See [`../README.md`](../README.md) for the full narrow-only contract.

## Format

Each rule file:

```markdown
# <Rule Set> — rules

The rules from [`corpus/domain/<file>.md`](corpus/domain/<file>.md).

## IRON LAWS

- **<RULE NAME IN CAPS.>** One-sentence rule. (When applicable.)

## GOLDEN PATH

- **<Rule name.>** Strongly-preferred behavior, deviation requires reasoning.
```

## Authorship

Domain expert + lead engineer. Agents propose additions via the Learning flywheel; humans gate every domain rule before it lands here — domain claims must be confirmed by someone who owns the business knowledge.

## Changes when

The product does. New invariant accepted, business rule shifts, vocabulary settles.
