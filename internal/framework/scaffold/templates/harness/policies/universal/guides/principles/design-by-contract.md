# Design by Contract — rules

The non-negotiable and strongly-preferred rules from [`corpus/principles/design-by-contract.md`](../../corpus/principles/design-by-contract.md).
Loaded ambient; enforced by the [drift sensor](../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**Trust your inputs at the boundary; trust your invariants inside.** Validate at the seam between trusted and untrusted code (user input, network, external API). Inside the trust boundary, assume the contract holds, and assert when it does not. Validating everywhere both hides the real boundary and makes every method harder to read.

## GOLDEN RULES

- **Aim for contracts that the type system can enforce.** Types are checked contracts; comments are wishes.
- **Aim to make impossible states unrepresentable.** A user with `age = -3` should not be constructible.
- **Aim for invariants that constructors establish and methods preserve.** If a public method can leave the object in an invalid state, the contract is wrong.

---

Traces to: [`corpus/principles/design-by-contract.md`](../../corpus/principles/design-by-contract.md).
