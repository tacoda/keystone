# 0003 — Ports and adapters

**Status:** Accepted
**Date:** 2026-06-08

## Context

The framework's surface is a set of markdown-based abstractions: guides, corpus, sensors, actions, playbooks, per-agent adapters, plus the flywheel sinks. Today they are partly documented and partly implicit. To make 1.0 a contract, we need to enumerate them, freeze the names, and write a one-page contract per abstraction.

## Decision

The 1.0 ports are:

| Port | Adapter location |
|---|---|
| **Guide** | `<port>/<topic>/<name>.md` |
| **Corpus** | `<port>/<topic>/<name>.md` |
| **Sensor** | `<port>/<name>.md` |
| **Action** | `<port>/<name>.md` |
| **Playbook** | `<port>/<name>.md` |
| **Adapter (agent)** | `adapters/<agent>/{lifecycle,sensors,activation}.md` |
| **Flywheel sink** | `learning/`, `archive/` (write-only) |

- Each port has a one-page contract at `docs/ports/<port>.md` documenting path convention, required shape, frontmatter, cascade behavior, drift-sensor pairing, an example, and the authoring command.
- Concrete files at `harness/<port>/...` (project) or `harness/plugins/<plugin>/<port>/...` (plugin) are *adapters* for that port.
- **Adding a new port** = minor version bump + a new `docs/ports/<port>.md` PR + loader support. No quiet additions.
- **Adding an adapter** = drop a markdown file at the conventional path. No code change.

## Consequences

- Positive: A finite, named API. Tooling (`keystone doctor`, generators, `verify`) enumerates the surface from `docs/ports/` rather than scanning runtime code.
- Positive: Plugin authors target named ports with documented contracts, not implicit conventions.
- Positive: The port list bounds the framework. Anything outside it doesn't exist as far as the runtime is concerned.
- Negative: New abstractions take a minor release. Acceptable — the framework stays small.
- Neutral: The 1.0 port list mirrors the 0.x abstractions almost 1:1, so this is naming and contract polish, not invention.

## Alternatives considered

- **No port enumeration, free-form markdown directories** — rejected. Without a finite map, generators and doctor have nothing to check.
- **Merging actions and playbooks** — rejected. The atomic-unit-vs-ordered-chain distinction carries weight in the lifecycle; collapsing them muddies it.
- **Hardcoding the port list in Go only** — rejected. The contract is the markdown at `docs/ports/`; the Go is an implementation of the contract.
