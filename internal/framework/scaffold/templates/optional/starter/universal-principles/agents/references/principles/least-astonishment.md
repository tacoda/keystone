# Principle of Least Astonishment

A component of a system should behave in the way most users will expect it to behave. Where expectation and behavior diverge, the component is *astonishing*, and astonishment is paid for in bugs, support tickets, and incidents. Articulated in the PL/I design community in the 1970s; widely cited in HCI and API design literature since.

The principle applies at every scale: a function name, a configuration default, an HTTP status code, a keyboard shortcut, a button position. *"Will the next reader/user predict what this does?"* is one question that links them all.

> **Rules extracted:** [`guides/principles/least-astonishment.md`](../../guides/principles/least-astonishment.md). This file holds the full reasoning, anti-patterns, and references.

## Whose expectations?

The principle is meaningless without a population. *Least astonishing to whom?*

- **Code:** the next engineer reading the diff, who knows the language and the codebase's conventions — see [[simple-design]] (reveals intention).
- **APIs:** developers fluent in the surrounding ecosystem (HTTP semantics for REST APIs, SQL semantics for query DSLs, etc.).
- **UIs:** users fluent in the platform's conventions (macOS users expect ⌘-W to close; Windows users expect Alt-F4).

The duty of design is to identify the right population and respect *their* expectations — not to invent novel behavior that "feels right" to the author.

## What it asks of you

- When you name something, name it for what it does — not what it currently *also* does as a side effect. A function called `getUser` that also writes to a cache and emits a metric is astonishing. See [[SOLID]] (SRP), [[coupling-cohesion]].
- When you choose a default, choose the value that the typical user would have chosen given full information. Defaults that bite the typical user are astonishing twice — once on misuse, once on the discovery that the default existed.
- When you reuse an existing concept (an HTTP verb, a SQL operator, a language idiom), respect its meaning. An endpoint that responds `200 OK` with `{"error": ...}` is astonishing.
- When you depart from a convention, you owe the reader an explanation — a comment, an ADR, a section in the README. Departures without justification are the most expensive kind of astonishment.

## Anti-patterns

- A "getter" that mutates state.
- A function whose return value is "the result you asked for, *or* a placeholder when the operation failed" — the caller does not know which without checking.
- An HTTP `DELETE` that is non-idempotent, or a `GET` that mutates.
- A configuration flag whose default is the *less safe* option.
- A keyboard shortcut that rebinds a near-universal convention to a custom action.
- An error message that says "an error occurred" — astonishing because it carries no information.

## References

- Cowlishaw, M. (1984). *The Design of the REXX Language*. IBM Systems Journal. (Early articulation of "least astonishment" as a language design principle.)
- Raymond, E. S. (2003). *The Art of Unix Programming*. Addison-Wesley. (Chapter on the Rule of Least Surprise.)
- Nielsen, J. (1994). "10 Usability Heuristics for User Interface Design." (Heuristic #4 — consistency and standards — is the UI form of this principle.)
- Bloch, J. (2006). "How to Design a Good API and Why It Matters." (Talk; emphasizes that API surprise is among the most expensive forms.)
