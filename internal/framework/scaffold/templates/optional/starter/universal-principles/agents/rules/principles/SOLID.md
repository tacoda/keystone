# SOLID — rules

The non-negotiable and strongly-preferred rules from [`corpus/principles/SOLID.md`](../../corpus/principles/SOLID.md).
Loaded ambient; enforced by the [drift sensor](../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**A subtype that breaks its base type's contract is wrong, regardless of what the compiler says.** LSP is the only SOLID principle that can produce silently-broken code — fail it and the program lies about its types.

---

Traces to: [`corpus/principles/SOLID.md`](../../corpus/principles/SOLID.md).
