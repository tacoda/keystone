# Corpus

Informational reference — what the agent should *know* when the rules aren't enough. The reasoning, the literature, the anti-patterns, the lived state of the codebase.

**Loaded on demand, not ambient.** The agent reaches a corpus file when it follows the forward-link from a guide, or when process explicitly references one. Rules live in [`../guides/`](../guides/README.md) and are always loaded; corpus is read only when the agent needs the *why* behind a rule, the history behind an ideal, or the anti-patterns the team has chosen to call out.

## Layers

| Layer | What it answers | When loaded |
|---|---|---|
| [`principles/`](principles/README.md) | What does good engineering look like, regardless of stack? | When following a forward-link from `guides/principles/<file>.md`, or when researching a topic |
| [`idioms/`](idioms/README.md) | How does *this* stack express those principles? | When following a forward-link from `guides/idioms/<stack>/<file>.md`, or when picking up a new region |
| [`domain/`](domain/README.md) | What does the product do, what does it ship, what survives a release? | When following a forward-link from `guides/domain/<file>.md`, or when reasoning about scope and invariants |
| [`state/`](state/README.md) | What is true about the codebase right now? | At the start of planning (**orient**), and whenever a sensor reads or writes state |

Process is not represented here — the workflow phases are entirely prescriptive, so they live under [`../guides/process/`](../guides/process/README.md).

## Pairing with guides

For each principle, idiom, or domain concern that has rules:

- The **corpus** file holds the full explanation, citations, anti-patterns, and references.
- The **guide** file at the parallel path holds the rule sections (IRON LAW / GOLDEN RULES) and a `Traces to:` footer pointing back.

Corpus files include a forward-link to the paired guide when one exists:

> **Rules extracted:** [`guides/<layer>/<name>.md`](../guides/<layer>/<name>.md).

## Authorship

- `principles/` — foundational literature; team-curated additions that meet the bar (universal, language-agnostic, cited).
- `idioms/` — lead engineer + agent (via Learning flywheel).
- `domain/` — domain expert + lead engineer.
- `state/` — agent + human (state sensors propose diffs; humans accept).

## Activation

On-demand only. The agent loads a corpus file when:

- It follows a forward-link from a paired guide (the most common path — "why does this rule exist?").
- A process phase explicitly names a corpus file (e.g., the **orient** action reads `state/CODEBASE_STATE.md`).
- The user asks a question that requires reasoning beyond what the rules carry.

The corpus is not pre-loaded into context. This keeps the always-on context budget small and lets the agent reach for deeper material only when it needs to.

## Format

Each layer documents its file format in its own README. Common convention:

- One-paragraph intro under the H1.
- Forward-link to the paired guide (when one exists), as a blockquote near the top.
- Body sections describe the topic in depth.
- No `## IRON LAW` or `## GOLDEN RULES` sections — those live in the paired guide.

## Changes when

- `principles/` — almost never. Editing one often suggests you've mis-classified an idiom.
- `idioms/` — when the stack does (new framework, version upgrade, pattern adopted).
- `domain/` — when the product does (new invariant, vocabulary shift).
- `state/` — continuously. State is empirical; sensors keep it current.
