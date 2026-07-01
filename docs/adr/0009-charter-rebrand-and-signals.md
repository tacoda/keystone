# 0009 — Charter rebrand + signals (4.0)

**Status:** Accepted
**Date:** 2026-07-01

This ADR is also the first-class **amendment record** for the 4.0 charter
changes (see `GOVERNANCE.md` → Charter amendments).

## Context

Keystone was branded "the agent harness framework," and its own artifact — the
tree of primitives — was called the "harness," living at `.harness/`. But a
*harness* is the engine that runs the model (Claude Code, the orchestrator, the
runner). Keystone doesn't run the model; it authors the standards that
constrain whatever harness does. Calling both "harness" conflated the thing
being governed with the thing doing the governing.

Separately, the `hook` primitive was a wiring middleman (event → run/agent) and
its `event:` set was a closed hard-coded list, so projects couldn't define
their own framework events, and "hook" collided with the mainstay host-hook
term.

## Decision

**Rebrand harness → charter.** Keystone is the *coding-agent charter manager*.
The artifact is a **charter** at `.charter/`; `HARNESS.md` → `CHARTER.md`. A
harness is the engine; a charter is the authored spec that constrains it
(authorship test — see `GLOSSARY.md`). "harness" survives in copy only where it
means the engine.

**Signals replace the closed framework-event set.** A **signal** is a keystone
framework event; host phases are a closed set, and *any other* `on:` value is a
signal — so projects define their own (`keystone.json signals:`, `keystone
signal fire|list`). The classifier is inverted: host phases closed, signals
open.

**Retire the `hook` kind.** Reactions self-subscribe via `on:` (like a skill
declares `triggers:`):
- `sensor` — a check → verdict (exit/HTTP status); gates. Inferential sensors
  return structured `returns:`.
- `tool` — an external callable (transport cli|http|mcp|plugin); on-demand, or
  a side-effect with `on:`. An MCP tool is the `mcp` transport — "tool" is not
  overloaded.
- `agent` — an inferential review.

**Operator surfaces.** `keystone charter coverage` (uncharted territory) and
`keystone charter show --effective` (post-cascade roster); the dashboard and
MCP expose signals + coverage.

## Consequences

- **Breaking.** `.harness/` → `.charter/` and `hooks/` → `sensors/` are on-disk
  layout changes; the `v4.0` migration performs both (frozen literal paths,
  idempotent). Ships as 4.0.0.
- Community-standard kind names are otherwise kept (guide/sensor/command/skill/
  agent/…) — a vocabulary rename beyond `hook` was considered and rejected as
  churn without payoff.
- `event:` is retained as a back-compat alias for `on:`.

## Provenance

Proposed and ratified by the maintainer (solo stage). Superseded terms
("harness" for the artifact, the `hook` kind) are retired, not deprecated —
4.0 is a clean break with a migration.
