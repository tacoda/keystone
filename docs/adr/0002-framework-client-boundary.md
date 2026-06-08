# 0002 — Framework / client boundary

**Status:** Accepted
**Date:** 2026-06-08

## Context

In 0.x, the framework code and the default content blur together. Go sources sit at the keystone repo root next to `harness/`, the markdown that the binary embeds and ships into consumer projects. There is no physical signal for "this is framework behavior" versus "this is content the framework happens to bundle." For 1.0 we want a clean line: code that defines the runtime on one side, conventional content on the other.

## Decision

- Framework code lives under `internal/framework/` (runtime, loader, manifest, lockfile, plugin vendoring, migrate, budget, adapters, scaffold) and `cmd/keystone/` (CLI entrypoint) inside the keystone repo.
- Default content lives as templates at `internal/framework/scaffold/templates/`, embedded via `go:embed`, and is **copied** into the consumer's `harness/<port>/` on `keystone init`.
- There is no `plugins/` directory inside the keystone repo. There is no "embedded plugin" intermediate concept.
- Consumer-side, everything under `harness/` is the user's, with one exception: `harness/plugins/` is read-only vendored content managed by the plugin flow.

## Consequences

- Positive: Clear ownership. A framework PR touches `internal/framework/` and templates; a consumer's edits live in their own `harness/`.
- Positive: One physical signal — anything outside `internal/framework/` and `cmd/keystone/` is content or docs, not behavior.
- Positive: Refactoring the loader or migration runner cannot accidentally rewrite shipped markdown.
- Negative: Moving 0.x's root-level Go files into `internal/framework/` is a multi-PR refactor (Phase 1).
- Neutral: External plugin authors live in their own repos; the framework repo never hosts third-party policies.

## Alternatives considered

- **Embedded default plugins** (earlier draft) — rejected. Coupled the "ship defaults" mechanism to the "share policy" mechanism. See [0005](0005-conventions-not-plugins.md).
- **Documentation-only boundary** — rejected. Layout drifts when only documented; directory structure doesn't.
