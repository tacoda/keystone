# Design by Contract

Every routine, every class, every module has obligations and guarantees that should be stated as plainly as its name. Articulated by Bertrand Meyer in *Object-Oriented Software Construction* (1988; 2nd ed., 1997) and embodied in the Eiffel language. Even in languages without first-class contract support, the discipline shapes how you write code.

## The contract

For each routine:

- **Precondition** — what the caller must guarantee before calling. If the caller violates it, the routine is *not obliged* to behave correctly; the bug is the caller's.
- **Postcondition** — what the routine guarantees on return, assuming the precondition held. If the routine violates it, the bug is the routine's.
- **Invariant** — what is true of the object *between* calls (and before/after any public method). Constructors must establish it; public methods must preserve it.

The three together form a contract: *given X, I will deliver Y, and I will always be in state Z when no one is calling me.* Without an explicit contract, every caller guesses, defensively re-validates, and the routine quietly grows tolerance for inputs that should be impossible.

## Where contracts live

Meyer's strong form attaches contracts to the code (assertions checked at runtime, sometimes only in debug builds). The weaker but widely-applicable form: contracts live in **types**, **assertions**, **tests**, and **documented preconditions** at the boundary.

Modern equivalents:

- A non-nullable type *is* a precondition that the argument is non-null.
- A constructor that validates and rejects invalid arguments *enforces* an invariant.
- A unit test that probes the boundary *exercises* the precondition.
- A `panic` / `assert` deep in a routine when a "can't happen" state appears *catches* an invariant violation.

## What it asks of you

- When you write a public method, write down what it expects and what it promises. The exercise is the value, even if the contract never reaches a comment. See [[information-hiding]] — the contract *is* the interface.
- When you find yourself re-validating an input you were just handed by trusted internal code, ask whether the validation belongs *upstream* at the trust boundary. Validation everywhere is validation nowhere. See [[separation-of-concerns]].
- When an invariant could be violated, **fail loudly** in the place that detects it — see [[fail-fast]]. A silently corrupted object will be detected far from the cause.
- When a subclass tightens a precondition or weakens a postcondition, LSP is broken — see [[SOLID]].

## IRON LAW

**Trust your inputs at the boundary; trust your invariants inside.** Validate at the seam between trusted and untrusted code (user input, network, external API). Inside the trust boundary, assume the contract holds, and assert when it does not. Validating everywhere both hides the real boundary and makes every method harder to read.

## GOLDEN RULES

- **Aim for contracts that the type system can enforce.** Types are checked contracts; comments are wishes.
- **Aim to make impossible states unrepresentable.** A user with `age = -3` should not be constructible.
- **Aim for invariants that constructors establish and methods preserve.** If a public method can leave the object in an invalid state, the contract is wrong.

## Anti-patterns

- A constructor that accepts any input and produces a half-built object the user must finish initializing.
- A method that returns `null` (or an empty value) when it could not produce a result, with no documentation of when this happens. See [[law-of-demeter]] anti-patterns on traversing nullable chains.
- Re-validating the same precondition at every layer of a call stack.
- A subclass that tightens what arguments it accepts compared to its parent.
- An object whose validity depends on calling methods in a specific order, with no enforcement.

## References

- Meyer, B. (1988, 1997). *Object-Oriented Software Construction* (1st and 2nd eds.). Prentice Hall. (The canonical exposition of design by contract.)
- Meyer, B. (1992). "Applying Design by Contract." *IEEE Computer*, 25(10), 40–51.
- Liskov, B., & Wing, J. M. (1994). "A Behavioral Notion of Subtyping." *ACM TOPLAS*, 16(6). (Formal underpinning of contracts in the presence of subtyping; complements LSP.)
- Hoare, C. A. R. (1969). "An Axiomatic Basis for Computer Programming." *Communications of the ACM*, 12(10). (Precursor — preconditions and postconditions as logical assertions.)
