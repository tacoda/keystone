# Modern Software Engineering — rules

The non-negotiable and strongly-preferred rules from [`corpus/principles/modern-software-engineering.md`](../../corpus/principles/modern-software-engineering.md).
Loaded ambient; enforced by the [drift sensor](../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**Falsifiability before authority.** A design decision that cannot be tested — by a unit, an integration, a measurement in production, or a deliberate experiment — is a guess. Guesses are allowed; calling them engineering is not.

## GOLDEN RULE

- **Aim for short feedback loops.** Compile time, test time, integration time, deploy time, time-to-detect-in-prod. Each one shorter is each one of these loops tighter.
- **Aim for changes that are small enough to be safe and large enough to be useful.** Increment size is a tuning parameter, not a fixed quantity.
- **Aim for designs that admit being wrong.** Reversibility is a property worth paying for.

---

Traces to: [`corpus/principles/modern-software-engineering.md`](../../corpus/principles/modern-software-engineering.md).
