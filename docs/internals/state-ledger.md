# State ledgers

<!-- voice-register: informational -->
<!-- voice-english: ste -->

State ledgers record what is true about this project right now: its stack, its debt, its risky areas. Rules say what should be true, and references say why, but a state ledger says what is. Commands write the ledgers, and you can edit them too. Because each ledger is a reference, the agent opens it on demand.

## Path convention

```
.agents/references/state/<name>.md
.agents/references/state/migrations/active/<NAME>.md
.agents/references/state/migrations/completed/<NAME>.md
```

The ledgers live in the reserved `state/` folder under `.agents/references/`. Policies don't ship state ledgers, because state belongs to one project.

## Required shape

Each ledger is Markdown with reference frontmatter (`kind: reference`, `id`, `description`), so `keystone index` lists it in `.keystone/INDEX.json`. The body is free-form. The scaffold ships these ledgers:

| File | Purpose | Written by |
|---|---|---|
| `CODEBASE_STATE.md` | The map of the codebase: stack, regions, which idioms apply, coverage. | `bootstrap` |
| `INSTALL_PROFILE.md` | The choices made at `keystone init`, with a link to the lockfile. | `keystone init` |
| `GLOBS_INDEX.md` | A reverse index of every rule that declares `paths:`. Don't edit it by hand. | `bootstrap`, `synthesize`, `audit` |
| `code-debt.md` | Known debt in the codebase. | `debt-review`, `audit` |
| `charter-debt.md` | Debt in the charter itself: stale rules, missing pairs. | `audit` |
| `quality-radar.md` | A five-dimension scorecard. | `audit`, `review` |
| `risk-fingerprints.md` | Complexity, coupling, and coverage patterns per region. | `bootstrap` |
| `traffic-topology.md` | Churn, recency, and business criticality per region. | `bootstrap` |

The `migrations/active/` folder tracks in-flight work such as expand-contract changes and deprecations. When that work finishes, its record moves to `migrations/completed/`.

Add your own ledger in the same folder, and the agent reads it like the others.

## Activation

The agent opens ledgers on demand, like any reference. But one command always reads them: at the start of planning, `orient` loads `CODEBASE_STATE.md` and any active migration for the touched region. Other commands and hooks read a ledger only when they need it, the way the `risk-fingerprint` hook reads `risk-fingerprints.md`.

## Update semantics

- **Diff, then confirm.** An agent proposes a ledger change as a diff, and you accept or edit it. No agent overwrites a state file without that step.
- **Drift is normal.** A ledger follows the code, so it changes whenever the code changes.

## Cascade behavior

State ledgers live only in the project. Since no policy ships one, the strict cascade doesn't apply to them.

## Authoring

To add a ledger:

1. Write `.agents/references/state/<name>.md` with reference frontmatter.
2. Update the command that should read or write it.
3. Run `keystone index`.

There is no `keystone new` generator for a state ledger, because its shape is specific to the project.
