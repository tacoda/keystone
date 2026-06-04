# SOC 2 (Service Organization Control 2) — rules

The rules from [`corpus/soc2.md`](../corpus/soc2.md).
Loaded ambient; enforced by the [drift sensor](../../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**A control without evidence is not a control.** The auditor cannot accept "we do this" as proof of doing it; they need the artifacts — logs, tickets, screenshots, signed reviews. Build operational practices that produce evidence as a by-product; do not build them in parallel to a separate "compliance ceremony."

## GOLDEN RULES

- **Aim for evidence that exists *because* of the work, not in addition to it.** If your access reviews are quarterly tickets that take an afternoon, the rest of the team is doing it because of the ticket — not because the review is meaningful. The audit will accept it; the underlying control is weak.
- **Aim for one identity provider, one access management surface.** Federated SSO with SCIM for provisioning/deprovisioning eliminates a class of findings before they arrive.
- **Aim for change management that the pipeline enforces.** A required PR review, a required test gate, a required code-owner approval — these are controls *the system* enforces, evidence-producing by construction. See [[continuous-delivery]].
- **Aim for observability that maps to the criteria.** Availability is measured; security events are alerted; processing integrity has audit logs. The auditor's "show me" questions become dashboards. See [[observability]].
- **Aim for vendor reviews as a checkpoint in onboarding new tools.** A SaaS adopted by one team becomes everyone's compliance exposure.

---

Traces to: [`corpus/soc2.md`](../corpus/soc2.md).
