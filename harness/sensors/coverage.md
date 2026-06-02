# Sensor: coverage

Reads test coverage and updates the State layer.

- **Trigger** — verification phase (proposes state update), **audit**.
- **Inputs** — the project's coverage command from `corpus/state/CODEBASE_STATE.md`. Skipped if no coverage tool is configured.
- **Exit condition** — coverage report produced. No minimum threshold by default — projects may set one in `corpus/state/CODEBASE_STATE.md`.
- **Output** — coverage stats per region.
- **State writes** — proposes a diff to `corpus/state/CODEBASE_STATE.md` updating coverage per region. User accepts or edits.
