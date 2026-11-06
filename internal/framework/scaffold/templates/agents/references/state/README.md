# State references

The map of the codebase as it is now, not as it should be. Hooks and people both write here, and these files change with the code.

## What lives here

The `bootstrap` command fills in these templates:

- `CODEBASE_STATE.md`: the regions, the stack, and the lint, test, and build commands the hooks need.
- `risk-fingerprints.md`: complexity, coupling, and coverage per region.
- `traffic-topology.md`: git churn and recency, plus business criticality.
- `quality-radar.md`: a five-dimension scorecard, updated by `audit` and `review`.
- `code-debt.md`: debt in the codebase, updated by `debt-review` and `audit`.
- `charter-debt.md`: debt in the charter itself, updated by `audit`.
- `INSTALL_PROFILE.md`: the selections `keystone init` captured.
- `GLOBS_INDEX.md`: a reverse index of the rules that declare paths.

Migrations get one file each:

- `migrations/active/<name>.md`: in-flight work, such as an expand-contract change.
- `migrations/completed/<name>.md`: finished migrations, kept as a record.

## When the agent reads it

The `orient` command reads `CODEBASE_STATE.md` at the start of planning. Other commands and hooks work on the file they own. For example, the `risk-fingerprint` hook proposes updates to `risk-fingerprints.md`.

## Who writes it

Hooks and commands propose a diff. A person accepts or edits it. No agent overwrites a state file without that confirmation.

Every file records the date it was last reconciled. Every migration record states its expected end state.
