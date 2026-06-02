# Error Handling

How a system fails is part of what the system *is*. Articulated formally by Barbara Liskov and Alan Snyder in "Exception Handling in CLU" (1979), refined by Bjarne Stroustrup and Herb Sutter into the **exception-safety guarantees**, and pushed in a different direction by Joe Armstrong's *let it crash* discipline in Erlang. [[fail-fast]] says *when* to fail; this file says *how to propagate* failure correctly once you have decided to.

> **Rules extracted:** [`guides/principles/error-handling.md`](../../guides/principles/error-handling.md). This file holds the full reasoning, anti-patterns, and references.

## The first distinction: expected vs. exceptional

Two failure categories, treated very differently:

- **Expected failures** are normal control flow: the user typed an invalid email; the file is not there; the API returned 404; the optimistic lock conflicted. These are **values the caller must handle**. Return types (`Result<T, E>`, `Either`, `Option`, error codes, multi-return) make them visible. The compiler, the type checker, or the linter can refuse to compile code that ignores them.
- **Exceptional failures** are things the immediate caller is not equipped to handle: memory exhausted, invariant violated, configuration missing at startup, a bug that produced an unreachable state. These should propagate quickly to a place that *is* equipped — a request handler, a supervisor, the process boundary. Exceptions and panics are the right tools.

Confusing the two is the most common error-handling mistake. Using exceptions for ordinary control flow makes the code expensive to read and slow to run; treating "impossible" states as recoverable errors silently corrupts. See [[fail-fast]]'s IRON LAW.

## Exception safety — Stroustrup / Sutter

When an operation can fail mid-way, what state is the system in afterward? Three named guarantees, in increasing strength:

- **Basic guarantee.** If the operation throws, no resources leak and all invariants still hold — but the object may be in some valid but unspecified state. The minimum acceptable level for production code.
- **Strong guarantee.** If the operation throws, the system is exactly as it was before the operation began. Commit-or-rollback semantics. The standard for operations that modify important state.
- **No-throw guarantee.** The operation cannot fail. Required for destructors, for cleanup in `finally` blocks, for the rollback half of a strong-guarantee implementation. Building blocks for the other two.

Code that does not state which guarantee it offers offers **none** — and "none" means a partially-applied update on failure, which is the worst outcome.

## Propagation, not swallowing

The single most common production-incident smell is **the swallowed exception**: a `catch` block that logs and continues, leaving the surrounding code to operate on a half-built result. Two rules:

1. **If you cannot handle it here, do not catch it here.** Let it propagate to a layer that can.
2. **If you must catch it here to handle it, do not lose the cause.** Wrap, do not replace — preserve the original error as the cause of the new one, so the stack trace reads top-down from the layer that knew what to do back to the layer that knew what went wrong.

The Erlang corollary — **let it crash** — applies the same logic at the process level: do not write defensive code to recover from broken state; let the process die and let a supervisor restart a fresh one. This works *because* the process is the unit of state isolation; it fails outside that context. Where it applies (Erlang, OTP, Akka, supervised actors), it is the right answer.

## Error context

A failure without context is a failure that has to be reproduced before it can be fixed. The propagation chain should accumulate:

- **What was being attempted** ("uploading file user-123/avatar.png").
- **Which inputs were in play** (the IDs, the request parameters — *never* the secrets — see [[secrets-management]]).
- **Where in the chain it failed** (preserved stack trace; the layer that wrapped vs. the layer that originated).
- **What kind of failure** (transient vs. permanent; retryable vs. not).

Languages and libraries vary in how this is done — Go's `fmt.Errorf("...: %w", err)`, Rust's `anyhow::Context`, Java's chained causes, structured-log fields. The principle is the same: each layer adds context, none strips it.

## Retries belong with idempotency

When an expected failure is *transient* — network blip, rate limit, optimistic-lock conflict — the right response is often a retry. But a retry of a non-idempotent operation is not "the same operation again"; it is "the operation, plus the unknown question of whether the first attempt happened." Always pair the retry with the discipline in [[idempotency]].

## What it asks of you

- When you sketch a function, sketch its failure modes alongside its return type. The error half of the signature is half the contract. See [[design-by-contract]].
- When a failure path is not exercised by a test, the failure path is broken. Test the throw, the wrap, the swallow-detection. See [[tdd]].
- When you wrap an error, ask whether the wrapping adds *information* or just *another layer*. Wrapping that loses the cause is worse than not wrapping at all.
- When you implement a strong-guarantee operation, write the rollback path before the forward path. The forward path is easy; the rollback is what makes the guarantee hold.
- When you `panic`, `assert`, or throw an unrecoverable error, write the message as if for the next on-call engineer — not as if for the test framework. See [[observability]].

## Anti-patterns

- `catch (Exception e) { /* ignore */ }`.
- `catch (Exception e) { logger.warn(e); }` followed by code that uses the half-built result.
- A custom exception type that *replaces* the underlying cause instead of *wrapping* it (the trace ends at the wrap; the original is gone).
- Using exceptions for ordinary control flow — looping until the iterator throws, parsing by catching a parse exception, "checking" file existence by catching `FileNotFoundException`.
- A retry around a non-idempotent operation. See [[idempotency]].
- Treating `Result<T, E>` as "success or zero-info-string-error." The error type is part of the design; treat it as such.
- An error message that contains the secret value that failed to validate. See [[secrets-management]].
- Re-raising a generic exception type after losing the specific one, so callers can no longer dispatch on the failure category.

## References

- Liskov, B., & Snyder, A. (1979). "Exception Handling in CLU." *IEEE Transactions on Software Engineering*, 5(6). (The first formal treatment of exception handling as a language feature.)
- Stroustrup, B. (1991, 1997, 2013). *The C++ Programming Language* (editions). Addison-Wesley. (Defines and discusses the exception safety guarantees.)
- Sutter, H. (2000). *Exceptional C++*. Addison-Wesley. (Detailed treatment of exception safety in practice.)
- Bloch, J. (2017). *Effective Java*, 3rd ed. Addison-Wesley. (Items 69–77 — the canonical modern statement of exception discipline.)
- Armstrong, J. (2003). *Making reliable distributed systems in the presence of software errors*. PhD thesis, KTH. (The "let it crash" philosophy in context.)
- Goodliffe, P. (2006). *Code Craft: The Practice of Writing Excellent Code*. No Starch Press. (Chapter on error-handling discipline across languages.)
