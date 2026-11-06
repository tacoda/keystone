---
kind: reference
id: corpus/idioms/charter-content/primitive-shape
description: 'Frontmatter, paths, and reference/rule pairing for charter primitives.'
---
# Primitive shape

Every charter primitive is one markdown file (or a SKILL.md inside a slug folder for skills). The agent reads the frontmatter to decide *what kind* of primitive it is and *when* to activate it; the body holds the content. Shape mistakes silently downgrade a primitive to "noise the agent skips."

> **Rules extracted:** [`rules/idioms/charter-content/primitive-shape.md`](../../../rules/idioms/charter-content/primitive-shape.md).

## Canonical paths

| Kind | Path |
|---|---|
| rule | `.agents/rules/<topic>/<name>.md` |
| reference | `.agents/references/<topic>/<name>.md` |
| hook | `.agents/hooks/<id>.md` |
| permission | `.agents/permissions/<id>.md` |
| eval | `.agents/evals/<id>/` |
| skill | `.agents/skills/<name>/SKILL.md` |
| command | `.agents/commands/<id>.md` |
| subagent | `.agents/agents/<id>.md` |
| tool | `.agents/tools/<id>.md` |
| document | `.agents/documents/<id>.md` |

Off-path content is invisible to `keystone index` and never lands in `INDEX.json`.

## Frontmatter

Required on every primitive. Keystone infers `kind` from the folder, so `kind:` is optional:

```yaml
---
id: <stable id used in INDEX.json>
description: 'One sentence. Used by the agent to decide whether to open the body.'
---
```

Optional, kind-specific:

- `paths:`: list of fileset patterns on a rule. A rule with `paths:` applies only when a touched file matches. The 4.x `globs:` still parses.
- `severity:`: `must` / `should`. Defaults to `should`. Used on rules where deviation triggers a hook-grade complaint.
- `triggers:`: a 4.x skill key. Hosts match a skill on its `description`, so move trigger phrases there.
- `tools:`: tool allowlist for a subagent (subagents only).
- `provenance:`: set automatically by `keystone index` to `project` or `policy`; do not hand-edit.

## Pair convention (reference ↔ rule)

A rule carries the directives. The matching reference carries the *why*. They share a topic + name:

- `rules/idioms/go/stdlib-first.md` (directives)
- `references/idioms/go/stdlib-first.md` (reasoning)

The rule links to the reference in its body; the reference links back. Either side existing without the other is a structural smell flagged by the **charter-debt** subagent.

## Paths scoping

A rule with `paths:` applies only when at least one touched file matches a pattern. A rule without `paths:` applies everywhere. `paths:` can only narrow where a rule applies, never widen it. Hosts with path-scoped rule files (Claude Code, Cursor, Copilot, Continue, Cline) load the rule through a shim, which keeps loaded-rule sets small.

## Anti-patterns

- A primitive at an off-canonical path (e.g., `.agents/notes/foo.md`). `keystone index` skips it.
- A rule with `paths:` it copy-pasted from another rule and never narrowed.
- A reference without a rule partner (or vice versa): orphan reasoning / orphan rules.
- Hand-edits to `INDEX.json` or `GLOBS_INDEX.md` — both are regenerated; the edit is lost on next `keystone index` / `synthesize`.
- A `description:` written as a placeholder (`'TODO'`, `'description'`). The agent uses it to gate body loading; a useless description gets the primitive ignored.

## Review checklist

- Frontmatter present? `kind`, `id`, `description` all real?
- File at the canonical path for its `kind`?
- If a rule declares `paths:`, do the patterns reference paths that exist?
- Paired reference / rule both present?
- `keystone index` re-run after the change?

**Traces to:** the keystone framework's primitive-resolution contract (see `references/process/runtime-resolution.md` and `.agents/README.md`).
