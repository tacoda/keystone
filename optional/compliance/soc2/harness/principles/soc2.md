# SOC 2 (Service Organization Control 2)

An attestation report, not a certification, defined by the American Institute of Certified Public Accountants (AICPA). Service organizations engage an independent CPA firm to assess their controls against the AICPA's **Trust Services Criteria**; the auditor's report is then shared with the organization's customers as evidence that controls exist and operate effectively.

A SOC 2 report is **the de facto contract requirement** for B2B SaaS in North America. Selling to enterprise customers usually requires a current SOC 2 Type II. Unlike HIPAA, PCI, or GDPR, SOC 2 is not law — but its market-gating role makes it effectively non-optional for many businesses.

This file states the engineering-relevant principles. It is **not legal or audit advice**; the program is run with auditors and is itself an organizational discipline.

## Type I vs. Type II

- **Type I** — controls *exist* at a point in time. Useful for an early-stage company; "we have a SOC 2" usually means Type I, and sophisticated customers know this.
- **Type II** — controls **operated effectively** over a period (typically 6 to 12 months). This is the version enterprise customers actually accept. It requires evidence — logs, tickets, change records — sustained over the audit window.

The implication: passing Type II is not a moment, it is a quarter or more of consistent operational discipline. Engineering practices that produce the right artifacts continuously are the program; the audit observes those artifacts.

## The Trust Services Criteria

Five categories; **Security** is mandatory, the others are scoped in by the organization based on what they offer.

- **Security** (mandatory) — Protection against unauthorized access, both physical and logical. The "common criteria" that all SOC 2 reports include.
- **Availability** — Systems are available for operation and use as agreed. SLAs, uptime monitoring, incident response.
- **Processing integrity** — Processing is complete, accurate, valid, authorized. Audit trails on transactions.
- **Confidentiality** — Information designated confidential is protected. Encryption, access control, retention.
- **Privacy** — Personal information is collected, used, retained, disclosed, and disposed of according to commitments. Overlaps with [[gdpr]] obligations.

For engineering teams, **Security** maps almost entirely to what [[security]] and [[security-threats]] already state — SOC 2 demands evidence that those principles are operating, not different principles.

## What it asks of you

- When you build any system that touches customer data, every control needs to produce **evidence**. The control "we review access quarterly" is real only if the review is recorded; "we deploy via reviewed PRs" is real only if PR reviews leave a trail. Build the trail into the workflow.
- When you change production, the change has to be recorded as: *what changed, who approved, when, with what outcome*. Continuous Delivery practices (see [[continuous-delivery]]) give you most of this for free if the pipeline records who triggered each deploy.
- When you add an access path to a production system, that path needs a documented authorization step, MFA, and an audit log. Standing access is the most common SOC 2 finding. See [[security]] (least privilege).
- When you onboard or offboard an employee, the access changes need to land within a documented timeframe — typically same-day for high-risk access on termination. Identity provider integration (SCIM) is the engineering answer.
- When you handle an incident, document it. The incident, the response, the resolution, the post-mortem. The auditor is not looking for "no incidents" (impossible); they are looking for "incidents are recognized, handled, and learned from."
- When you choose a sub-processor or vendor that handles customer data, you take on their controls as part of yours. Vendor security review — including their SOC 2 — becomes part of procurement. See [[security-threats]] (supply chain).

## IRON LAW

**A control without evidence is not a control.** The auditor cannot accept "we do this" as proof of doing it; they need the artifacts — logs, tickets, screenshots, signed reviews. Build operational practices that produce evidence as a by-product; do not build them in parallel to a separate "compliance ceremony."

## GOLDEN RULES

- **Aim for evidence that exists *because* of the work, not in addition to it.** If your access reviews are quarterly tickets that take an afternoon, the rest of the team is doing it because of the ticket — not because the review is meaningful. The audit will accept it; the underlying control is weak.
- **Aim for one identity provider, one access management surface.** Federated SSO with SCIM for provisioning/deprovisioning eliminates a class of findings before they arrive.
- **Aim for change management that the pipeline enforces.** A required PR review, a required test gate, a required code-owner approval — these are controls *the system* enforces, evidence-producing by construction. See [[continuous-delivery]].
- **Aim for observability that maps to the criteria.** Availability is measured; security events are alerted; processing integrity has audit logs. The auditor's "show me" questions become dashboards. See [[observability]].
- **Aim for vendor reviews as a checkpoint in onboarding new tools.** A SaaS adopted by one team becomes everyone's compliance exposure.

## Anti-patterns

- A "compliance" workstream parallel to engineering, with engineers screenshotting evidence after the fact. The audit passes the first time; sustaining it across years is the failure mode.
- Standing root access to production for engineers. The auditor will ask; the answer "we trust them" is not the answer they accept. See [[security]].
- Manual access provisioning by emailing IT. No SCIM; no audit log of access changes; offboarding lag of days. The single most expensive control to remediate post-audit.
- A new SaaS adopted by a team without security review. The auditor will discover it; the data flow is now part of scope, with unclear controls.
- Reactive vulnerability patching with no SLA. The auditor wants evidence of *timely* remediation; "we got around to it" is not timely.
- A change-management process that everyone routes around because it's slow. The bypass is the control failure.

## References

- AICPA (2017, updated 2022). *Trust Services Criteria for Security, Availability, Processing Integrity, Confidentiality, and Privacy*. (The standard.)
- AICPA. *SOC 2® - SOC for Service Organizations: Trust Services Criteria*. aicpa-cima.com. (Practitioner-facing materials.)
- Center for Internet Security. *CIS Controls v8*. cisecurity.org. (Industry-standard control framework that maps cleanly to SOC 2.)
- NIST Cybersecurity Framework 2.0 (2024). nist.gov/cyberframework. (Higher-level control framework many SOC 2 programs use as their internal organizing principle.)
- Cooke, K. (2023). *SOC 2 for Engineers: What You Actually Need to Know*. (Industry-written practical guides — multiple exist; the genre is useful regardless of which.)
