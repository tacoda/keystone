# HIPAA (Health Insurance Portability and Accountability Act)

US federal law, 1996, with the Security Rule (2003), Privacy Rule (2003), and HITECH amendments (2009) defining current obligations for handling **Protected Health Information (PHI)** in the United States. Applies to *covered entities* (healthcare providers, plans, clearinghouses) and *business associates* (anyone processing PHI on their behalf).

This file states the engineering-relevant principles. It is **not legal advice**; HIPAA has specific scoping, business-associate-agreement requirements, breach reporting, and state-law interactions that require counsel. The compliance program is the lawyers'; the engineering job is to make PHI-handling systems *defensible by design*.

## What counts as PHI

Any individually identifiable health information held or transmitted by a covered entity or business associate. The HHS lists **18 identifiers** whose presence (with health information) makes data PHI: names, addresses smaller than state, dates more specific than year, phone numbers, fax numbers, emails, SSNs, MRNs, health plan IDs, account numbers, certificate/license numbers, vehicle IDs, device IDs, URLs, IPs, biometric IDs, photos, and any other unique identifier or code.

The corollary: **de-identified** data (per the Safe Harbor or Expert Determination methods) is not PHI and is out of scope. The decision to de-identify, and to do it correctly, has architectural consequences — see [[separation-of-concerns]].

## The Security Rule's three categories of safeguards

Engineering-relevant; each has multiple required and "addressable" implementations.

- **Administrative safeguards** — risk analysis, sanction policy, contingency plan, access management, training. The policy half.
- **Physical safeguards** — facility access, workstation security, device and media controls. The hardware half.
- **Technical safeguards** — access control, audit controls, integrity controls, transmission security. The *code-and-config* half — what the engineering team owns.

The technical safeguards translate roughly to: **authenticate every actor, authorize every access, log every access, encrypt PHI in transit and at rest, prove that PHI was not altered, and detect when controls fail.** See [[security]] and [[security-threats]] — HIPAA is mostly a strong form of the same principles, with specific evidence requirements.

## What it asks of you

- When you store PHI, encrypt it at rest with keys not held by the same identity that accesses the data. See [[secrets-management]]. Cloud KMS with workload-identity-scoped grants is the modern baseline.
- When you transmit PHI, encrypt in transit with TLS 1.2+ (1.3 preferred), authenticated endpoints, no fallback. "Internal" networks are still networks. See [[postels-law]] — be strict about what you accept.
- When you authorize access, **log it.** Every read of PHI, every write, every export. The log itself is PHI-adjacent and has its own access controls. Audit logs must be tamper-evident; the same identity that can edit them is the threat. See [[observability]].
- When you build a feature on top of PHI, ask whether the feature *needs* PHI or can work on a de-identified projection. Most analytics and most internal dashboards can. The non-PHI path is the cheaper compliance path.
- When you integrate with a third party that touches PHI, you need a **Business Associate Agreement (BAA)** with them — and a confirmation that they pass the BAA chain to their sub-processors. See [[security-threats]] (supply chain).
- When a breach occurs, breach notification is required within **60 days** for affected individuals (and HHS for incidents of 500+). The engineering implication: forensic readiness — logs, IDS, and a defined incident-response runbook.

## IRON LAW

**Every access to PHI is authenticated, authorized, and logged.** No exceptions for service accounts, batch jobs, debugging sessions, or "trusted" internal users. An access path that bypasses the audit log is a HIPAA finding waiting to be discovered.

## GOLDEN RULES

- **Aim for the smallest PHI footprint.** De-identify where possible; project to non-PHI fields for downstream uses; redact in logs; expire data on a schedule. Each PHI field is exposure.
- **Aim for minimum-necessary access.** A clinician needs their patients; a researcher needs the dataset; a developer needs neither — *especially in production*. See [[security]] (least privilege).
- **Aim for tamper-evident audit logs.** Write-once, signed, exported to a destination outside the access scope of the systems being audited. The audit log that auditors look at is the one that wasn't editable.
- **Aim for break-glass access with alerting.** Emergency access exists; it must be exceptional, time-boxed, logged, and reviewed.

## Anti-patterns

- "Developer access to prod for debugging" — a standing path into PHI with no audit. The single most common HIPAA finding.
- PHI in application logs because "we'll redact it later." Logs propagate immediately; redaction never catches up. See [[secrets-management]].
- A test environment populated with a copy of production PHI. The test environment is now in scope, with weaker controls than production.
- Email or chat as a channel for PHI ("just sending you this patient record"). Outbound channels lack the controls; the moment the data leaves, the BAA boundary is crossed.
- An audit log stored in the same database as the data being audited, editable by the same role.
- A third-party SaaS adopted to "speed up the team" with no BAA and no review of what data flows to it. The SaaS is now a regulatory exposure surface no one mapped.

## References

- 45 CFR §§ 160, 162, 164 — *HIPAA Administrative Simplification Regulations*. ecfr.gov. (The text.)
- HHS Office for Civil Rights. *HIPAA Security Rule Crosswalk to NIST Cybersecurity Framework*. (Maps HIPAA requirements to a control framework — useful for engineering teams that already use NIST.)
- NIST SP 800-66 Rev. 2 (2024). *Implementing the HIPAA Security Rule*. (NIST's operational guide; the engineering-facing companion to the regulation.)
- HHS. *Guidance Regarding Methods for De-identification of Protected Health Information*. (Safe Harbor and Expert Determination methods.)
- HITRUST CSF. *Common Security Framework*. hitrustalliance.net. (Industry-standard control framework; commonly required by US healthcare partners as the operational layer above HIPAA.)
