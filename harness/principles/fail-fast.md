# Fail Fast

Detect failures as close to their source as possible, and stop. Articulated by Jim Shore in "Fail Fast" (*IEEE Software*, 2004), formalizing a discipline that had circulated in the Smalltalk and Lisp communities for decades. The opposite of fail-fast is **fail-slow**: a corrupted state propagates through layers of "tolerant" code until it surfaces somewhere far from the cause.

## The principle

A bug that is detected the instant it occurs is cheap: the stack trace points at the cause, the state is still intact, the fix is local. A bug that propagates through five layers of defensive code is expensive: the stack trace points at a symptom, the state is partially corrupted, the cause is anyone's guess.

Fail-fast is therefore a debugging principle as much as a runtime one — it pays for itself in *time to diagnose*, not in runtime safety. The savings show up the first time something goes wrong.

Shore's distinction: fail-fast is *not* the same as "fail-safe" (the security default — see [[security]]'s IRON LAW). Fail-safe asks: when a control fails, what is the safe outcome? Fail-fast asks: when something *unexpected* happens, how soon do we notice?

## What it asks of you

- When you detect an "impossible" state, **assert and stop**, do not paper over it. A `panic("unreachable")` or `assert(invariant)` is information; a silent skip is a bug waiting for production.
- When you discover that data violates your assumptions, refuse it at the **earliest** point you can recognize the violation — see [[design-by-contract]]. The deeper the validation lives, the worse the diagnostics.
- When you write a `try { ... } catch (Exception) { /* ignore */ }`, ask what failure that hides. Swallowed exceptions are the canonical fail-slow pattern.
- When startup configuration is wrong, crash at startup, not at the first request. A web app that boots with a missing secret should not serve any request.

## IRON LAW

**Never continue past a violated invariant.** If your code detects that something it required to be true is false, it has two correct responses: refuse the operation, or terminate the process. Continuing is not one of them, regardless of how disruptive the alternative seems. The disruption is real; the silent corruption is worse.

## GOLDEN RULES

- **Aim to crash at startup, not at runtime.** Configuration errors, missing dependencies, unreachable services discovered at boot should prevent the process from accepting traffic. Failing under load is failing at the worst possible time.
- **Aim for diagnostics at the source.** A failure message should answer *what failed, where, and what was expected* — not just "operation failed."
- **Aim to distinguish expected from unexpected.** Expected errors (user input invalid, network blip) have handlers. Unexpected errors (invariant violated, "impossible" branch taken) have alarms.

## Distinguishing fail-fast from defensive paranoia

Fail-fast is not "check everything everywhere." That is its opposite — see [[design-by-contract]] on the trust boundary. The rule is:

- **At trust boundaries:** validate explicitly, fail with a clear message.
- **Inside trust boundaries:** assert invariants, fail loudly if violated.
- **Throughout:** never silently continue past an unexpected condition.

The discipline is *one* check at the right place, not many checks everywhere.

## Anti-patterns

- `catch (Exception e) { log.warn(...); }` followed by code that uses the half-built result.
- A null-tolerant chain (`a?.b?.c?.d`) used to suppress a NullPointerException whose cause was a missing migration.
- "Lenient parsers" that accept malformed input and best-effort interpret it. (Compare [[postels-law]] — even Postel's liberal acceptance has limits, and silently accepting clearly-malformed input is past them.)
- Health checks that always return 200 because "the service is technically running."
- A startup path that logs "warning: config missing" and serves anyway.

## References

- Shore, J. (2004). "Fail Fast." *IEEE Software*, 21(5), 21–25.
- Goodliffe, P. (2006). *Code Craft: The Practice of Writing Excellent Code*. No Starch Press. (Chapter on error handling discipline.)
- Hunt, A., & Thomas, D. (1999). *The Pragmatic Programmer*. Addison-Wesley. ("Crash early" — same idea, different phrasing.)
- Meyer, B. (1997). *Object-Oriented Software Construction*, 2nd ed. Prentice Hall. (Assertion-based design — fail-fast as an outgrowth of [[design-by-contract]].)
