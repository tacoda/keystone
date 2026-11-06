# Learning

The staging area for the learning flywheel. When the agent makes a judgment call the charter doesn't cover, it parks a candidate here. A person reviews it. Then `synthesize` promotes it into a rule, a reference, or both, or rejects it with a reason.

## Layout

- `inbox/`: raw candidates. The `learn` command writes here.
- `promoted/`: candidates `synthesize` folded into the charter, kept as a record.
- `rejected/`: candidates reviewed and not promoted, kept with the reason.
- `wishlist.md`: known gaps the team plans to address. People curate it, not the agent.

## The flow

1. The agent meets a gap and writes `inbox/<timestamp>-<slug>.md`.
2. A person runs `synthesize` to review the inbox.
3. `synthesize` classifies each candidate as a rule, information, or both.
4. A rule lands in `.agents/rules/<topic>/<name>.md`. Information lands in `.agents/references/<topic>/<name>.md`.
5. The candidate moves to `promoted/` or `rejected/`.
6. After a promotion, run `keystone index && keystone project`. A running session may still hold the old rule set, so start a new one.

## Rule or information?

| Question | If yes |
|---|---|
| Must the agent follow this when it sees a matching situation? | A rule |
| Does the agent need this only to reason about why? | A reference |
| Both? | Write both. The rule names the reference under `corpus:`. |

When in doubt, choose a reference. A new rule narrows the agent everywhere it applies, so its bar is higher.

## Rule tiers

New rules land under `## Rules`. `synthesize` may suggest a golden rule or an iron law, but a person confirms before a rule lands in either tier.

| Tier | Strength |
|---|---|
| Rules | The default. A plain directive. |
| Golden rule | A strong standard. Deviation needs a stated reason. |
| Iron law | Non-negotiable. Violation causes real damage. Rare by design. |

## Inbox item format

```markdown
---
captured_at: <ISO datetime>
captured_by: <command or agent>
confidence: <high | medium | low>
candidate_kind: <rule | information | both>
suggested_target: <.agents/rules/... or .agents/references/...>
---

# <One-line title>

## What happened
The situation the charter didn't cover.

## The judgment call
What the agent decided, and why.

## Suggested rule (if any)
A directive. `synthesize` decides the tier.

## Suggested information (if any)
The reasoning or context, phrased as explanation.
```

The agent fills `candidate_kind` and `suggested_target` as a hint. `synthesize` confirms or overrides them. The agent doesn't read its own inbox during normal work.
