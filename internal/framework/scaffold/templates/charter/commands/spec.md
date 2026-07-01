---
kind: command
id: spec
description: 'Capture intent and acceptance criteria for a unit of work.'
---
# spec

**Capture intent and acceptance criteria** for a unit of work. First action on any task. Read [`charter/guides/process/spec.md`](guides/process/spec.md) for the full discipline.

## Inputs

- A tracker card ID (Jira / Linear / GitHub Issues / Asana), **or**
- A conversation describing the work, **or**
- A `TODO.md` entry.

If a tracker card was named, fetch its description via whatever tooling the agent has (MCP server, `gh issue view`, `jira` CLI, web fetch). Copy any acceptance criteria into the spec.

## Activities

1. **Restate the intent** in your own words.
2. **List acceptance criteria** — explicit, testable bullets a reviewer can verify objectively.
3. **List non-goals** — what is *not* in scope, to prevent feature creep.
4. **Flag uncertainty** — open questions the user must answer before planning.
5. **Save the spec** to `docs/specs/YYYY-MM-DD-<topic>.md` with the frontmatter shape described in `charter/guides/process/spec.md` (if the project has that location; otherwise capture inline in the chat).

## Gate

Do not proceed to planning until the user has explicitly accepted the spec.

## Iron law

**No proceeding without explicit acceptance criteria.** A spec that says "make X better" is not yet a spec.
