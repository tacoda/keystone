---
name: keystone:spec
description: Enter the spec phase — capture intent and acceptance criteria
argument-hint: "[<tracker-card-id-or-url>]"
---

You are entering the **spec** phase of the project's keystone harness. Read `harness/guides/process/spec.md` and follow its activities.

If `$1` is provided, treat it as a tracker card identifier and fetch it (Atlassian / Linear / GitHub Issues / Asana MCP server, or `gh issue view`). Copy any acceptance criteria into the spec.

If `$1` is omitted, author the spec inline: ask the user for intent and acceptance criteria.

## Activities

1. **Restate the intent** in your own words.
2. **List acceptance criteria** — explicit, testable bullets a reviewer can verify objectively.
3. **List non-goals** — what is not in scope.
4. **Flag uncertainty** — open questions the user must answer before planning.
5. **Save the spec** to `docs/specs/YYYY-MM-DD-<topic>.md` with the frontmatter shape described in `harness/guides/process/spec.md`.

## Gate

Do not proceed to planning until the user has explicitly accepted the spec.

## Iron law

**No proceeding without explicit acceptance criteria.** A spec that says "make X better" is not yet a spec.
