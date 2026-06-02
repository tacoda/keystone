---
description: Enter the spec phase — capture intent and acceptance criteria
argument-hint: "[<tracker-card-id-or-url>]"
---

Enter the **spec** phase of the project harness.

Read `harness/guides/process/spec.md` and follow its activities exactly. The IRON LAW applies: no proceeding without explicit, testable acceptance criteria.

If `$1` is provided, treat it as a tracker card identifier. Fetch the card via whatever tracker tool you have access to (`gh issue view`, `jira` CLI, an MCP server, etc.). If the card has acceptance criteria, copy them into the spec. If it does not, prompt the user to author them.

If no `$1` is provided, author the spec inline by asking the user for intent and acceptance criteria.

Save the spec to `docs/specs/YYYY-MM-DD-<topic>.md` with the frontmatter shape described in `harness/guides/process/spec.md`.
