# 0007 — No backward compatibility at 1.0

**Status:** Accepted
**Date:** 2026-06-08

## Context

The 0.x line passed through several layout shapes: tier enum (Org/Team/Project), YAML config, an embedded-plugins draft, three-tier migrations. Carrying every prior shape forward as a runtime shim would weigh down the 1.0 framework with code paths the new model does not need.

1.0 is, by definition, the moment to break.

## Decision

- 1.0 is a clean break from 0.x. No silent shims, no auto-migration of 0.x layouts, no dual-format readers.
- The 0.x → 1.0 path is a documented one-time `keystone init --reset --i-understand-this-is-destructive`, performed by the user after they commit and tag their 0.x state for reference.
- The existing 0.x migration chain (`migrations/0.7.0/...0.13.0/`) is **removed**, not converted to JSON, in the Phase 1 sweep.
- Post-1.0 deprecation cycle: one minor release with a warning shim before removal in the next major. (Applies to *future* breaks. The 0.x → 1.0 break itself is the one-time exception.)

## Consequences

- Positive: The framework carries one layout, one config format, one set of conventions. Less surface to test, document, and explain.
- Positive: Refactors and new ports do not have to consider how old shapes still work.
- Negative: Users with extensive 0.x customization re-port by hand. Mitigated by: the user's own git tag of pre-1.0 state, the upgrade guide's `git diff <tag>` recipe, and a `.keystone-reset.diff` artifact written by `init --reset` (planned, Phase 6).
- Negative: A visible compatibility break may produce user friction at the moment of release. Acceptable — 1.0 is the contract, not 0.x.
- Neutral: External plugins authored against 0.x must publish a 1.0-shaped tag before consumers can vendor them at 1.0. The plugin author's deprecation timeline is their own.

## Alternatives considered

- **Shim every 0.x layout through 1.0** — rejected. The shim code becomes permanent maintenance tax for a one-time event.
- **Two majors of overlap (0.x readers in 1.x)** — rejected. Pre-1.0 is not a stable contract; forward-compat from before stability is not owed.
- **Auto-migrate 0.x → 1.0 inside the binary** — rejected. The 0.x layout is heterogeneous (users customized variably); a generic auto-migrate would either be wrong for many installs or carry as much shim code as option 1.
