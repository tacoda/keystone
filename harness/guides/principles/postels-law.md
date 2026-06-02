# Postel's Law (The Robustness Principle) — rules

The non-negotiable and strongly-preferred rules from [`corpus/principles/postels-law.md`](../../corpus/principles/postels-law.md).
Loaded ambient; enforced by the [drift sensor](../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**Ambiguous input is invalid input.** If two reasonable readings of a payload exist, you do not get to pick one. Reject the input and surface the ambiguity. Silent interpretation of ambiguous input is how request smuggling, JSON parser confusion, and header-injection vulnerabilities get into production.

## GOLDEN RULES

- **Aim to send the strictest, most canonical form your spec allows.** It is the most interoperable.
- **Aim to accept only what is unambiguously valid.** Reject everything else with a useful, structured error.
- **Aim to log unexpected-but-tolerated variation.** If you ever decide to tighten the parser, the log tells you who will break.

---

Traces to: [`corpus/principles/postels-law.md`](../../corpus/principles/postels-law.md).
