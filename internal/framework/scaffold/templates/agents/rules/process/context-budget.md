---
kind: guide
id: process/context-budget
description: 'How much the agent reads, runs, or spawns before making a change.'
---
# Context Budget

How much the agent reads, runs, or spawns before making a change. Counterpart to [[scoping]] (the *output* size limit); this is the *input* size limit.

## RULES

- **Read what is relevant to the touched region.** The orient action lazy-loads by region precisely so the agent does not load the whole codebase. Honor it.
- **Grep before reading.** A single `grep` that returns three files is cheaper than reading thirty files looking for the same information.
- **Stop reading when the question is answered.** Reading "to be thorough" past the point of confidence wastes context.
- **Bound exploratory reads to a budget.** Default: ≤10 files for orientation, ≤5 files for a follow-up. Bootstrap may tune.
- **Use sensors before reading.** The state-region sensor exists so the agent does not re-derive what is already recorded.

## GOLDEN RULE

- **Aim to defer subagent spawning until the problem is well-shaped.** A subagent is most useful when the parent has a clear question; spawning to "explore" usually returns a worse summary than a direct read. See [[subagent-trust]].
- **Aim to compact long sessions at phase boundaries.** A 90-minute session that crosses phases has earned a context reset. See `planning.md` session hygiene.
- **Aim to leave the inbox the way you found it.** Reading every previous inbox entry "just in case" is rarely worth it — synthesize handles that.

## Why this is agent-specific

A coding agent has a finite context window and a finite attention budget within it. Reading too much degrades the quality of the response: the relevant information competes with the irrelevant. Reading too little produces hallucinated grounding — see [[grounding]]. The discipline is to read *enough and no more* — which requires explicit budgets, because the agent's natural pull is toward "let me check one more file."

The cost of getting this wrong is mostly invisible: the user sees a slower response, a larger token bill, a vaguer answer. The corrective is structural — budgets, grep-first, lazy-by-region.

## Sensors

- The **state-region sensor** runs at the start of planning so the agent reads recorded knowledge before deriving it.

## Anti-patterns

- Opening the same file three times in one session.
- Reading the full git log when the last three commits would do.
- Spawning a subagent to "look around" without a specific question.
- Running `find` from the repo root with no pattern.
- Compacting mid-thought and losing the active question.

---

See also: [[scoping]], [[grounding]], [[subagent-trust]], `planning.md` (session hygiene), [`modes.md`](modes.md).
