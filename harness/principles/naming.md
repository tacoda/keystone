# Naming

> There are only two hard things in Computer Science: cache invalidation and naming things. — Phil Karlton

A name is a promise. A reader who knows nothing about the implementation should be able to predict, from the name alone, what a function does, what an object represents, what a file contains, and roughly how it behaves. When the prediction is right, the code is easy to read; when the prediction is wrong, the reader is misled, and the rest of the file is now suspect. McConnell devotes an entire chapter (*Code Complete*, 2nd ed., ch. 11) to naming for a reason: it is the highest-leverage form of documentation in software, and the only kind that compilers preserve.

This file is short by design — naming has fewer principles than rules of thumb. The principles are stable; the rules of thumb belong in `../idioms/<stack>/` since the conventions are language-specific.

## What a name has to do

- **Predict behavior.** The name should let a reader guess what the named thing does or holds. `calculateTax` predicts; `process` does not. `dueDate` predicts; `data` does not.
- **Match the domain.** Use the words the business uses for the same concept — see [[bdd]] on ubiquitous language and [[separation-of-concerns]] on keeping the domain vocabulary distinct from infrastructure vocabulary.
- **Survive scope.** A name's required precision is proportional to the size of its scope. A one-line loop variable `i` is fine; a class field named `i` is malpractice. Short names for short-lived things; full names for everything that outlives a screen.
- **Be searchable.** A name you cannot grep for is a name you cannot maintain. `e` is harder to search than `event`; `id` is harder than `userId`.
- **Reveal level of abstraction.** A function should be named at the level of *what it does*, not *how it does it*. `sendEmail` is at the level of *what*; `openSmtpConnectionAndPushBytes` is at the level of *how*. Mixing levels in a name is the same smell as mixing them in a function — see [[SOLID]] (SRP).

## The renaming discipline

A wrong name is a bug whose fix is cheap until the name spreads. Modern IDEs make rename a one-keystroke operation; the resistance to rename is cultural, not technical.

- When you learn the right name, rename. Do it in its own commit — see [[refactoring]] (the two-hat rule).
- When the domain vocabulary changes, propagate the change through the code. Diverging vocabularies between team and code are a slow tax on everyone who has to translate.
- When you spot a name that says less than it should — `process`, `handle`, `manager`, `data`, `util`, `helper` — treat it as a smell to investigate, not necessarily to change. The vague name is sometimes covering up a vague concept; renaming without fixing the concept moves the problem rather than solving it.

## IRON LAW

**Names are part of the contract.** A function, class, or module that does more or less than its name implies has lied — and a reader who acted on the lie will be wrong. If the name no longer fits the behavior, change one of them. See [[least-astonishment]] (names as promises) and [[hyrums-law]] (callers depend on observable properties; the name is one of them).

## GOLDEN RULES

- **Aim for names that survive review without a glossary.** A reviewer who has never seen the code should be able to guess what each name means.
- **Aim for the most specific accurate name.** Not the shortest, not the longest — the most *informative*. `fetchActiveUsers` beats `fetch` and beats `getActiveUsersFromDatabaseAndCacheThem`.
- **Aim for names that match domain language exactly.** If the business says *invoice*, the code does not say *bill*, *receipt*, or *Statement*.
- **Aim to rename early.** A bad name caught in the first week is one rename; caught after a year, it is a multi-file change with a long review.

## Anti-patterns

- `Manager`, `Helper`, `Util`, `Common`, `Base`, `Abstract`, `Generic` as the most specific word in a name. Each is a confession that no specific name was found — see [[coupling-cohesion]].
- Hungarian notation in a typed language. The type system already encodes the type; the prefix is duplicate and rots when the type changes.
- `data`, `info`, `value`, `result`, `obj` as a field or parameter name. None of these tell the reader anything they did not already know.
- Boolean fields named without polarity: `flag`, `enabled` (enabled what?), `status` (true means good or bad?). Prefer `isPublished`, `hasPaid`, `requiresMfa`.
- Two near-synonyms used for different concepts: `user` and `customer` and `account` for the same entity, or for *different* entities with no documented distinction.
- Abbreviations that save three characters and cost the reader twenty seconds: `usrCtrlMgr`, `acctRecv`, `pmtPrcsr`.
- Negative names whose negation is hard to read: `isNotEmpty`, `disallowed`, then double-negated at the call site: `if (!isNotEmpty(...))`.

## References

- McConnell, S. (2004). *Code Complete*, 2nd ed. Microsoft Press. (Chapter 11 — the most thorough treatment of naming as engineering.)
- Martin, R. C. (2008). *Clean Code: A Handbook of Agile Software Craftsmanship*. Prentice Hall. (Chapter 2 — meaningful names.)
- Ottinger, T., & Langr, J. (n.d.). *Ottinger's Rules for Variable and Class Naming*. (Practical short guide.)
- Evans, E. (2003). *Domain-Driven Design: Tackling Complexity in the Heart of Software*. Addison-Wesley. (Ubiquitous language — names as the seam between the domain and the code.)
- Karlton, P. — quoted by Tim Bray and others; the original source is conversational and unrecorded, but the attribution is universal.
