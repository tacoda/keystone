---
name: keystone-spec
description: 'Capture intent + acceptance criteria for a unit of work. First action on any task. Triggers: keystone spec; keystone:spec; /keystone:spec; capture acceptance criteria; spec this task; what are we building.'
---

# keystone:spec — capture intent + AC

Phase 1 of the six-phase lifecycle. Forces the agent to write down what
"done" looks like before any code is touched. Output lands in the task
notes / PR description / equivalent — acceptance criteria the agent and
user both sign off on.

Full discipline
at `.agents/rules/process/spec.md`.

## Run

**Capture intent and acceptance criteria** for a unit of work. First command on any task. Read [`.agents/rules/process/spec.md`](../../rules/process/spec.md) for the full discipline.

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
5. **Save the spec** to `docs/specs/YYYY-MM-DD-<topic>.md` with the frontmatter shape described in `.agents/rules/process/spec.md` (if the project has that location; otherwise capture inline in the chat).

## Gate

Do not proceed to planning until the user has explicitly accepted the spec.

## Iron law

**No proceeding without explicit acceptance criteria.** A spec that says "make X better" is not yet a spec.

## When to trigger

- Start of every task. Non-negotiable on anything beyond a typo fix.
- When mid-task scope drift makes the original intent stale — re-spec.
