# PCI DSS (Payment Card Industry Data Security Standard)

A contractual standard, not a law. Maintained by the PCI Security Standards Council (founded 2006 by Visa, Mastercard, American Express, Discover, JCB); now at version 4.0 (effective March 2024) with v3.2.1 retired March 2024. Any organization that **stores, processes, or transmits cardholder data** must comply, on penalty of fines from the card brands and the inability to process card payments.

This file states the engineering-relevant principles. It is **not legal advice**; PCI DSS has merchant levels, attestation requirements, and audit obligations that require a QSA (Qualified Security Assessor) to navigate.

## The 12 high-level requirements

PCI DSS v4.0 retains the same 12 top-level requirements as previous versions, grouped into six control objectives:

1. **Install and maintain network security controls** (firewalls, segmentation).
2. **Apply secure configurations** to all components.
3. **Protect stored account data** (the cardholder data — see below).
4. **Protect cardholder data with strong cryptography during transmission** over public networks.
5. **Protect all systems and networks from malicious software.**
6. **Develop and maintain secure systems and software.**
7. **Restrict access** to system components and cardholder data by business need to know.
8. **Identify users and authenticate** access to system components.
9. **Restrict physical access** to cardholder data.
10. **Log and monitor all access** to system components and cardholder data.
11. **Test security of systems and networks regularly.**
12. **Support information security with organizational policies and programs.**

Each requirement decomposes into specific sub-requirements with required/testing evidence.

## What counts as cardholder data (CHD)

The PCI vocabulary distinguishes carefully:

- **Cardholder Data (CHD)** — PAN (Primary Account Number), cardholder name, expiry, service code.
- **Sensitive Authentication Data (SAD)** — full magnetic stripe data, CVV/CVC, PIN/PIN block. **Must never be stored after authorization.** This is the single most-violated PCI rule.

The PAN is the central concept. Storing it triggers most of the standard's obligations. The dominant compliance strategy is therefore **scope reduction** — minimize the systems that ever see a PAN, ideally by handing off to a PCI-compliant payment processor before the PAN ever reaches your application.

## What it asks of you

- When you design a payment flow, ask first: *can our application avoid touching the PAN entirely?* Hosted payment fields, redirect/iframe-based payment pages, and tokenization services (Stripe Elements, Braintree Hosted Fields, Adyen drop-in) reduce scope dramatically. The simplest PCI scope is *none*. See [[simplicity]].
- When you must store cardholder data, encrypt the PAN at rest with strong cryptography and managed keys. The PAN may be stored; CVV/CVC may not, ever, post-authorization.
- When you transmit CHD over public networks, TLS 1.2+ with strong ciphers, mutual auth where applicable. Email, SMS, and unencrypted chat are forbidden as transports.
- When you log, **do not log the PAN.** Mask it (first 6 + last 4 is the standard mask) before any log line is written. The CVV is forbidden in logs in any form. See [[secrets-management]] (rotate, don't redact).
- When you authenticate, MFA is required for all non-console access into the CDE (Cardholder Data Environment) and for all administrative access regardless of source. Passwords must rotate, complexity standards apply, the works. See [[security]] and [[secrets-management]].
- When you provision access, restrict by **business need to know.** A standing developer access path into production card data is a finding. Break-glass with alerting, time-boxed, audited. See [[security]] (least privilege).
- When you change a system in scope, the change has security implications and audit trail requirements. CI/CD pipelines for the CDE need their own controls.

## IRON LAW

**Sensitive Authentication Data is never stored after authorization. The PAN, when stored, is encrypted; when displayed, is masked; when logged, is omitted.** The auditor will look for SAD in logs, in dev databases, in test fixtures, in support tickets, in chat. Every place it appears is a finding.

## GOLDEN RULES

- **Aim to reduce scope.** Each system that touches a PAN multiplies obligations. Tokenize at the edge; pass tokens internally; never let the PAN cross from the payment integration into the rest of the system.
- **Aim for a single-tenant CDE.** Mixing PCI-scoped and non-PCI systems on shared infrastructure pulls the non-PCI systems into scope.
- **Aim for masking by default everywhere.** A masking helper used at the display, log, and export boundary is much safer than per-call redaction.
- **Aim for change control with audit trails.** Every code change touching the CDE needs reviewed, approved, and recorded. The auditor will trace a change to its ticket.
- **Aim for quarterly vulnerability scans and annual penetration tests** as code-shipping cadence assumptions, not surprise events. The auditor expects evidence of both.

## Anti-patterns

- Logging the full PAN "for debugging." The single most common finding; remediation is expensive and embarrassing.
- Storing CVV/CVC "for chargeback disputes." The standard is explicit: never.
- A developer-prod link with no MFA and no auditing. "It's only for debugging."
- A test database refreshed from prod with cardholder data left intact. The test environment is now in PCI scope.
- A Slack channel where the customer-support team pastes screenshots of the PAN field "to help the customer." See [[secrets-management]].
- Believing the QSA "won't notice" — every recurring finding compounds; the cost of remediation is greater than the cost of doing it right.
- Outsourcing to a "PCI-compliant processor" but having the unencrypted PAN flow through the application's own servers on the way to the processor. The application is still in scope.

## References

- PCI Security Standards Council (2022). *PCI DSS v4.0*. pcisecuritystandards.org/document_library. (The current standard; v4.0.1 is the latest revision as of 2024.)
- PCI Security Standards Council. *PCI DSS Quick Reference Guide v4.0*. (Plain-language summary; useful for engineers not ready to read the full standard.)
- PCI Security Standards Council. *Information Supplement: Best Practices for Maintaining PCI DSS Compliance*. (Operationalizing compliance beyond the audit.)
- Visa, Mastercard, et al. — each card brand publishes its own PCI program with overlapping but distinct obligations. The application's acquirer is the right starting point.
- NIST SP 800-53. *Security and Privacy Controls for Information Systems and Organizations*. (PCI's controls map cleanly onto NIST's families; useful for teams that operate against both.)
