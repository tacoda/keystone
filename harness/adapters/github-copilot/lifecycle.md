# GitHub Copilot — Lifecycle binding

How each abstract lifecycle action is invoked in GitHub Copilot.

This adapter covers both **Copilot in VS Code** (chat + edit + agent modes) and the **Copilot CLI** (`gh copilot suggest`, `gh copilot explain`, and the standalone `copilot` CLI). Both read the same project menu file (`.github/copilot-instructions.md`) and share the same conventions.

## Action → invocation

Copilot has no slash-command primitive for project actions. Lifecycle actions are invoked by **asking the agent in natural language**, optionally referencing the matching phase doc.

| Action | Invocation | What happens |
|---|---|---|
| **spec** | "Start the spec phase for `<task>`." | Reads `harness/process/spec.md` and follows its activities. |
| **orient** | "Orient for work in `<region>`." | Reads `harness/state/CODEBASE_STATE.md` and matching idioms; sketches a plan. |
| **check-drift** | "Check the diff for drift." | Compares `git diff` against loaded corpus rules. |
| **verify** | "Run the verify action." | Invokes sensors via the shell; reports results inline. |
| **review** | "Run the review action." | Walks the diff sequentially against spec AC, functional concerns, security concerns. |
| **learn** | "Capture the learnings from this work." | Writes a candidate to `harness/learning/inbox/<timestamp>-<slug>.md`. |
| **bootstrap** | "Bootstrap the harness." | One-time; populates `harness/idioms/<stack>/` and `harness/state/`. |
| **audit** | "Audit the corpus." | Full Learning + Pruning flywheel pass. |
| **synthesize** | "Synthesize the inbox." | Promotes inbox items into the right corpus layer. |
| **mode** | Edit `harness/process/modes.md` directly. | Copilot has limited autonomy levers; the file is informational. |

## GitHub-native integrations

Copilot's structural advantage: native, first-class GitHub integration. The harness should reach for these primitives where they exist:

- **Issue and PR context** — Copilot can fetch issue/PR descriptions and comments natively. The **spec** action uses this for tracker-card-fetching (where the tracker is GitHub Issues).
- **`gh` CLI** — available in every Copilot environment; the **release** phase uses `gh pr create` and `gh pr view`.
- **Code search via `gh`** — Copilot can run `gh search code` and `gh search issues` for cross-repo lookups.
- **Workflow runs** — `gh run list` / `gh run view` for CI status; used in the **release** phase as the CI sensor.

Where tracker is **not** GitHub Issues (Jira / Linear / Asana), Copilot falls back to the user pasting the card content.

## VS Code vs. CLI differences

| Concern | Copilot in VS Code | Copilot CLI |
|---|---|---|
| Menu file | `.github/copilot-instructions.md` | Same |
| Shell execution | ✓ in agent mode | ✓ native |
| Edit application | ✓ inline diffs in editor | Stdout suggestions; user applies |
| Multi-file edits | ✓ in agent mode | Manual |
| Sub-agents | ✗ | ✗ |
| Context-reset | "New chat" in chat panel | New session |

The agent mode in VS Code is the closest match to the harness's full workflow; the CLI is well-suited to one-shot suggestions and quick edits inside an existing session.

## Sub-agent support

None. The **review** action runs each review concern sequentially over the diff.

## Modes

Copilot's autonomy lever is **per-command approval** — agent mode in VS Code prompts before running each shell command; the user approves or denies. There is no `autopilot` equivalent. Treat all sessions as `paired`.

## Capability matrix

| Capability | Supported? |
|---|---|
| Autonomous shell execution | ✓ (agent mode in VS Code; native in CLI) |
| Sub-agent parallelism | ✗ |
| Autonomy levels (paired/solo/autopilot) | ✗ — effectively paired |
| Lazy-by-region rule loading | ✗ |
| Context-reset primitive | New chat / new session |
| Tracker integration | ✓ native for GitHub Issues; paste for others |
| GitHub integration | ✓ (`gh` CLI, native PR/issue context) |
