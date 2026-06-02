# HIPAA (Health Insurance Portability and Accountability Act) — rules

The rules from [`corpus/principles/hipaa.md`](../../corpus/principles/hipaa.md).
Loaded ambient; enforced by the [drift sensor](../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**Every access to PHI is authenticated, authorized, and logged.** No exceptions for service accounts, batch jobs, debugging sessions, or "trusted" internal users. An access path that bypasses the audit log is a HIPAA finding waiting to be discovered.

## GOLDEN RULES

- **Aim for the smallest PHI footprint.** De-identify where possible; project to non-PHI fields for downstream uses; redact in logs; expire data on a schedule. Each PHI field is exposure.
- **Aim for minimum-necessary access.** A clinician needs their patients; a researcher needs the dataset; a developer needs neither — *especially in production*. See [[security]] (least privilege).
- **Aim for tamper-evident audit logs.** Write-once, signed, exported to a destination outside the access scope of the systems being audited. The audit log that auditors look at is the one that wasn't editable.
- **Aim for break-glass access with alerting.** Emergency access exists; it must be exceptional, time-boxed, logged, and reviewed.

---

Traces to: [`corpus/principles/hipaa.md`](../../corpus/principles/hipaa.md).
