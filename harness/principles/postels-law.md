# Postel's Law (The Robustness Principle)

> Be conservative in what you do, be liberal in what you accept from others.

Stated by Jon Postel in RFC 760 (1980) and codified in RFC 1122 (1989) for Internet protocol design. The most quoted heuristic for **interoperability** in networked systems. Also the most argued-about — its modern critique is part of the principle, and applying it well means knowing where each half stops.

## The two halves

**Conservative in what you do.** What you produce — requests, responses, payloads — should conform strictly to the specification. No exotic encodings, no extension fields others may not understand, no relying on implementations that "happen to work." If the spec gives you ten ways to express the same thing, pick the most universally supported.

**Liberal in what you accept.** What you receive should be parsed forgivingly within the limits of unambiguous interpretation: tolerate extra whitespace, unknown-but-ignorable fields, optional elements omitted, casing variations explicitly permitted by the spec. The goal is interop with conforming peers that happen to disagree on edge cases.

## The modern critique

In "The Robustness Principle Reconsidered" (Eric Allman, *ACM Queue*, 2011) and the IETF draft "The Harmful Consequences of the Robustness Principle" (Martin Thomson, 2019), the second half has been challenged. The argument:

- Liberal acceptance lets non-conformant senders survive in the wild. Those senders become dependencies on the leniency.
- Receivers that "fix up" malformed input mask bugs in senders, so the bugs never get fixed.
- Over time, the protocol drifts from the spec toward "whatever the most permissive receiver accepts," which is unknowable and untestable.
- Security-sensitive parsing must be strict — liberal acceptance is a frequent source of vulnerabilities (smuggling, request splitting, ambiguous encodings).

So the modern reading is: **strict in what you send; strict by default in what you accept**; *liberal only where the spec explicitly permits variation and the variation is harmless.* The principle survives, with the second half on a tighter leash than Postel originally suggested.

## What it asks of you

- When you emit a payload — an API response, an event, a message — produce the canonical form. Other systems will depend on what you send; do not give them ambiguity to depend on.
- When you accept a payload, distinguish three categories: **clearly valid** (process), **clearly invalid** (reject with a useful error), **ambiguous** (reject — do not guess). The third category is where the original Postel formulation gets systems in trouble.
- When you must tolerate variation for backward compatibility, document it explicitly and put a sunset on it. Permanent leniency is permanent legacy. See [[fail-fast]].
- When parsing for a security-relevant decision (authentication tokens, signed payloads, authorization scopes), be **strict** — no leniency, no second guesses. See [[security]].

## IRON LAW

**Ambiguous input is invalid input.** If two reasonable readings of a payload exist, you do not get to pick one. Reject the input and surface the ambiguity. Silent interpretation of ambiguous input is how request smuggling, JSON parser confusion, and header-injection vulnerabilities get into production.

## GOLDEN RULES

- **Aim to send the strictest, most canonical form your spec allows.** It is the most interoperable.
- **Aim to accept only what is unambiguously valid.** Reject everything else with a useful, structured error.
- **Aim to log unexpected-but-tolerated variation.** If you ever decide to tighten the parser, the log tells you who will break.

## Anti-patterns

- An HTML parser that "fixes up" malformed input differently from every other parser, creating cross-parser desynchronization.
- An API that quietly accepts both `user_id` and `userId` because "users get confused" — now both are forever in the contract.
- A signature verifier that accepts multiple encodings of the same payload (canonicalization bugs are a security-vulnerability factory).
- A log line that says "received malformed request, attempting recovery."
- Liberal acceptance with no measurement — you do not know who is sending non-conformant data, so you cannot ever tighten the rule.

## References

- Postel, J. (1980). RFC 760: *DOD Standard Internet Protocol*. (Original statement of the principle.)
- Braden, R. (Ed.) (1989). RFC 1122: *Requirements for Internet Hosts — Communication Layers*. (Codification with the now-canonical wording.)
- Allman, E. (2011). "The Robustness Principle Reconsidered." *ACM Queue*, 9(6).
- Thomson, M. (2019). "The Harmful Consequences of the Robustness Principle." IETF draft-iab-protocol-maintenance.
- Sassaman, L., Patterson, M. L., Bratus, S., & Locasto, M. E. (2013). "Security Applications of Formal Language Theory." *IEEE Systems Journal*. (The "LangSec" critique — ambiguous parsing as a vulnerability class.)
