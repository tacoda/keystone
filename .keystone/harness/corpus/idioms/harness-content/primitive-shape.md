---
kind: corpus
id: corpus/idioms/harness-content/primitive-shape
description: 'Frontmatter, paths, and corpus/guide pairing for harness primitives.'
---
# Primitive shape

Every harness primitive is one markdown file (or a SKILL.md inside a slug folder for skills). The agent reads the frontmatter to decide *what kind* of primitive it is and *when* to activate it; the body holds the content. Shape mistakes silently downgrade a primitive to "noise the agent skips."

> **Rules extracted:** [`guides/idioms/harness-content/primitive-shape.md`](../../../guides/idioms/harness-content/primitive-shape.md).

## Canonical paths

| Kind | Path |
|---|---|
| action | `harness/actions/<id>.md` |
| playbook | `harness/playbooks/<id>.md` |
| guide | `harness/guides/<topic>/<id>.md` |
| corpus | `harness/corpus/<topic>/<id>.md` |
| sensor | `harness/sensors/<id>.md` |
| persona | `harness/personas/<id>.md` |
| skill | `harness/skills/<slug>/SKILL.md` |
| subagent | `harness/agents/<id>.md` |
| command | `harness/commands/<id>.md` |
| rule | `harness/rules/<id>.md` |
| computational | `harness/guides/computational/<tool>.md` |
| adapter | `harness/adapters/<agent>/{lifecycle,sensors,activation}.md` |

Off-path content is invisible to `keystone index` and never lands in `INDEX.json`.

## Frontmatter

Required on every primitive:

```yaml
---
kind: <one of the twelve kinds>
id: <stable id used in INDEX.json>
description: 'One sentence. Used by the agent to decide whether to open the body.'
---
```

Optional, kind-specific:

- `globs:` — list of fileset patterns. **Narrow-only**: globs can only narrow ambient activation, never broaden it.
- `severity:` — `must` / `should`. Defaults to `should`. Used on guides where deviation triggers a sensor-grade complaint.
- `triggers:` — list of strings the host matches against user input (skills only).
- `tools:` — tool allowlist for the persona's subagent (personas only).
- `provenance:` — set automatically by `keystone index` to `project` or `plugin`; do not hand-edit.

## Pair convention (corpus ↔ guide)

A guide carries rules. The matching corpus carries the *why*. They share a topic + name:

- `guides/idioms/go/stdlib-first.md` (rules)
- `corpus/idioms/go/stdlib-first.md` (reasoning)

The guide links to the corpus in its body; the corpus links back. Either side existing without the other is a structural smell flagged by the **harness-debt** sensor.

## Narrow-only globs

A guide with `globs:` activates only when:

1. Its ambient activation rule says yes (e.g., idiom guide for a stack region), AND
2. At least one touched file matches a glob.

A guide without `globs:` activates per its topic default (ambient). Globs can never *force* a guide to activate outside its topic. This is the contract pointer-style adapters (Claude Code, Codex, Aider) rely on to keep loaded-rule sets small.

## Anti-patterns

- A primitive at an off-canonical path (e.g., `harness/notes/foo.md`). `keystone index` skips it.
- A guide with `globs:` it copy-pasted from another guide and never narrowed.
- A corpus entry without a guide partner (or vice versa) — orphan reasoning / orphan rules.
- Hand-edits to `INDEX.json` or `GLOBS_INDEX.md` — both are regenerated; the edit is lost on next `keystone index` / `synthesize`.
- A `description:` written as a placeholder (`'TODO'`, `'description'`). The agent uses it to gate body loading; a useless description gets the primitive ignored.

## Review checklist

- Frontmatter present? `kind`, `id`, `description` all real?
- File at the canonical path for its `kind`?
- If a guide declares `globs:`, do the patterns reference paths that exist?
- Paired corpus / guide both present?
- `keystone index` re-run after the change?

**Traces to:** the keystone framework's primitive-resolution contract (see `corpus/process/runtime-resolution.md` and `harness/README.md`).
