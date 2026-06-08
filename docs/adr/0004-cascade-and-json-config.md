# 0004 — Cascade and JSON config

**Status:** Accepted
**Date:** 2026-06-08

## Context

0.x layers policy with a flat tier enum (Org → Team → Project) and stores configuration in YAML (`keystone-policy.yaml`, `.keystone.lock`). The tier enum prevents legitimate setups (an org wants to ship a platform-specific layer on top of its org-wide rules) and YAML invites two-format drift if any companion format gets added later. For 1.0 we want arbitrary-depth policy layering and a single, tooling-friendly config format.

## Decision

**Cascade.**

- The plugin cascade is declared in `keystone.json` as a **nested tree** under a top-level `plugins[]` array. Each node has `name`, `source`, `version`, optional `strict`, optional `children[]`.
- Resolution is a **pre-order walk** of `plugins[]`. Earlier nodes win against later nodes for the same `<port>/<name>`.
- Project files (`harness/<port>/...`) sit at the front of the priority list, always.
- `strict` on a node blocks deeper nodes (its descendants and later-tree siblings deeper in the tree) from shipping the named item. `keystone verify` reports violations at install time.

**Config format.**

- All config is **JSON only**: `keystone.json`, `keystone.lock.json`, plugin manifests (`keystone-plugin.json`), framework migrations.
- JSON Schemas for each format live under `docs/schemas/`.
- Markdown remains the format for content. JSON is the format for config.

## Consequences

- Positive: Arbitrary policy layering. Orgs can ship a platform plugin nested under an org plugin nested under a team plugin without forking.
- Positive: The tier enum disappears from runtime code paths.
- Positive: Pre-order precedence is readable from `keystone.json` alone — no implicit ordering rules.
- Positive: One config format means uniform editor support, schema validation, jq pipelines, and generator output.
- Negative: Authors lose YAML's inline-comment convenience. Mitigated by tight JSON Schemas, descriptive field names, and `keystone doctor` reports.
- Negative: The existing 0.x lockfile and policy YAML must be replaced (Phase 1 sweep). `gopkg.in/yaml.v3` drops from `go.mod`.

## Alternatives considered

- **YAML for human-edited files, JSON for machine files** — rejected. Two-format drift is the worst of both worlds.
- **Flat `plugins` list with explicit `parent` field** — rejected. Reading order becomes computed; nested arrays make precedence grep-friendly.
- **TOML** — rejected. JSON has a stronger schema and `jq` ecosystem. Most edits to `keystone.json` happen via `keystone plugin add/update/remove`, weakening the human-editing argument.
