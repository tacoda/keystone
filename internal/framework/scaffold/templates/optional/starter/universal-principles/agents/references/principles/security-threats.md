# Security Threat Classes

Operational catalog of the vulnerability and attack classes that secure design must defend against. Companion to [[security]], which states the *principles*; this file states the *threats* those principles exist to mitigate.

The catalog has two parts: the **OWASP Top 10** (the canonical web-application vulnerability taxonomy, updated by OWASP every 3–4 years) and **modern threats** that have crossed into the mainstream since the 2010s — supply chain attacks chief among them.

Threat rankings shift; the underlying categories are durable. Each entry below names the Saltzer-Schroeder principle it most directly violates, so the catalog stays anchored to first principles even as the ranking churns.

> **Rules extracted:** [`guides/principles/security-threats.md`](../../guides/principles/security-threats.md). This file holds the full reasoning, anti-patterns, and references.

## OWASP Top 10 (2021)

The current OWASP Top 10. Re-check the source periodically — OWASP updates the ranking and merges/splits categories as the landscape shifts.

### A01 — Broken Access Control
Failures in enforcing what authenticated users are allowed to do. Vertical (privilege escalation), horizontal (one user accessing another's data), or insecure direct object references (changing `/orders/123` to `/orders/124` and getting someone else's order). Violates **complete mediation** and **least privilege**.

### A02 — Cryptographic Failures
Sensitive data exposed because crypto was missing, weak, or misused. Plaintext storage of credentials, weak hashing (MD5, SHA-1, unsalted), hardcoded keys, missing TLS, broken random number generation, vulnerable cipher modes. Violates **open design** (rolling your own crypto) and **fail-safe defaults**.

### A03 — Injection
Untrusted input interpreted as code or commands by an interpreter — SQL, NoSQL, OS command, LDAP, XPath, ORM, template, expression-language injection. The fix at the principle level is **separation** of code from data; at the idiom level, parameterized queries, prepared statements, safe-by-default templates. See [[postels-law]] on strict input acceptance.

### A04 — Insecure Design
Vulnerabilities baked into the architecture, not the implementation. Missing rate limiting on credential endpoints, lack of secure-by-default patterns, threat models never written. Cannot be fixed by patching code; requires re-design. The category OWASP added to call attention to this in 2021.

### A05 — Security Misconfiguration
Default credentials, debug endpoints exposed in production, verbose error pages, missing security headers, permissive CORS, overly-broad cloud IAM roles, public S3 buckets. Violates **fail-safe defaults** and **least privilege**.

### A06 — Vulnerable and Outdated Components
Using a library, framework, runtime, or container image with known CVEs. The substrate has become as important as the application code; the attacker's first move is often a public vulnerability scan against your dependency manifest. Bridges directly into the **supply chain** category below.

### A07 — Identification and Authentication Failures
Weak password policies, credential stuffing tolerance, missing or broken MFA, predictable session tokens, session fixation, missing logout, "remember me" cookies that never expire. Violates **complete mediation** (sessions outliving their authority).

### A08 — Software and Data Integrity Failures
Code or data that is loaded from a source whose integrity is not verified. Insecure deserialization, unsigned updates, untrusted package sources, CI/CD pipelines that pull and execute arbitrary code from third-party actions. The OWASP category that absorbs much of the **supply chain** threat surface.

### A09 — Security Logging and Monitoring Failures
You cannot respond to what you cannot see. Missing audit logs on authentication events, no alerting on anomalies, logs that can be tampered with by the same accounts they audit. Violates **complete mediation** in spirit — a control you cannot observe is a control you cannot trust. See [[security]]'s "designs that are auditable."

### A10 — Server-Side Request Forgery (SSRF)
The server can be tricked into making requests to internal addresses on behalf of an attacker — cloud metadata services (`169.254.169.254`), internal admin panels, internal databases. Devastating in cloud environments where the metadata endpoint hands out IAM credentials. Violates **least privilege** (the server should not be able to reach what it does not need) and the network-layer form of **narrow trust boundaries**.

## Modern threats

Classes that have crossed into mainstream prominence since the 2010s, beyond what the OWASP Top 10 captures directly.

### Supply chain attacks
The attacker compromises something *upstream* of your code — a dependency, a package registry, a CI provider, a build system, a code-signing key — and your build pipeline delivers the malicious code to your users.

Canonical incidents:

- **SolarWinds Orion** (2020) — build-system compromise inserted a backdoor into a signed update consumed by ~18,000 organizations.
- **log4shell** (CVE-2021-44228) — a remote-code-execution vulnerability in Apache Log4j, present transitively in vast portions of the JVM ecosystem; exploitation required only that a log line contain attacker-controlled text.
- **XZ Utils backdoor** (CVE-2024-3094) — a multi-year social-engineering campaign installed a maintainer who then introduced a backdoor in the `xz-utils` package, near-shipped into mainstream Linux distributions before discovery.
- **Package-registry attacks** — typosquatting (`reqeusts` for `requests`), dependency confusion (a private package name resolved to a malicious public package), maintainer-account takeover, post-install scripts that exfiltrate environment variables.

