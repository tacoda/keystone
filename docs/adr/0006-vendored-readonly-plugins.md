# 0006 — Vendored read-only plugins

**Status:** Accepted
**Date:** 2026-06-08

## Context

Plugins exist to ship policy from outside a project. Users need three properties:

1. **Reproducibility** — the same plugin at the same tag produces the same content for everyone who installs it.
2. **Tamper-resistance** — a teammate accidentally (or deliberately) editing a vendored file should not silently change the cascade.
3. **Cleanliness** — plugin content should not bloat the project's git history.

0.x already pins per-source SHAs and per-file hashes in `.keystone.lock`. The user asked for an additional layer: on any drift between the vendored copy and the pinned tag's content, wipe and reinstall from scratch.

## Decision

- Plugins live under `harness/plugins/<name>/`, **gitignored**, materialized by `keystone install` from `keystone.json` + `keystone.lock.json`.
- Install flow:
  1. Resolve `source@version` → git-clone into a content-addressable cache at `~/.cache/keystone/plugins/<sha>/`.
  2. Copy from cache into `harness/plugins/<name>/`.
  3. Compute per-file SHA-256 and write to `keystone.lock.json`.
  4. Chmod 0444 where the OS supports it (best-effort UX hint, not enforcement).
- **Before every cascade resolution**, the runtime walks `harness/plugins/<name>/`, recomputes hashes, and compares them to the lockfile. Any drift — extra file, missing file, modified content — triggers `rm -rf harness/plugins/<name>/` followed by a fresh re-materialize from cache. No partial recovery, no warning prompt.
- Read-only filesystem marking is a UX hint. The hash check is the enforcement.
- Editing a plugin file is unsupported and silently reverted on the next run, with a log line naming the reset and the offending file(s).

## Consequences

- Positive: Plugins are reproducible from `keystone.json` + the lockfile alone — git clone the project, `keystone install`, run.
- Positive: Local edits to plugin content cannot accumulate. "Works on my machine" is excluded by construction.
- Positive: The project repo stays lean — no vendored markdown in git.
- Positive: The drift-reset is a single layer of protection above the existing SHA pin; the pin guarantees the source tag hasn't moved, the per-file hash guarantees the local copy hasn't been edited.
- Negative: Network is required on a fresh clone before plugins resolve. Mitigated by the local content-addressable cache (`~/.cache/keystone/plugins/`).
- Negative: Read-only chmod is POSIX-only; Windows falls back to the hash check alone. Acceptable — the hash check is the real enforcement on every platform.
- Neutral: Plugin authors must publish via tag. Every consumer pin is a tag; every reinstall is reproducible.

## Alternatives considered

- **Committed (Go vendor/ style)** — rejected. Bloats history, signals "edit this" exactly where edits must not happen, and drift-reset would produce constant no-op churn in PRs.
- **Read-write plugins with hash warning only** — rejected. Warnings get ignored; the user asked for hard enforcement.
- **Per-user global install of plugins** — rejected. Two projects on the same machine with different pinned versions would fight over a shared install directory.
- **Drift-reset on explicit commands only (not every load)** — partial; resolved via the verification cache for the duration of a single process invocation (an open question in PLAN §6, default: every load).
