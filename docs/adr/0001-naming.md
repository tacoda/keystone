# 0001 — Naming

**Status:** Accepted
**Date:** 2026-06-08

## Context

The 0.x project ships as `keystone` — a CLI binary, a Homebrew tap, an install script, and a documented public name. Repositioning to a "harness framework" raises the question of whether the name still fits, or whether the rename should accompany the conceptual reframe.

Two forces:
- A rename signals the category change clearly. Users encountering the project today see "installer." A new name (e.g., `harness`, `keystone-framework`) would carry the new framing on the wrapper.
- A rename breaks every existing install. `brew install tacoda/tap/keystone`, the binary on `$PATH`, the curl-bootstrap URL, every blog post and Slack share, every `keystone init` invocation in user history. Import paths for any Go consumers move too.

## Decision

Keep the binary name **`keystone`**. Reposition the project as **"Keystone - the agent harness framework for any project"** in the README, docs, ADRs, and release notes. No CLI rename, no Homebrew tap rename, no Go import-path change.

## Consequences

- Positive: Zero migration burden for existing installs. The 0.x → 1.0 break is contained to layout and config — not to the name on `$PATH`.
- Positive: All documentation surfaces (README, release notes, docs) carry the new framing without code or distribution changes.
- Negative: Short-term confusion as users encounter the framework framing against residual "installer" framing in older posts and external links. Mitigated by a tight README rewrite at 1.0.
- Neutral: The word "keystone" already implies the load-bearing piece of an arch — it generalizes cleanly from "installer" to "framework."

## Alternatives considered

- **Rename to `harness`** — rejected. Generic, conflicts with existing tooling in many ecosystems, breaks every install path.
- **Rename to `keystone-framework`** — rejected. Long, awkward at the shell, would still need a `keystone` alias for compatibility (which defeats the rename).
- **Dual-name (`keystone` + `harness` as alias)** — rejected. Doubles documentation surface and confuses what users should type.
