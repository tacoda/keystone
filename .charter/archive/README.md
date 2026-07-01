# Archive

Pruned content from `guides/` and (rarely) `corpus/`, kept with the reasoning for why it was pruned. The Pruning flywheel writes here; nothing else does.

## Why not just delete?

History matters. A rule that was true for two years and then became wrong is more valuable as an archived record than as an absence. Future readers ask "did we ever consider X?" — the archive answers.

## Asymmetric pruning

The flywheel prunes **guides** regularly and **corpus** rarely:

- **Guides** churn with the codebase. A rule that names a removed API, contradicts a newer rule, or no longer reflects how the team writes code is stale and should be archived. Expect to run the guide audit on a cadence (e.g., quarterly, or alongside major refactors).
- **Corpus** is the stable layer. A principle file does not become stale because the codebase shape changed; it becomes stale when the team's *thinking* has changed — a strategic shift, a new ideal, a deprecated design philosophy. Corpus pruning is rare and deliberate.

When the two diverge — a corpus file argues for X but the guide rule for X has been pruned — the corpus *also* needs review. But pruning the corpus is the second-order effect; pruning the guide is the first.

## Layout

Mirrors the original structure:

```
archive/
├── README.md
├── guides/
│   ├── principles/
│   ├── idioms/
│   ├── domain/
│   └── process/
├── corpus/             (rare)
│   ├── principles/
│   ├── idioms/
│   ├── domain/
│   └── state/
└── sensors/            (very rare)
```

When the **audit** action archives a file, it moves it under the matching path, prefixed with the archival date.

## Archive entry format

The archived file keeps its original content, plus a header inserted at archive time:

```markdown
---
archived_at: <ISO date>
archived_by: <command / human>
source: <guides | corpus | sensors>
reason: <factually-wrong | aspirationally-stale | domain-stale | process-stale | strategy-shift>
replaced_by: <path to new file, or "(none)">
---

# (original file contents follow)
```

For **corpus** archives only: include a `strategy-shift` reason describing what the team now believes that they did not before. This is the rare case — capture it well.

## Activation

Never. The archive is read only by humans during audit cadence, and by the **audit** action to avoid re-promoting recently-pruned ideas.

## Reload after archive

When the **audit** action archives a **guide**, the active session still has the rule loaded — guides are ambient. The action ends with a **reload prompt**: reset the agent's context (see `.charter/adapters/<your-agent>/activation.md`) and re-prompt to drop the archived rule.

Archiving a **corpus** file does not require a reload — corpus is on-demand.

## Changes when

The Pruning flywheel fires. Guide audit on a cadence; corpus audit only when the team's design or strategy has moved on.
