# Archive

Pruned corpus content, kept with the reasoning for why it was pruned. The Pruning flywheel writes here; nothing else does.

## Why not just delete?

History matters. A rule that was true for two years and then became wrong is more valuable as an archived record than as an absence. Future readers ask "did we ever consider X?" — the archive answers.

## Layout

Mirrors the corpus structure as needed:

```
archive/
├── README.md
├── principles/      (rare — principles do not go stale)
├── idioms/
├── domain/
└── process/
```

When the **audit** action archives a file, it moves it under the matching layer directory, prefixed with the archival date.

## Archive entry format

The archived file keeps its original content, plus a header inserted at archive time:

```markdown
---
archived_at: <ISO date>
archived_by: <command / human>
reason: <factually-wrong | aspirationally-stale | domain-stale | process-stale>
replaced_by: <path to new file, or "(none)">
---

# (original file contents follow)
```

## Activation

Never. The archive is read only by humans during audit cadence, and by the **audit** action to avoid re-promoting recently-pruned ideas.

## Reload after archive

When the **audit** action archives a file, the active session still has the archived rule loaded — ambient corpus content is loaded once per session. The action ends with a **reload prompt**: reset the agent's context (see `harness/adapters/<your-agent>/activation.md`) and re-prompt to drop the archived rule from context.

## Changes when

The Pruning flywheel fires. The **audit** action detects discrepancies between corpus claims and codebase reality; staleness gets classified and archived.
