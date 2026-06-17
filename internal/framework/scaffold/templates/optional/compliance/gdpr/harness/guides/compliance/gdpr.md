# GDPR (General Data Protection Regulation) — rules

The rules from [`corpus/gdpr.md`](../corpus/gdpr.md).
Loaded ambient; enforced by the [drift sensor](../../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**Personal data has a stated purpose, a stated retention period, and a documented deletion path — or it does not exist in the system.** A field with no documented basis is technical debt with regulatory exposure; the absence of documentation does not absolve the obligation to delete on request.

## GOLDEN PATH

- **Aim to collect less.** Every field of personal data is a future erasure-request line item, a future breach-disclosure line item, a future portability-export line item. The cheapest field is the one not collected.
- **Aim for deletion to be a one-action operation.** A "right to erasure" workflow that takes 22 manual steps across 6 systems will be skipped, batched, or done wrong. Automate it.
- **Aim for purpose tags on every personal-data field.** Schema-level annotations that link a field to its purpose, basis, and retention. Auditing becomes a query, not a project.
- **Aim for breach detection that wakes you up.** The 72-hour clock starts at *awareness*, not at *willingness to acknowledge*. Observability and on-call discipline are the precondition. See [[observability]].

---

Traces to: [`corpus/gdpr.md`](../corpus/gdpr.md).
