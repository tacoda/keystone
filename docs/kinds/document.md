# Document

<!-- voice-register: informational -->
<!-- voice-english: ste -->

A document is a template for governed output, such as a plan, a review, or an ADR. The template lives in `.agents/documents/`, and each filled copy lives in `.keystone/work/`. Every copy moves forward through an ordered list of gates, and `keystone document promote` refuses a backward move. The cost is that keystone writes no copy for you: the agent or a person fills it in.

## Example

The template, `.agents/documents/plan.md`:

```markdown
---
kind: document
id: plan
description: The approved plan for one unit of work.
type: plan
produced_by: plan
gates:
  - draft
  - in-review
  - approved
  - done
---

# plan

## Goal
## Steps
## Risks
```

A filled copy, `.keystone/work/PROJ-142/plan.md`:

```markdown
---
kind: document
id: PROJ-142-plan
description: Plan for PROJ-142.
type: plan
gates: [draft, in-review, approved, done]
gate: draft
---
```

When the draft is ready, advance your copy:

```
keystone document promote .keystone/work/PROJ-142/plan.md in-review
```

## Where it lives

```
.agents/documents/<id>.md       # the template
.keystone/work/<task>/<file>.md # each filled copy
```

The template is a charter file, so `keystone index` lists it. Each copy is state, so the index skips it; use `keystone document list` to see your copies.

## Frontmatter

| Key | On | Meaning |
|---|---|---|
| `kind` | both | `document`. Keystone infers it for a template in the `documents/` folder. |
| `id` | both | The template id, or a unique id for the copy. |
| `description` | both | One line. |
| `type` | both | The subtype, such as `plan` or `adr`. |
| `produced_by` | template | The [command](command.md) that writes copies of this document. |
| `gates` | both | The ordered lifecycle. The copy needs its own list, because `promote` reads the copy and not the template. |
| `gate` | copy | Where the copy is now. If it's empty, `promote` treats the first gate as current. |

A command names the documents it writes in `produces:` and the ones it reads in `consumes:`. `keystone lint` warns when either names a document that doesn't exist.

## Commands

| Command | What it does |
|---|---|
| `keystone new document <id>` | Writes a template with the gates `draft`, `in-review`, `approved`, `executed`, and `done`. |
| `keystone document list [dir]` | Prints the id, gate, and path of every copy under `.keystone/work/`. It skips files with no frontmatter. |
| `keystone document promote <path> <gate>` | Moves the copy to a later gate and prints the change. |

`promote` refuses four cases:

- The copy has no `gates:` list.
- Your target gate isn't in the list.
- The current gate isn't in the list.
- The move isn't forward.

A forward move can skip gates, so you can go from `draft` straight to `approved`.

After a promote, keystone fires the `on-gate` signal. Each [hook](hook.md) or [tool](tool.md) with `event: on-gate` then runs, with the copy's `type` in `KEYSTONE_TYPE` and the new gate in `KEYSTONE_COMMAND`. If that hook fails, keystone logs the failure, but the gate has already moved.

## How each host gets it

No host gets a document file. `AGENTS.md` describes the kind and says copies live in `.keystone/work/`, and the agent opens a template when a command tells it to.

## What a document does not do

- It doesn't create copies. The agent writes a copy, as the command that `produces` it says, or you write one by hand.
- It doesn't check the sections of a copy against the template.
- It doesn't move backward. If you need to reopen a copy, edit its `gate:` by hand.
- It doesn't block on a failed `on-gate` hook, because the gate moves first. Don't use that hook as a gate check.

## Failure modes

- **`promote` says the document declares no `gates:`.** You copied the body but not the list. Copy `gates:` from the template into the copy.
- **`promote` says the current gate isn't in the list.** Someone edited `gate:` by hand. Fix the value, then promote again.
- **`list` doesn't show a copy.** The file has no frontmatter, or it sits outside `.keystone/work/`.

## Authoring

```
keystone new document <id>
```

This writes `.agents/documents/<id>.md` with empty `type` and `produced_by` values and a TODO description. Fill in all three, cut the gates you don't need, then run `keystone index`.
