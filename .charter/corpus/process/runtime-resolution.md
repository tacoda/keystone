---
kind: corpus
id: corpus/process/runtime-resolution
description: Why the runtime resolution flow escalates instead of fanning out — staged lookup costs less and keeps decisions auditable.
---
# Runtime resolution — reasoning

The staged flow exists because all-at-once retrieval is expensive,
noisy, and unauditable.

## Why escalate instead of fanning out

A naive approach is "load every rule, every corpus entry, every
external source for the touched files." The result:

- **Context blowout.** Even modest projects have dozens of guides and
  many more corpus entries. Loading them all on every scenario fills
  the agent's window with content that doesn't apply.
- **Noise crowds out signal.** Five corpus entries scrolling past one
  iron-law rule means the iron law gets weighted less in the agent's
  reasoning. The signal is the rule; corpus is supporting evidence,
  not equal voice.
- **External APIs are slow and rate-limited.** Linear, Confluence,
  team wikis all carry latency and quota. Fanning out on every
  scenario burns both.
- **Untraceable decisions.** When everything is in context, you
  can't tell which rule actually informed the agent's choice. Staged
  retrieval makes the lineage obvious — "we acted on rule X because
  stages 1–2 covered it."

Staged escalation flips this: cheap first, expensive last, ask before
applying. Each stage is a deliberate choice with a clear "why didn't
the previous stage suffice?" answer.

## Anti-patterns

- **Loading corpus by default.** Corpus is the *why*, not the *what*.
  Opening it for every rule pollutes context with reasoning the agent
  doesn't need to act.
- **Querying external sources without a rule prompt.** Sources exist
  to fill gaps the charter left. Going outside-in (external first,
  charter second) means the charter no longer governs — the source
  does.
- **Applying external findings silently.** "I found this in the team
  wiki" is not authorization. The user controls what gets recorded
  and at what layer.
- **Swallowing contradictions.** When two sources disagree, picking
  one quietly hides the conflict. The user has the context to
  resolve it; the agent does not.

## References

- *A Philosophy of Software Design*, John Ousterhout — chapter 5
  (information hiding) and chapter 10 (defining errors out of
  existence) inform why staged retrieval beats firehose retrieval.

Back to the rules: [`guides/process/runtime-resolution.md`](../../guides/process/runtime-resolution.md).
