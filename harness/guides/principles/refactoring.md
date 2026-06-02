# Refactoring — rules

The non-negotiable and strongly-preferred rules from [`corpus/principles/refactoring.md`](../../corpus/principles/refactoring.md).
Loaded ambient; enforced by the [drift sensor](../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**Behavior preservation is not optional.** A "refactoring" that changes behavior is not a refactoring; it is a bug-shaped change wearing the wrong hat. If you discover during refactoring that the existing behavior is wrong, stop, finish the refactoring, commit it, and then fix the behavior in a separate commit.

## GOLDEN RULES

- **Aim for steps small enough that the test suite stays green between each.** If two steps must land together to keep tests passing, the step was too large.
- **Aim for refactorings with names.** The catalog is large enough that nearly every move you want to make has a name. Use it.
- **Aim for separate commits for tidying and behavior.** Reviewers can read either kind quickly; the mixed kind is slow and error-prone.

---

Traces to: [`corpus/principles/refactoring.md`](../../corpus/principles/refactoring.md).
