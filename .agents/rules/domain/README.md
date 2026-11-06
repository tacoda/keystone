# Domain rules

Business-rule constraints. The domain references in `.agents/references/domain/` describe what the project is. The rules here say what the agent must obey when its code touches that domain.

## What lives here

Teams typically add files such as:

- `invariants.md`: rules that must always hold, such as "an order can't be both shipped and refunded."
- `vocabulary.md`: terminology rules, such as "never call a `Customer` a `User` in surfaced text."
- `<feature>.md`: rules for one feature area.

This folder ships empty. Rules arrive through `learn` and `synthesize` as the agent meets domain constraints in the code.

## Scope with paths

A domain rule with a clear boundary should declare `paths:`, so it applies only there:

```markdown
---
id: domain/billing
description: Billing invariants.
paths:
  - "src/billing/**"
  - "tests/billing/**"
corpus:
  - domain/billing
---

# Billing invariants

## Iron law

- An invoice total never changes after it is issued.
```

A domain rule without `paths:` applies everywhere.

## Who writes it

A domain expert and the lead engineer. The agent can propose a rule, but a person who owns the business knowledge confirms it before it lands.
