# Error Handling — rules

The non-negotiable and strongly-preferred rules from [`corpus/principles/error-handling.md`](../../corpus/principles/error-handling.md).
Loaded ambient; enforced by the [drift sensor](../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**Never silently continue past an error.** The two acceptable responses to detecting an error are *handle it* (the operation completes with a defined outcome) or *propagate it* (the operation does not complete, and the caller is told). Logging-and-continuing is neither; it is a third option that pretends to handle the error while leaving the system in a state no one designed. See [[fail-fast]].

## GOLDEN RULES

- **Aim for failure modes that are part of the type signature.** Errors as values, where the language supports it. The compiler will then refuse to let callers ignore them.
- **Aim for the strong guarantee on operations that mutate important state.** Either the change happens entirely, or it does not happen at all.
- **Aim for one swallow per layer, deliberately.** The boundary that converts an exception to a user-visible error is the right place to "stop" it. Every other layer propagates.
- **Aim for errors that tell the next operator what to do.** "Connection refused" is information; "an error occurred" is noise.

---

Traces to: [`corpus/principles/error-handling.md`](../../corpus/principles/error-handling.md).
