# Security Principles — rules

The non-negotiable and strongly-preferred rules from [`corpus/principles/security.md`](../../corpus/principles/security.md).
Loaded ambient; enforced by the [drift sensor](../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAWS

**FAIL CLOSED.** When in doubt, deny. A bug in the authorization path that grants access is a security incident; a bug that denies access is a usability incident. The asymmetry is intentional.

**NO SECRETS IN SOURCE.** Keys, tokens, credentials, signing material — never committed to a repository, never in a tracked config file, never in error messages or logs. Secrets live in environment variables, secret managers, or hardware modules.

---

Traces to: [`corpus/principles/security.md`](../../corpus/principles/security.md).
