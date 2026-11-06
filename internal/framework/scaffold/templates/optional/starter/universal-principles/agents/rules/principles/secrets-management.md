# Secrets Management — rules

The non-negotiable and strongly-preferred rules from [`references/principles/secrets-management.md`](../../references/principles/secrets-management.md).
Loaded ambient; enforced by the [drift hook](../../hooks/drift.md). The reference file holds the full reasoning, anti-patterns, and references.

## IRON LAWS

**ROTATE, DON'T REDACT.** A secret that has appeared in source control, a chat log, a screenshot, a CI build log, an error report, or an AI assistant's context window must be treated as compromised. Removing the *visible* copy does not remove the copies that were already harvested. Rotate the underlying credential; treat the redaction as a courtesy, not a fix.

**EVERY SECRET HAS AN EXPIRY.** If a secret has no rotation date, the rotation date is "never," and the secret will outlive the threat model it was issued under. Long-lived secrets must either be replaced by short-lived ones or scheduled into a rotation cadence — there is no third option.

---

Traces to: [`references/principles/secrets-management.md`](../../references/principles/secrets-management.md).
