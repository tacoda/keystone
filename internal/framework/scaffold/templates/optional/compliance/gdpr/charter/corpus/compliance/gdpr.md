# GDPR (General Data Protection Regulation)

EU regulation 2016/679, in force since May 2018. Governs the processing of personal data of people in the EU and EEA — regardless of where the processor is located. The single most cited modern data-protection regime; the model for GDPR-like regulations in California (CCPA), Brazil (LGPD), the UK, and elsewhere.

This file states the engineering-relevant principles. It is **not legal advice**; the regulation has scope, exceptions, and remedies that require legal counsel to apply correctly to a specific business. The compliance program is the lawyers' deliverable; what the engineering team owns is *making the system compliant by design*.

> **Rules extracted:** [`guides/gdpr.md`](../guides/gdpr.md). This file holds the full reasoning, anti-patterns, and references.

## The engineering-relevant principles

- **Lawful basis for processing.** Every collection and use of personal data must rest on one of six lawful bases (consent, contract, legal obligation, vital interests, public task, legitimate interests). At the system level, this means: *can you, for any field of personal data, point to the basis?*
- **Data minimization.** Collect only what is necessary for the stated purpose; retain only as long as necessary. A schema that collects "just in case" is a non-compliance generator. See [[information-hiding]] — the data you do not hold cannot leak.
- **Purpose limitation.** Data collected for one stated purpose may not be silently used for another. "We added analytics on the signup flow" is a purpose change.
- **Right to access.** A subject can request a copy of all their data. The system must produce a structured, machine-readable export.
- **Right to erasure ("right to be forgotten").** A subject can request deletion. The system must actually delete — including from backups, caches, analytics warehouses, search indices, and exports. Soft-delete that leaves data forever in some downstream is *not erasure*.
- **Right to rectification.** A subject can correct inaccurate data.
- **Right to data portability.** A subject can request data in a portable format.
- **Privacy by design and by default.** The system's defaults must protect privacy without user action.
- **Breach notification.** A personal-data breach must be reported to the supervisory authority within **72 hours** of becoming aware. The engineering implication: incident-detection has a hard deadline.

## What it asks of you

- When you add a field of personal data to a schema, the design questions are: *what is the lawful basis, how long do we keep it, who is allowed to see it, and how do we delete it?* If any of those four cannot be answered, the field is not ready.
- When you build a feature, ask whether it processes personal data. If yes, the user-facing privacy disclosures and the data-processing record must be updated *before* the feature ships. Privacy notices are not catch-up work.
- When you build a delete path, **delete from everywhere** — primary stores, replicas, caches, search indices, ETL warehouses, analytics, logs, archives. Erasure is comprehensive or it does not exist.
- When you ship a third-party integration, you are sending personal data to a processor. Data-processing agreements, sub-processor lists, and (for cross-border transfers) Standard Contractual Clauses or equivalent become part of the integration's design. See [[security-threats]] (supply chain).
- When you log, scrub personal data unless there is a legitimate basis for logging it. Logs replicate everywhere; one log line per request becomes thousands of copies of the contained PII. See [[secrets-management]] (rotate, don't redact — the principle generalizes).

## Anti-patterns

- A "GDPR delete" that flips a flag and leaves the data intact. The flag is a wish; the data is the obligation.
- A data warehouse that retains personal data forever because nobody owns deletion there. Out of sight, in scope.
- A "consent" UI that pre-checks the boxes. GDPR consent is opt-in, informed, and revocable; pre-checked is none of those.
- A logging pipeline that copies request bodies, including PII, to a SaaS provider in another jurisdiction without a transfer mechanism in place.
- A "we'll get to it before audit" attitude. Audits arrive on a calendar; breaches arrive without warning.
- Backups that retain deleted personal data forever, with no recovery process that re-applies pending erasures. The deletion was undone the moment the backup ran.

## References

- *Regulation (EU) 2016/679* — the General Data Protection Regulation. eur-lex.europa.eu/eli/reg/2016/679/oj. (The text.)
- European Data Protection Board. *Guidelines and Recommendations*. edpb.europa.eu. (Operational interpretation.)
- ICO (UK). *Guide to the UK GDPR*. ico.org.uk. (Post-Brexit but substantially aligned; clearer prose than the EU original.)
- Article 29 Working Party (now EDPB). *Working Documents on Data Protection Impact Assessments*. (DPIA guidance — the regulatory equivalent of a threat model.)
- ENISA (European Union Agency for Cybersecurity). *Privacy and Data Protection in Mobile Applications*. (Engineering-focused complement to the regulation.)
