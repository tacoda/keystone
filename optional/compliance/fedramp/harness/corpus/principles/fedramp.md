# FedRAMP (Federal Risk and Authorization Management Program)

The US federal government's standardized approach to security assessment, authorization, and continuous monitoring for cloud products and services used by federal agencies. Established in 2011 under OMB Memorandum M-11-29; restructured under the *FedRAMP Authorization Act of 2022* (passed December 2022). Run by the General Services Administration (GSA) in coordination with the Joint Authorization Board (JAB) and individual federal agencies.

If a cloud service is consumed by a federal agency, it must hold a current FedRAMP **Authorization to Operate (ATO)**. The program is the highest-bar compliance regime most commercial engineering teams encounter — substantially more rigorous than SOC 2, with overlap to FISMA, NIST 800-53, and (for DoD) DoD CC SRG and CMMC.

This file states the engineering-relevant principles. It is **not legal advice**; the program is run with a 3PAO (Third Party Assessment Organization), legal counsel, and (often) a dedicated federal compliance team.

> **Rules extracted:** [`guides/principles/fedramp.md`](../../guides/principles/fedramp.md). This file holds the full reasoning, anti-patterns, and references.

## Impact levels

A FedRAMP package is scoped to one of three impact levels, defined by FIPS 199:

- **Low** — Loss of confidentiality, integrity, or availability would have *limited* adverse effect. ~125 NIST 800-53 controls.
- **Moderate** — *Serious* adverse effect. The most common level for commercial SaaS. ~325 controls.
- **High** — *Severe or catastrophic* adverse effect. Reserved for systems where loss would meaningfully harm life, finances, or the mission. ~425 controls.

A separate **FedRAMP Tailored** (Low LI-SaaS) baseline exists for low-impact, low-data SaaS. The vast majority of commercial cloud offerings selling to civilian agencies target **Moderate**.

## What the engineering team owns

FedRAMP draws controls from NIST SP 800-53 Rev. 5. The control families relevant to engineering decisions:

- **AC — Access Control.** Identity, authorization, separation of duties, least privilege, session management. See [[security]].
- **AU — Audit and Accountability.** Comprehensive logging, log retention, tamper protection, log review. See [[observability]].
- **CM — Configuration Management.** Baseline configurations, change control, software inventory, separation of dev/test/prod environments.
- **CP — Contingency Planning.** Backup, restore, alternate processing site, exercises.
- **IA — Identification and Authentication.** MFA (PIV/CAC for privileged users in many cases), federation, identifier management.
- **IR — Incident Response.** Detection, reporting, handling, post-incident analysis with strict reporting timelines.
- **SA — System and Services Acquisition.** Supplier risk, secure development lifecycle, supply chain integrity. See [[security-threats]] (supply chain).
- **SC — System and Communications Protection.** Network segmentation, encryption in transit, cryptographic protections, denial-of-service protections.
- **SI — System and Information Integrity.** Flaw remediation, malicious code protection, monitoring, integrity verification.
- **SR — Supply Chain Risk Management.** SBOMs, provenance, vendor risk.

Most of these map directly to engineering principles already in this corpus; FedRAMP enforces them at a level of **evidence and rigor** higher than other regimes.

## The hard parts

The technical bar is achievable; the operational bar is what most commercial teams underestimate:

- **FIPS 140-3 validated cryptography.** Every cryptographic module used to protect federal data must be FIPS-validated. Most general-purpose libraries are not validated by default; AWS, Azure, and GCP provide FIPS-enabled endpoints, but their use is opt-in and substantially constrained.
- **US-based personnel for the system boundary.** Operations, on-call, and incident response staff need US-person status; for some agencies, US-citizen status. The system boundary cannot rely on engineers outside the US.
- **Continuous monitoring (ConMon).** Monthly vulnerability scans, monthly POA&M (Plan of Action and Milestones) updates, quarterly security reviews — all delivered to the agency or JAB and reviewed.
- **Significant change management.** Material changes to the authorized system require notification, sometimes re-authorization. The "ship fast" model from non-FedRAMP environments has explicit limits.
- **Government cloud regions.** AWS GovCloud, Azure Government, GCP Assured Workloads, etc. Different APIs, different feature availability, often different operational paths.

The compliance program is therefore an *architectural* commitment — not a checkbox at the end. A system not designed from the start for FedRAMP will need significant rework to achieve it.

## What it asks of you

- When you design a system intended for FedRAMP, treat the **boundary** as a first-class architectural artifact. Every component inside is in scope; every connection across must be authorized. See [[separation-of-concerns]] applied at the network and identity layers.
- When you choose cryptographic primitives, choose FIPS-validated modules from day one. Retrofitting non-FIPS code is expensive. The cloud provider's "FIPS endpoints" must be the default, not the exception.
- When you log, log enough to support **forensic reconstruction** of any security-relevant event. SIEM integration, central log retention, tamper-evident storage. See [[observability]].
- When you operate, the on-call rotation, the build pipeline, the deployment surface — all are in the boundary. Off-boundary tooling and personnel become significant change events.
- When you depend on a third-party service, that service must itself be FedRAMP-authorized (at equal or higher impact level) for use inside the boundary. The dependency graph is a constraint.
- When you respond to an incident, the reporting timelines are short and the language is specific. Pre-defined playbooks, pre-trained personnel, pre-stated channels.

## Anti-patterns

- A "we'll FedRAMP it later" plan added halfway through commercial-first development. The architectural decisions already made — multi-tenant data, non-FIPS crypto, global engineering team — each cost a quarter or more to undo.
- Privileged access from non-US-person engineers. Not "we mostly don't"; the auditor wants evidence the system enforces it.
- Use of a non-authorized SaaS for in-boundary work (a chat tool, a debugger, an AI assistant). The data flow is in scope; the SaaS is not authorized; finding.
- Treating ConMon as ceremony rather than operating practice. The monthly cadence demands real remediation, not just reporting.
- Logs not tamper-evident; logs editable by the same identity that performs sensitive operations.
- Confusing FedRAMP Low with the bar for SaaS used by a federal agency. Most agency contracts require Moderate at minimum; Low is a narrow LI-SaaS category.

## References

- *FedRAMP Authorization Act of 2022*, Pub. L. 117-263, § 5921. (Statutory basis.)
- FedRAMP Program Office. *FedRAMP Documents and Templates*. fedramp.gov/documents. (The forms; the SSP template alone reshapes how teams document the system.)
- NIST SP 800-53 Rev. 5 (2020, updated). *Security and Privacy Controls for Information Systems and Organizations*. (The control catalog FedRAMP draws from.)
- FIPS PUB 199 (2004). *Standards for Security Categorization of Federal Information and Information Systems*. (Impact-level definitions.)
- FIPS PUB 140-3 (2019). *Security Requirements for Cryptographic Modules*. (The crypto validation standard.)
- NIST SP 800-37 Rev. 2 (2018). *Risk Management Framework for Information Systems and Organizations*. (The RMF FedRAMP operates within.)
- DoD Cloud Computing SRG (current revision). *Department of Defense Cloud Computing Security Requirements Guide*. (For DoD-targeted offerings; impact levels IL2/IL4/IL5/IL6 stack atop FedRAMP.)
