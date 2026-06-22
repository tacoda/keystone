# Principle of Least Astonishment — rules

The non-negotiable and strongly-preferred rules from [`corpus/principles/least-astonishment.md`](../../corpus/principles/least-astonishment.md).
Loaded ambient; enforced by the [drift sensor](../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**Names are promises.** A function, class, file, or endpoint that does more than its name implies has lied — and the lie will be discovered by someone who acted on the promise. If you cannot honor the name, change the name. See [[refactoring]] (rename is the cheapest refactoring; do it early).

## GOLDEN RULE

- **Aim to be boring in the choices that do not differentiate you.** Use the conventional name, the conventional default, the conventional status code. Save novelty for the parts of the system that earn it.
- **Aim to satisfy the reader's first guess.** When two designs are both correct, prefer the one the reader was already going to expect.
- **Aim to document deliberate astonishment.** If a behavior must be unusual, surface the unusualness — in the name, the docs, the type signature, or the error message.

---

Traces to: [`corpus/principles/least-astonishment.md`](../../corpus/principles/least-astonishment.md).
