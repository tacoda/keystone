# 0005 — Conventions, not plugins

**Status:** Accepted
**Date:** 2026-06-08

## Context

An earlier draft of the 1.0 plan made the built-in defaults — universal guides and corpus, the lifecycle playbook and its actions, default sensors, per-agent adapters — into "first-class policy plugins" loaded through the same engine as user-installed plugins. The promise was symmetry: defaults and external policy travel the same pipeline.

That coupling has a cost. It bundles the *share-policy-across-projects* mechanism (external plugins) with the *ship-defaults* mechanism (built-ins). Editing a default stops being "edit a markdown file." It becomes "fork the embedded plugin or override it from the project layer." Rails-like ergonomics for the defaults — drop in, edit, commit — disappears.

The reframe at 1.0: defaults are project content from the moment `init` finishes. Plugins are only for sharing policy across projects.

## Decision

- Built-in defaults (universal guides/corpus, lifecycle playbook + actions, default sensors, per-agent adapters) are **scaffolded** into the consumer's `harness/<port>/` on `keystone init` from embedded templates at `internal/framework/scaffold/templates/`.
- After `init`, defaults are ordinary project content. The user edits them in markdown, commits them, and versions them with the project's git tags.
- Plugins exist **only** to share policies across projects. A plugin is an external, read-only dependency.
- The framework has no "embedded plugin" concept and no `plugins/` directory inside the keystone repo.

## Consequences

- Positive: "Editing the harness" = "edit a markdown file." Rails-like.
- Positive: The plugin mechanism does one thing — vendored, hash-verified, drift-reset, read-only. See [0006](0006-vendored-readonly-plugins.md).
- Positive: The framework binary carries templates and a scaffolder, not a default-plugin registry.
- Negative: A user's scaffolded copy can drift from upstream defaults across framework versions. Mitigated by `keystone doctor`'s template-diff report (Phase 4).
- Negative: Users on 0.x who relied on the "override the default plugin" pattern re-port by hand. Acceptable because the 0.x → 1.0 path is a documented `init --reset`. See [0007](0007-no-backward-compat-at-1.0.md).

## Alternatives considered

- **Built-ins as first-class embedded plugins** (prior draft) — rejected. See Context.
- **Built-ins streamed from the binary on every load** — rejected. Treats defaults as binary state; the editing UX becomes "edit-and-pray-the-loader-prefers-yours."
- **Built-ins scaffolded but tracked for auto-update** — rejected. The framework rewriting user-owned files crosses the boundary in [0002](0002-framework-client-boundary.md). Updates are opt-in via doctor.
