# GitHub Copilot — Lifecycle binding

How each abstract lifecycle action is invoked in GitHub Copilot.

This adapter covers both **Copilot in VS Code** (chat + edit + agent modes) and the **Copilot CLI** (`gh copilot suggest`, `gh copilot explain`, and the standalone `copilot` CLI). Both read the same project menu file (`.github/copilot-instructions.md`) and share the same conventions.

## Invocation

Every action is invoked via natural language: "run task on TICKET-123," "run verify," "do a review pass." The agent reads `.github/copilot-instructions.md` at session start, finds the action in the bulleted list, follows the link to `charter/actions/<action>.md`, and executes the playbook. Copilot has no slash-command primitive for project actions — natural-language invocation is the only path, which is why the playbooks live agent-agnostic in `charter/actions/`.

The canonical kickoff phrase is **"run task on `<ticket-id>`"** (or "run the task workflow") — `charter/actions/task.md` orchestrates `spec → orient → implementation → check-drift → verify → review`.

## GitHub-native integrations

Copilot's structural advantage: native, first-class GitHub integration. The charter should reach for these primitives where they exist:

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

The agent mode in VS Code is the closest match to the charter's full workflow; the CLI is well-suited to one-shot suggestions and quick edits inside an existing session.

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
