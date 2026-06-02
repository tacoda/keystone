# Single-Page Application (SPA) — rules

The rules from [`corpus/principles/spa.md`](../../corpus/principles/spa.md).
Loaded ambient; enforced by the [drift sensor](../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**The server is the source of truth. The client is a cache and a UI.** Any authorization, validation, or business rule enforced only in the SPA is unenforced. A user can pause your JavaScript and edit the request; an attacker can replay it. The SPA is presentation; the API is the gate.

## GOLDEN RULES

- **Aim to make the back button work.** Browser history, scroll position, form state on return. Users do not forgive a back button that surprises them. See [[least-astonishment]].
- **Aim to measure what users experience.** Time to first byte, time to interactive, largest contentful paint, cumulative layout shift. Real User Monitoring on a representative slice — not just the developer's M1 on fiber. See [[observability]].
- **Aim to render *something* fast.** Skeleton screens, optimistic UI, streamed responses. Blank pages followed by sudden full pages are the worst-of-both-worlds.
- **Aim to keep the bundle small.** Code-split routes; lazy-load heavy components; audit dependencies. The bundle is the product's perceived performance.
- **Aim to keep critical paths free of business logic in the client.** Validation in the UI is for *user feedback*; validation in the API is for *correctness*. Do both — never replace the second with the first.

---

Traces to: [`corpus/principles/spa.md`](../../corpus/principles/spa.md).
