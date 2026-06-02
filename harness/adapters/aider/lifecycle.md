# Aider — Lifecycle binding

How each abstract lifecycle action is invoked in Aider.

## Action → invocation

Aider has no slash-command primitive (its `/<command>` set is for shell, git, and chat control — not for project actions). Lifecycle actions are invoked by **typing the action name as natural language in the chat**, optionally preceded by reading the matching phase doc.

| Action | Invocation | What happens |
|---|---|---|
| **spec** | "Start the spec phase for `<task>`." | Aider reads `harness/guides/process/spec.md` and follows its activities. |
| **orient** | "Orient for work in `<region>`." | Aider reads `harness/corpus/state/CODEBASE_STATE.md` and the matching idioms; sketches a plan. |
| **check-drift** | "Check the diff for drift." | Aider compares `git diff` against loaded corpus rules. |
| **verify** | "Run the verify action." | Aider invokes sensors via `/run <cmd>` and `/test`; reports results. |
| **review** | "Run the review action." | Aider walks the diff against spec AC, then runs functional and security review concerns **sequentially**. |
| **learn** | "Capture the learnings from this work." | Aider writes a candidate to `harness/learning/inbox/<timestamp>-<slug>.md`. |
| **bootstrap** | "Bootstrap the harness." | One-time; detects stack, frameworks, and libraries; seeds corpus (idioms/<stack>/, state/), paired guides (idioms/<stack>/); confirms sensor commands; inventories computational guides (LSPs, formatters, editor enforcement) into `guides/computational/`; classifies sensors by kind. Post-bootstrap, every applicable guide and sensor is recorded in `corpus/state/CODEBASE_STATE.md`. |
| **audit** | "Audit the corpus." | Full Learning + Pruning flywheel pass. |
| **synthesize** | "Synthesize the inbox." | Promotes inbox items into the right corpus layer. |
| **mode** | Edit `harness/guides/process/modes.md` directly. | Aider has no autonomy levers; the file is informational. |

## Phase docs as the playbook

Aider's value is that it works straightforwardly with markdown — reading a phase doc and following its activities is its native shape. The phase docs themselves (`harness/guides/process/*.md`) are the playbook; the lifecycle action just names which one to read.

The corollary: keep the phase docs clear and executable. Aider does not have a slash command layer to paper over a vague phase doc.

## Sub-agent support

None. Aider runs a single conversation; the **review** action runs each review concern sequentially over the diff in the same session.

## Modes

Aider has no autonomy levels in the harness sense. There is:

- **Default chat** — Aider applies edits as it suggests them; you can undo via `/undo`.
- **`/architect`** mode — Aider explains intent first and asks before applying. Useful for the spec and planning phases.
- **`/ask`** mode — Aider answers questions without editing. Useful for review.

The harness's `paired`/`solo`/`autopilot` modes collapse to **paired** in practice — Aider applies edits immediately; the user reviews each diff in `/undo` reach.

## Capability matrix

| Capability | Supported? |
|---|---|
| Autonomous shell execution | ✓ (`/run <cmd>` and `/test`) |
| Sub-agent parallelism | ✗ |
| Autonomy levels (paired/solo/autopilot) | ✗ — effectively paired |
| Lazy-by-region | ✗ — Aider reads what the user passes via `--read` or includes via `/add`; no auto-load by path glob |
| Context-reset primitive | `/clear` (clear chat history); `/drop` (remove files from context) |
| Tracker integration | none native — user pastes card content |
| GitHub integration | partial (via `gh` in `/run`) |
