# Hyrum's Law — rules

The non-negotiable and strongly-preferred rules from [`corpus/principles/hyrums-law.md`](../../corpus/principles/hyrums-law.md).
Loaded ambient; enforced by the [drift sensor](../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**The contract is what callers can observe, not what you intended.** The honest version of "this is an implementation detail" is "no caller has noticed this yet" — which is true until it isn't. Plan changes against the observable surface, not against the documented one. The documented one is a subset.

## GOLDEN PATH

- **Aim to shrink the observable surface.** Hide what you can; the smaller the surface, the smaller the contract that emerges. See [[information-hiding]] — Parnas's principle is Hyrum's defense.
- **Aim for versioning at every public interface.** Versioning gives you a way to evolve; the only alternative is freezing the behavior forever.
- **Aim to randomize or scramble what *should* be unobservable.** Randomize iteration order on hash maps; shuffle the order of equally-valid responses; inject latency jitter where determinism is not a contract. If you make a property non-deterministic from the start, no caller will depend on it.
- **Aim to learn from breakages, not avoid them.** When a change breaks a caller, the artifact is information: an emergent contract item you didn't know existed. Promote it or remove it deliberately.

---

Traces to: [`corpus/principles/hyrums-law.md`](../../corpus/principles/hyrums-law.md).