Defenses are layered, not single-point: pinned versions and lockfiles, signed artifacts, provenance attestations (SLSA), reproducible builds, software bills of materials (SBOM), minimal base images, build isolation, dependency review on changes to lockfiles. Violates **least common mechanism** at scale — every shared dependency is shared infrastructure with shared blast radius.

### Secrets in source control, CI logs, and AI assistants
Credentials accidentally committed (still the single most common breach root cause for cloud incidents), exposed in CI build logs, or pasted into AI assistants and code-review tools that retain them. Mitigated by secret scanning (pre-commit and post-commit), short-lived credentials, OIDC-federated auth to cloud providers (no long-lived keys at all), and treating any exposed secret as compromised — *rotate, don't redact*. Reinforces [[security]]'s IRON LAW (**NO SECRETS IN SOURCE**).

### Cloud and IAM misconfiguration
The blast radius of a single over-permissioned IAM role in a cloud environment is the cloud-era equivalent of the buffer overflow — small mistake, total compromise. Common forms: wildcard `iam:*` policies, cross-account trust with no condition keys, public storage buckets, service accounts with persistent keys instead of workload identity, IMDSv1 endpoints reachable from compromised application code (the SSRF pathway). Violates **least privilege** at the infrastructure layer.

### Confused deputy
A privileged component is tricked into using its authority on behalf of an unauthorized caller. Classic in cloud (cross-tenant access via an over-trusting service), in OS internals (the original 1988 Hardy paper), and in OAuth flows where a client is tricked into granting consent on behalf of a third party. The countermeasure is to make authority *explicit* at the point of use — capabilities, audience-scoped tokens, condition keys — never ambient.

### Session and token theft
Cookies and bearer tokens stolen via XSS, malicious browser extensions, infostealer malware on developer machines, or man-in-the-middle on misconfigured TLS. Bearer tokens with no binding to the requesting client are the weak link. Mitigations: short-lived access tokens, refresh-token rotation, token binding (DPoP, mTLS-bound tokens), httpOnly/Secure/SameSite cookies, device-bound credentials where the platform supports it.

### Prompt injection in AI-integrated applications
LLM-backed systems that consume untrusted input (user messages, fetched web pages, retrieved documents) can be steered by **adversarial instructions inside that input** into ignoring their original system prompts, exfiltrating data, or taking unauthorized actions through their tools. A rapidly evolving class; the principles still apply — **trust boundary** at the input, **least privilege** for tool access, **complete mediation** of every action the model can take.

## What this catalog asks of you

- When you build something user-facing, walk the OWASP Top 10 against your design. Most categories will not apply; the discipline is to know *why* each does not apply, not to skip the walk.
- When you add a dependency, treat it as code you wrote — same review bar, same security posture, same expectation that it could be malicious. See [[security]] (least common mechanism).
- When you handle a secret, assume it will eventually leak. Design for *rotation*, not for *secrecy in perpetuity*.
- When a control depends on the attacker not knowing something other than a secret, it is not a control. See [[security]] (open design).

## References

- OWASP Foundation. *OWASP Top 10: 2021*. owasp.org/Top10/. (Re-check for the current edition.)
- OWASP Foundation. *OWASP Application Security Verification Standard (ASVS)*. (Operational checklist that elaborates the Top 10.)
- NIST SP 800-218 (2022). *Secure Software Development Framework (SSDF)*.
- NIST SP 800-161r1 (2022). *Cybersecurity Supply Chain Risk Management Practices for Systems and Organizations*.
- SLSA Project. *Supply-chain Levels for Software Artifacts*. slsa.dev. (Provenance and build-integrity framework.)
- CISA & FBI (2021). *Alert AA20-352A: Advanced Persistent Threat Compromise of Government Agencies, Critical Infrastructure, and Private Sector Organizations* (SolarWinds).
- Apache Software Foundation (2021). *CVE-2021-44228* (Log4Shell).
- Freund, A. (2024). Disclosure of the *xz-utils* backdoor (CVE-2024-3094) to the oss-security mailing list.
- US Executive Order 14028 (2021). *Improving the Nation's Cybersecurity*. (Source of the SBOM mandate that has reshaped procurement.)
- Hardy, N. (1988). "The Confused Deputy: (or why capabilities might have been invented)." *ACM SIGOPS Operating Systems Review*, 22(4).
- Greenberg, A., & various authors at *Wired*, *Krebs on Security*, *The Record* — ongoing reporting is the most current source for novel incident classes.
