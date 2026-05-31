# Security Principles

Foundational design principles for systems that are secure by construction, not by accident. Drawn from Jerome Saltzer and Michael Schroeder's "The Protection of Information in Computer Systems" (*Proceedings of the IEEE*, 1975) — the canonical articulation that still grounds modern security engineering.

Stack-specific countermeasures (parameterized queries, CSRF tokens, output encoding, dependency pinning) belong in `../idioms/<stack>/`. The principles below are what those countermeasures *instantiate*.

## The eight Saltzer-Schroeder principles

### Economy of mechanism
Keep the design as simple and small as possible. Complex security mechanisms are harder to audit and more likely to contain unnoticed flaws.

### Fail-safe defaults
Base access decisions on permission rather than exclusion. The default is *deny*; access is granted explicitly. A misconfigured or partially-implemented system should be safe, not open.

### Complete mediation
Every access to every object must be checked for authority. Caches of authorization decisions that survive permission changes are a complete-mediation violation.

### Open design
The security of a mechanism should not depend on the secrecy of its design. Keys, passwords, and tokens are secrets; algorithms and architecture are not.

### Separation of privilege
A protective mechanism that requires two keys is more robust than one that allows access on the basis of a single key. Multi-party authorization, two-person rules, separation of duties.

### Least privilege
Every program and user should operate using the least set of privileges necessary to complete the job. Privilege scopes should be narrow in space (which resources) and time (how long).

### Least common mechanism
Minimize the amount of mechanism common to more than one user or security boundary. Shared mechanism is a covert channel; shared state across security boundaries is the source of side-channel attacks.

### Psychological acceptability
The human interface must be designed for ease of use, so users routinely and automatically apply the protection mechanisms correctly. Security that users circumvent because it gets in their way is no security.

## IRON LAWS

**FAIL CLOSED.** When in doubt, deny. A bug in the authorization path that grants access is a security incident; a bug that denies access is a usability incident. The asymmetry is intentional.

**NO SECRETS IN SOURCE.** Keys, tokens, credentials, signing material — never committed to a repository, never in a tracked config file, never in error messages or logs. Secrets live in environment variables, secret managers, or hardware modules.

## GOLDEN RULES

- **Aim for narrow trust boundaries.** Every component should distrust its inputs. The boundary between trusted and untrusted should be small, explicit, and reviewable.
- **Aim for defense in depth.** No single control is a complete defense. Pair authentication with authorization; pair authorization with auditing; pair encryption-in-transit with encryption-at-rest.
- **Aim for designs that are auditable.** A security control whose effect cannot be observed is a control you cannot trust.

## Anti-patterns

- "Security by obscurity" — relying on attackers not finding the secret URL / endpoint / format.
- Permissions that opt-out (default-allow) instead of opt-in (default-deny).
- Caching authorization decisions across permission changes.
- Sharing one privileged account across services or humans.
- Logging request bodies that may contain credentials.
- Treating cryptographic algorithms as proprietary trade secrets.

## References

- Saltzer, J. H., & Schroeder, M. D. (1975). "The Protection of Information in Computer Systems." *Proceedings of the IEEE*, 63(9), 1278–1308.
- Anderson, R. (2020). *Security Engineering*, 3rd ed. Wiley.
- OWASP Foundation. *OWASP Top 10*. (Updated periodically — operational guidance derives from these principles.)
- NIST SP 800-160. *Systems Security Engineering: Considerations for a Multidisciplinary Approach in the Engineering of Trustworthy Secure Systems*.
