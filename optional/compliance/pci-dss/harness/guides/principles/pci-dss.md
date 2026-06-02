# PCI DSS (Payment Card Industry Data Security Standard) — rules

The rules from [`corpus/principles/pci-dss.md`](../../corpus/principles/pci-dss.md).
Loaded ambient; enforced by the [drift sensor](../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**Sensitive Authentication Data is never stored after authorization. The PAN, when stored, is encrypted; when displayed, is masked; when logged, is omitted.** The auditor will look for SAD in logs, in dev databases, in test fixtures, in support tickets, in chat. Every place it appears is a finding.

## GOLDEN RULES

- **Aim to reduce scope.** Each system that touches a PAN multiplies obligations. Tokenize at the edge; pass tokens internally; never let the PAN cross from the payment integration into the rest of the system.
- **Aim for a single-tenant CDE.** Mixing PCI-scoped and non-PCI systems on shared infrastructure pulls the non-PCI systems into scope.
- **Aim for masking by default everywhere.** A masking helper used at the display, log, and export boundary is much safer than per-call redaction.
- **Aim for change control with audit trails.** Every code change touching the CDE needs reviewed, approved, and recorded. The auditor will trace a change to its ticket.
- **Aim for quarterly vulnerability scans and annual penetration tests** as code-shipping cadence assumptions, not surprise events. The auditor expects evidence of both.

---

Traces to: [`corpus/principles/pci-dss.md`](../../corpus/principles/pci-dss.md).
