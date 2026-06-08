# Simple Design — rules

The non-negotiable and strongly-preferred rules from [`corpus/principles/simple-design.md`](../../corpus/principles/simple-design.md).
Loaded ambient; enforced by the [drift sensor](../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**Simplicity is a property of the *next* reader, not the author.** The author already understands the code; their judgment that "this is simple" is uncalibrated. Simplicity is whether someone unfamiliar with the change can follow it.

## GOLDEN RULES

- **Aim for code that explains itself.** A comment that translates code into English is a sign the code was not written for the reader.
- **Aim for two commits, not one.** Tidying then behavior. Or behavior then tidying. Never both.
- **Aim for the smallest design that solves today's problem.** Tomorrow's problem will arrive with information you do not have yet.

---

Traces to: [`corpus/principles/simple-design.md`](../../corpus/principles/simple-design.md).
