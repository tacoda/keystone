# Learning

The staging area for the Learning flywheel. Novel patterns the agent surfaces during work get parked here, reviewed by a human, and then either promoted into a guide, a corpus file, or both — or rejected with reasoning.

## Layout

- `inbox/` — raw candidates. The agent writes here when it makes a judgment call not covered by the existing harness.
- `promoted/` — accepted items, after the **synthesize** action has folded them into a guide and/or corpus file. Kept as a record.
- `rejected/` — reviewed and not promoted. Kept with the reason.

## The flow

1. Agent encounters a gap → writes a candidate to `inbox/<timestamp>-<slug>.md`.
2. Human reviews via the **synthesize** action.
3. **synthesize** **classifies** the candidate as rule or information:
   - **Rule.** A constraint the agent must follow next time — an IRON LAW or GOLDEN RULE. Lands in the right `guides/<layer>/<name>.md`. May also update or create a paired corpus file with the *why* (especially for non-obvious rules).
   - **Information.** Supplemental context — an ideal, design rationale, history, an anti-pattern catalog, a citation. Lands in the right `corpus/<layer>/<name>.md`. No guide change.
4. After landing, **synthesize** moves the inbox item to `promoted/`.
5. If anything landed in `guides/`, **synthesize** ends with a **reload prompt** (guides are ambient — the active session needs a context reset to pick them up). If nothing landed in `guides/`, no reload is needed; corpus is on-demand.
6. Rejection: **synthesize** records the reason and moves the candidate to `rejected/`.

## How to classify

Ask, for each candidate:

| Question | If "yes" → |
|---|---|
| Is this a constraint the agent must follow when it sees a matching situation? | **Rule** — lands in `guides/` |
| Is this background the agent only needs when reasoning about *why* something is the way it is? | **Information** — lands in `corpus/` |
| Both? | Update both files. The guide carries the constraint; the corpus carries the rationale. |

Default to **information** when in doubt. Adding a rule narrows the agent's behavior across the whole project; the bar should be higher than adding context.

## Inbox item format

```markdown
---
captured_at: <ISO datetime>
captured_by: <command / agent>
confidence: <high | medium | low>
candidate_kind: <rule | information | both>
suggested_target: <guides/principles/tdd.md | corpus/idioms/mvc/controllers.md | ...>
---

# <One-line title>

## What happened
The situation. What the agent encountered that was not covered by the harness.

## The judgment call
What the agent decided, and why.

## Suggested rule (if any)
The IRON LAW or GOLDEN RULE this becomes. Phrase as a directive.

## Suggested information (if any)
The reasoning, ideal, or context this becomes. Phrase as explanation.
```

The agent fills `candidate_kind` and `suggested_target` as a hint. **synthesize** confirms or overrides.

## Authorship

Agent writes the inbox. Humans gate. **synthesize** is the promotion ritual — and the only place where the rule-vs-information classification is finalized.

## Activation

Not ambient. The inbox is operational — it accumulates between **synthesize** runs. The agent does not read its own inbox during normal work.

## Changes when

Every time the **learn** action fires (post-commit) and the agent has something to record. Often. The inbox is meant to be a stream.
