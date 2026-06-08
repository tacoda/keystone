# 0008 — Versioning policy

**Status:** Accepted
**Date:** 2026-06-08

## Context

Post-1.0, users will pin Keystone and expect to know what they can depend on across minor versions. The framework needs an explicit promise about what is stable and what is free to evolve. Without that, every minor release is a gamble.

## Decision

**Scheme.** SemVer: `MAJOR.MINOR.PATCH`. Released via tag push to `goreleaser` (per project release policy).

**Stable across minor versions:**
- Port names and contracts (`docs/ports/<port>.md`).
- `keystone.json` schema.
- `keystone.lock.json` schema.
- Plugin manifest schema (`keystone-plugin.json`).
- Conventions table (`docs/conventions.md`).
- CLI surface: `init`, `install`, `verify`, `plugin add | update | remove`, `new <port>`, `doctor`, `migrate`, `version`.

**Free to evolve in minor versions:**
- `internal/framework/` package layout and APIs.
- Scaffold template *contents* (treated as content, not API — once scaffolded they live in user repos).
- Warning text and log format.
- Internal heuristics (e.g., token estimator in the budget port).

**Deprecation cycle.** Any stable surface that needs to break gets one minor release with a warning shim before removal in the next major.

**Plugin versioning.** Plugin manifests carry their own `version`. Plugin versioning is the plugin author's responsibility; the framework only pins what the consumer's `keystone.json` declares.

**Pre-release tags.** `-alpha`, `-beta`, `-rc` suffixes are used during phase rollout. Consumers can pin to them but receive no stability guarantee.

## Consequences

- Positive: Users can pin a major and trust their `keystone.json` and `harness/` layout to keep working.
- Positive: The framework can refactor Go internals freely without being held hostage by minor-version commitments.
- Positive: Plugin authors version independently; the framework does not gate their cadence.
- Negative: Template content drift versus user-edited copies is unsolved by version policy alone — handled at the `doctor` layer (Phase 4).
- Neutral: The `migrate` runner is reserved for stable-surface schema bumps only (see PLAN §5).

## Alternatives considered

- **CalVer (`YEAR.MONTH.PATCH`)** — rejected. Communicates calendar, not compatibility. Users want to know what breaks, not when.
- **ZeroVer (stay on 0.x indefinitely)** — rejected. The framework is mature enough at 1.0 to make a stability promise; users have asked for one.
- **Stricter promise (everything stable across majors)** — rejected. The framework needs room to refactor internals; over-promising leads to either broken promises or stagnation.
