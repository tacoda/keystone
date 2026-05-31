# Learning

The staging area for the Learning flywheel. Novel reasoning the agent surfaces during work gets parked here, reviewed by a human, and then either promoted into the corpus or rejected with reasoning.

## Layout

- `inbox/` — raw candidates. The agent writes here when it makes a judgment call not covered by the corpus.
- `promoted/` — accepted items, after the **synthesize** action has folded them into a layer. Kept as a record.
- `rejected/` — reviewed and not promoted. Kept with the reason.

## The flow

1. Agent encounters a gap → writes a candidate to `inbox/<timestamp>-<slug>.md`.
2. Human reviews via the **synthesize** action.
3. **synthesize** either:
   - **Promotes** the candidate into the right corpus layer (`principles/` / `idioms/` / `domain/` / `process/`), updates the relevant file, and moves the inbox item to `promoted/`.
   - **Rejects** it, records the reason, and moves it to `rejected/`.
4. On any promotion, **synthesize** ends with a **reload prompt**. The new rules took effect on disk; they are not yet in the active session's ambient context. Reset the agent's context (see `harness/adapters/<your-agent>/activation.md`) and re-prompt to pick them up.

## Inbox item format

```markdown
---
captured_at: <ISO datetime>
captured_by: <command / agent>
confidence: <high | medium | low>
suggested_layer: <principles | idioms | domain | process>
---

# <One-line title>

## What happened
The situation. What the agent encountered that was not covered by the corpus.

## The judgment call
What the agent decided, and why.

## Suggested rule
The pattern, if any, that should become a corpus entry.
```

## Authorship

Agent writes the inbox. Humans gate. **synthesize** is the promotion ritual.

## Activation

Not ambient. The inbox is operational — it accumulates between **synthesize** runs. The agent does not read its own inbox during normal work.

## Changes when

Every time the **learn** action fires (post-commit) and the agent has something to record. Often. The inbox is meant to be a stream.
