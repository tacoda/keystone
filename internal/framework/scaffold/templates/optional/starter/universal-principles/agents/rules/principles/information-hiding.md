# Information Hiding — rules

The non-negotiable and strongly-preferred rules from [`references/principles/information-hiding.md`](../../references/principles/information-hiding.md).
Loaded ambient; enforced by the [drift hook](../../hooks/drift.md). The reference file holds the full reasoning, anti-patterns, and references.

## GOLDEN RULE

- **Aim for deep modules.** A module with a small, stable interface and a large, valuable implementation hides more. Shallow modules — large interface, thin implementation — leak complexity into their callers.
- **Aim for interfaces that survive their implementations.** The implementation may be rewritten; the interface should not need to be.

---

Traces to: [`references/principles/information-hiding.md`](../../references/principles/information-hiding.md).
