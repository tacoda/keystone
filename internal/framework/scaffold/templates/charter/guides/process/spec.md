---
kind: guide
id: process/spec
description: 'The phase that captures what needs to happen, before deciding how.'
---
# Spec

The phase that captures *what* needs to happen, before deciding *how*. The entry point to every task.

## Entry condition

Some unit of work to anchor on — any of:

- A tracker card identifier (Jira, Linear, GitHub Issues, Asana, etc.) when the project uses a tracker.
- An item in a `TODO.md` or similar in-repo task list.
- An idea or request from the user.

If none of these → ask the user "what do you want to do?" and refuse to proceed without an answer.

A full issue tracker is not required. A small project tracked by a todo list is a valid entry point; the spec phase only needs an identifiable unit of work to write acceptance criteria against.

## Activities

### 1. Source the spec

Two paths:

**A. Tracker card provided** — fetch the card via whatever tracker integration the agent has available (MCP server, CLI, policy). Read:
- Title (becomes the spec title).
- Description (becomes the spec body).
- Acceptance criteria, if present (becomes the AC section).
- Labels and links (kept as cross-references).

If acceptance criteria are missing or weak, prompt the user to author them — to the card *and* to the spec.

**B. No tracker card** — author the spec inline. Ask the user for:
- A one-sentence statement of the intent.
- The acceptance criteria (numbered list, each verifiable).

The card identifier (if any) becomes the **task's reference** — every downstream phase carries it.

### 2. Write explicit acceptance criteria

A spec without testable acceptance criteria is not a spec — it is a wish. Each criterion must be:

- **Observable** — describable as a yes/no check on output, behavior, or artifact.
- **Independent** — does not depend on the resolution of another criterion.
- **Bounded** — clear what is in scope and what is not.

Example:

```
## Acceptance criteria
1. The command prints the matching idiom files within the region.
2. If no idioms match, the report says so explicitly (not silently empty).
3. Exit code 0 on success; non-zero on any failure with a clear message.
```

### 3. Author non-goals and open questions

Ask:

- What is *not* in scope?
- What questions are still open that planning will need to resolve?

These do not block the spec; they document the contour of the work.

### 4. Save the spec

Write to `docs/specs/YYYY-MM-DD-<topic>.md`:

```markdown
---
created_at: <ISO datetime>
tracker_card: <ID or URL, if any>
status: approved
---

# <Title>

## Intent
<One-sentence statement of what to do and why.>

## Acceptance criteria
1. <criterion>
2. <criterion>

## Non-goals
<What is explicitly out of scope.>

## Open questions
<Questions resolved during spec authoring, with answers — or left explicit for planning.>
```

## IRON LAW

**NO PROCEEDING WITHOUT EXPLICIT ACCEPTANCE CRITERIA.**

A task without AC has no verification gate, no review gate, and no done condition. Refuse to leave the spec phase until the AC are written and the user has approved them.

## Sensors

Spec runs:

- **Tracker-card fetcher** — if a card ID is provided, fetches via the agent's tracker integration. Read-only.

The binding (how to invoke the **spec** action and which sensors run) lives in `charter/adapters/<your-agent>/lifecycle.md`.

## Gate condition

An **approved** spec exists in `docs/specs/` with explicit acceptance criteria. "Approved" means the user has either confirmed the AC are correct or fetched them from a card whose contents they own.

If AC cannot be elicited (the user does not know, the card is empty), do not proceed. Surface the gap.

## Artifacts

| Kind | Location |
|---|---|
| Spec | `docs/specs/YYYY-MM-DD-<topic>.md` |

## Anti-patterns

- A spec with acceptance criteria like "make it work" or "it should be good".
- Proceeding to planning without acceptance criteria.
- Fetching a tracker card and accepting whatever it says without confirming the AC are testable.
- Skipping the spec phase because "the change is small". A small change still has a measurable outcome.
