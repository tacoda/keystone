# Domain references

What the project is: its product shape, its business rules, and the invariants the code exists to enforce. Idioms say how to write code. State says what the code looks like now. Domain says what the code is for.

## What lives here

- `product-shape.md` ships as a template. The `bootstrap` command fills it in.

Teams often add more:

- `glossary.md`: the domain words the agent should use.
- `invariants.md`: business rules that must always hold, such as "an order can't be both shipped and refunded."
- `<feature>.md`: notes for one feature area.

When a domain fact must constrain code, write the directive as a rule in `.agents/rules/domain/` and point its `corpus:` here.

## Who writes it

A domain expert and the lead engineer. The agent can propose additions through `learn`, but a person who owns the business knowledge confirms each claim.
