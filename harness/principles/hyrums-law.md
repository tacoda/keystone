# Hyrum's Law

> With a sufficient number of users of an API, it does not matter what you promise in the contract: all observable behaviors of your system will be depended on by somebody. — Hyrum Wright

Coined at Google by Hyrum Wright, codifying a pattern that anyone who has tried to change a popular API has felt. The contract you wrote is what *you* promised; the contract that exists is everything any caller can observe and has, somewhere, started depending on. The two are not the same.

## The principle

Every observable property of a system is, at scale, a contract. Not the documented contract — *all* of it. Examples of "observable" that have broken downstream callers in shipped systems:

- The exact wording of an error message.
- The order of keys in a JSON object.
- The microsecond-level latency of a fast path.
- The set of bits in a stack trace.
- The timing of a callback relative to a side effect.
- A bug fix that changed an output by one digit in the last decimal place.
- Removing an undocumented HTTP header.
- The fact that two unrelated requests happened to return in the order they were sent.

If a caller could see it, a caller might depend on it. With enough callers, *some* caller does. The defense is not "we documented otherwise" — Hyrum's Law specifies that documentation does not matter.

This is the operational dual of [[design-by-contract]]: Meyer's contract is the **intended** specification; Hyrum's law describes the **emergent** specification. The gap between the two is where compatibility breaks live.

## What it asks of you

- When you ship a public interface, treat **every observable property** as part of the contract until you have evidence otherwise. The set of things that "shouldn't matter" is smaller than it looks.
- When you must change an observable behavior, the right tool is usually not "we'll announce it"; it is **versioning** — leave the old behavior in place, expose the new one alongside, migrate callers, retire the old. See [[separation-of-concerns]] (version boundary as a concern).
- When you discover an unintended dependency on observable behavior, ask whether to *fix the dependency* or *promote the behavior to documented contract*. Both are valid; "ignore it" is not.
- When you write tests for code others will call, write tests that pin the **documented** behavior, not every observable detail. Tests that lock in incidental properties make the same mistake Hyrum's Law warns about — internally.
- When you build a *private* interface — used only by code you control — Hyrum's Law applies in proportion to your callers. A function called from three places can be changed freely; one called from three hundred places cannot. See [[refactoring]] on the cost of large-radius changes.

## The xkcd corollary

The famous illustration (xkcd 1172, "Workflow"): some user, somewhere, depends on a behavior you would never have predicted — in the cartoon, a user who relies on the spacebar making the computer overheat to warm their lunch. The joke is funny because it is not exaggerated. *Whatever* is observable will be depended on by *somebody* whose use case you cannot anticipate.

## IRON LAW

**The contract is what callers can observe, not what you intended.** The honest version of "this is an implementation detail" is "no caller has noticed this yet" — which is true until it isn't. Plan changes against the observable surface, not against the documented one. The documented one is a subset.

## GOLDEN RULES

- **Aim to shrink the observable surface.** Hide what you can; the smaller the surface, the smaller the contract that emerges. See [[information-hiding]] — Parnas's principle is Hyrum's defense.
- **Aim for versioning at every public interface.** Versioning gives you a way to evolve; the only alternative is freezing the behavior forever.
- **Aim to randomize or scramble what *should* be unobservable.** Randomize iteration order on hash maps; shuffle the order of equally-valid responses; inject latency jitter where determinism is not a contract. If you make a property non-deterministic from the start, no caller will depend on it.
- **Aim to learn from breakages, not avoid them.** When a change breaks a caller, the artifact is information: an emergent contract item you didn't know existed. Promote it or remove it deliberately.

## Anti-patterns

- "It's a private API, callers shouldn't have depended on it" — said about a private API with two hundred callers, after the change broke production.
- "We documented the breaking change" — accurate, and irrelevant; documentation is not a defense against Hyrum's Law.
- A JSON response whose keys are emitted in dictionary order, where downstream parsers happened to depend on that order, that "shouldn't" matter.
- A stable-sort assumption made into an unstable-sort change for performance, breaking a test suite that relied on the previous order.
- A "no-op refactoring" that altered a side-effect's relative ordering, breaking a caller that observed the order.
- Bumping the version of a transitive dependency whose behavior changed in a way the maintainer considered a bugfix and your code considered the contract.

## References

- Wright, H. (n.d.). *Hyrum's Law*. hyrumslaw.com. (Canonical statement.)
- Winters, T., Manshreck, T., & Wright, H. (2020). *Software Engineering at Google: Lessons Learned from Programming Over Time*. O'Reilly. (Chapter 1 introduces Hyrum's Law in print; the entire book is a meditation on its consequences at scale.)
- Munroe, R. (2013). *xkcd 1172: "Workflow"*. xkcd.com/1172. (The canonical illustration.)
- Parnas, D. L. (1972). "On the Criteria To Be Used in Decomposing Systems into Modules." *Communications of the ACM*, 15(12). (The pre-emptive defense — see [[information-hiding]].)
