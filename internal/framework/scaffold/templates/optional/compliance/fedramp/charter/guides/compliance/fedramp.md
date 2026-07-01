# FedRAMP (Federal Risk and Authorization Management Program) — rules

The rules from [`corpus/fedramp.md`](../corpus/fedramp.md).
Loaded ambient; enforced by the [drift sensor](../../../sensors/drift.md). The corpus file holds the full reasoning, anti-patterns, and references.

## IRON LAW

**FIPS-validated crypto, MFA on every privileged access, comprehensive audit, US-person operations.** These are the four non-negotiables that distinguish FedRAMP from lesser regimes. A system missing any of them does not pass; cosmetic adjustments do not bridge the gap.

## GOLDEN RULE

- **Aim to architect for FedRAMP from the start, or not at all.** Mid-project conversion is the most expensive path. Estimate 9–18 months for first ATO from a non-FedRAMP starting point.
- **Aim for a separate authorized environment.** Mixing FedRAMP and commercial workloads on shared infrastructure expands the boundary; isolation is cheaper than co-mingling.
- **Aim for tooling that produces evidence by construction.** SIEM, IaC scanning, vulnerability management, change control — each leaves an artifact the 3PAO can sample. See [[continuous-delivery]] for how to make the pipeline a control surface.
- **Aim for the FedRAMP Marketplace before custom**: GovCloud's authorized services are large; whenever a feature is available from an authorized service, use it rather than building it.
- **Aim for the smallest viable boundary.** Every system in scope multiplies obligations.

---

Traces to: [`corpus/fedramp.md`](../corpus/fedramp.md).
