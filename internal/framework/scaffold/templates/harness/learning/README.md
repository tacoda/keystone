# Learning

The staging area for the Learning flywheel. Novel patterns the agent surfaces during work get parked here, reviewed by a human, and then either promoted into a guide, a corpus file, or both — or rejected with reasoning.

## Layout

- `inbox/` — raw candidates. The agent writes here when it makes a judgment call not covered by the existing harness.
- `promoted/` — accepted items, after the **synthesize** action has folded them into a guide and/or corpus file. Kept as a record.
- `rejected/` — reviewed and not promoted. Kept with the reason.
- `wishlist.md` — known gaps the team plans to address eventually. Team-curated, not agent-generated; items promote to `inbox/` when a real situation triggers them.

## The flow

1. Agent encounters a gap → writes a candidate to `inbox/<timestamp>-<slug>.md`.
2. Human reviews via the **synthesize** action.
3. **synthesize** **classifies** the candidate as rule or information:
   - **Rule.** A directive the agent must follow next time. Lands in the right `guides/<layer>/<name>.md`. May also update or create a paired corpus file with the *why* (especially for non-obvious rules). Rules have three tiers — see below.
   - **Information.** Supplemental context — an ideal, design rationale, history, an anti-pattern catalog, a citation. Lands in the right `corpus/<layer>/<name>.md`. No guide change.
4. After landing, **synthesize** moves the inbox item to `promoted/`.
5. If anything landed in `guides/`, **synthesize** ends with a **reload prompt** (guides are ambient — the active session needs a context reset to pick them up). If nothing landed in `guides/`, no reload is needed; corpus is on-demand.
6. Rejection: **synthesize** records the reason and moves the candidate to `rejected/`.

## How to classify

Ask, for each candidate:

| Question | If "yes" → |
|---|---|
| Is this a directive the agent must follow when it sees a matching situation? | **Rule** — lands in `guides/` |
| Is this background the agent only needs when reasoning about *why* something is the way it is? | **Information** — lands in `corpus/` |
| Both? | Update both files. The guide carries the directive; the corpus carries the rationale. |

Default to **information** when in doubt. Adding a rule narrows the agent's behavior across the whole project; the bar should be higher than adding context.

## Rule tiers

A rule promoted to `guides/` lands in one of three tiers. **Default to regular RULES.** IRON LAW and GOLDEN RULE are opt-in — synthesize may *suggest* either when the candidate warrants it, but the user must confirm before promotion lands there.

| Tier | Strength | When to use | Drift sensor |
|---|---|---|---|
| **RULES** (regular) | Standard directive. The default tier. | Most rules. A normal "do this / don't do that" without ceremonial weight. | Warn |
| **GOLDEN RULE** | Strong, explicit standard. Deviation requires reasoning. | An explicit standard the team holds itself to — concrete ("inject dependencies via the constructor; do not new them up inside other classes") or aspirational ("controllers should be thin; delegate to services"). | Warn (strongly) |
| **IRON LAW** | Non-negotiable. | Violation causes real damage — incidents, security exposure, lost work. Extremely rare by design. | Fail |

The special tiers derive their force from being rare. If every rule gets promoted to IRON LAW or GOLDEN RULE, the labels stop signaling anything. When in doubt, land at **RULES** — promote later if a pattern of recurring violations justifies it.

### Synthesize prompt for tier

When synthesize is ready to land a rule, it confirms the tier with the user:

1. Show the rule text and the proposed `guides/<layer>/<name>.md` target.
2. Default tier: **RULES**. If the candidate's content reads as non-negotiable or as an explicit aspiration, synthesize may say "this looks like it could be a GOLDEN RULE / IRON LAW — promote to that tier?" — but the user decides.
3. Land under the confirmed heading (`## RULES`, `## GOLDEN PATH`, or `## IRON LAW(S)`) inside the target file.

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
The rule this becomes. Phrase as a directive. Tier (regular RULES, GOLDEN RULE, or IRON LAW) is decided at **synthesize**, not here — the inbox just captures the directive.

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
