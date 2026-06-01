# Secrets Management

A *secret* is any value whose disclosure breaks the security model — credentials, API keys, OAuth client secrets, signing keys, encryption keys, session tokens, TLS private keys, database passwords, webhook signing secrets. The discipline of managing them is not "keep them hidden"; it is **managing the lifecycle of values you assume will eventually leak.**

This file states the principles. Stack-specific mechanisms (AWS Secrets Manager, HashiCorp Vault, GCP Secret Manager, Kubernetes external-secrets, sealed-secrets, sops, etc.) belong in `../idioms/<stack>/`. Companion to [[security]] (IRON LAW: NO SECRETS IN SOURCE) and to [[security-threats]] (the threat class that secrets-management defends against).

## The lifecycle

A secret has **six lifecycle stages.** A secrets-management program is the sum of how it handles each.

1. **Generation** — created by a CSPRNG of sufficient length, or issued by a trusted authority (CA, IdP, KMS). Never derived from low-entropy inputs (timestamps, usernames, predictable counters).
2. **Distribution** — delivered to the workload that needs it through a channel the secret never leaves in plaintext. *Pull* (workload fetches at startup from a secret store) is preferred to *push* (CI injects into a deployment artifact); both beat *bake* (secret committed to an image or config file).
3. **Storage at rest** — encrypted by a key managed by a separate system. The encryption key has its own lifecycle; that lifecycle is shorter and tighter than the data key's.
4. **Use** — held in memory only as long as needed; not logged, not serialized into error messages, not echoed to stdout, not exfiltrated through debug endpoints. See [[fail-fast]] — if a secret-handling path receives an unexpected error, fail loudly without including the secret.
5. **Rotation** — replaced on a schedule and on demand. The interval is short enough that an undetected compromise has bounded blast radius; the mechanism is automated so the interval can stay short.
6. **Revocation** — invalidated immediately when compromise is suspected, when an employee departs, when a workload retires. Revocation must propagate faster than the secret can be used; this is what limits how long-lived secrets are allowed to be.

A "secrets management" program that handles generation and storage but not rotation and revocation is half a program.

## The hierarchy of secret handling

From most secure to least, in order of preference:

1. **No secret at all** — workload identity federated to the cloud provider (AWS IRSA, GCP Workload Identity, Azure managed identity, OIDC token exchange to third parties). The strongest secret is the one that never exists.
2. **Short-lived, dynamically-issued credentials** — Vault dynamic database credentials, STS session tokens, OAuth tokens with minutes-to-hours TTL, mTLS certificates with short validity. Compromise has a built-in expiration.
3. **Long-lived secrets stored in a managed secrets system** — fetched at startup, held only in memory, never written to disk. Acceptable when (1) and (2) are not feasible.
4. **Environment variables injected by an orchestrator** — better than files, worse than (1)–(3). Vulnerable to process listing, child-process inheritance, and accidental logging.
5. **Files on disk** — encrypted at rest, permissions restricted, mounted read-only. Acceptable for bootstrap secrets the orchestrator itself needs.
6. **Source control** — **never**. See [[security]]'s IRON LAW.

The right level for a given secret is the highest one the platform supports for that workload. Choosing a lower level "for simplicity" is a debt.

## What it asks of you

- When you introduce a secret, ask: *can I replace it with workload identity?* If yes, that is the answer. If no, ask: *can it be short-lived?* If yes, that is the next answer.
- When you handle a secret in code, scope its lifetime to the operation that needs it. A long-lived in-memory cache of a secret is a long-lived attack surface.
- When you write an error path that includes context, redact known-secret fields explicitly. Implicit "we don't log that" is a wish, not a control. See [[fail-fast]] — fail loud, but not with the keys.
- When a secret is rotated, **the new secret must work before the old one is revoked.** Otherwise the rotation *is* an outage. Dual-secret windows are how this is done at scale.
- When a secret is suspected to have leaked, **rotate first, investigate second.** Investigation time is exposure time.

## IRON LAWS

**ROTATE, DON'T REDACT.** A secret that has appeared in source control, a chat log, a screenshot, a CI build log, an error report, or an AI assistant's context window must be treated as compromised. Removing the *visible* copy does not remove the copies that were already harvested. Rotate the underlying credential; treat the redaction as a courtesy, not a fix.

**EVERY SECRET HAS AN EXPIRY.** If a secret has no rotation date, the rotation date is "never," and the secret will outlive the threat model it was issued under. Long-lived secrets must either be replaced by short-lived ones or scheduled into a rotation cadence — there is no third option.

## GOLDEN RULES

- **Aim to make the right thing the easy thing.** If developers must choose between secure secret handling and shipping the feature, secure handling will lose. Invest in tooling — pre-commit secret scanning, environment-aware fetch helpers, paved-road examples — until the secure path is the default path.
- **Aim for blast-radius isolation.** One secret, one purpose, one workload, one environment. Secrets shared across environments or across services magnify every breach.
- **Aim for audit trails.** Every read of a high-value secret should be logged. An undetected exfiltration is the worst outcome; making detection cheap is worth its cost.
- **Aim for symmetric controls on humans and machines.** A production credential a human can read at will is a credential an attacker phishes the human to obtain. Break-glass, MFA-gated, time-boxed, audited reads — for humans too.

## Anti-patterns

- A `.env` file committed to the repository "just for local development." Local secrets become real secrets through the same git history.
- A long-lived AWS access key in an environment variable, rotated "when we get around to it."
- Secrets passed through CI as plain environment variables that print into build logs on failure.
- A "secrets manager" that every workload reads from with a single shared root token.
- An incident response that redacts the leaked secret from the commit but leaves it valid.
- Encryption keys checked into source control "because the data they protect isn't there."
- Logging the request body — including credentials, tokens, and signed payloads — on errors.
- A secret with no recorded owner, no rotation schedule, and no recorded purpose. ("Why does this exist?" — "It's been there for years.") See [[security-threats]] on supply chain and credential exfiltration.

## References

- NIST SP 800-57 Part 1 Rev. 5 (2020). *Recommendation for Key Management — Part 1: General*.
- NIST SP 800-63B (2017, with revisions). *Digital Identity Guidelines: Authentication and Lifecycle Management*.
- OWASP Foundation. *Secrets Management Cheat Sheet*. cheatsheetseries.owasp.org.
- Wagner, A., et al. (2017). *The Twelve-Factor App*, Factor III (Config). 12factor.net. (Origin of the "config in environment, never in code" discipline; widely adopted, with known limitations for high-value secrets.)
- CIS Benchmarks for cloud providers (AWS, Azure, GCP) — operational guidance on IAM and secret handling derived from the principles above.
- HashiCorp. *Vault Architecture and Threat Model*. (One of the more thorough public threat models for a secrets system; useful regardless of which manager you use.)
- Google. *BeyondProd: A New Approach to Cloud-Native Security* (2019). (Production architecture built on the assumption that long-lived secrets are not enough.)
