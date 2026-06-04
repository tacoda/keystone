# Lifecycle actions

Each file here is the **canonical playbook** for one harness lifecycle action. They're agent-agnostic: read by Claude Code, Cursor, Codex, Aider, GitHub Copilot, Continue, Cline, Goose, pi, or any other coding agent the project's menu file points here.

## How invocation works

The agent reads its menu file (`CLAUDE.md`, `AGENTS.md`, `CONVENTIONS.md`, `.continuerules`, `.cursor/rules/keystone.mdc`, etc.) on session start. The menu file lists every action with a one-line description and a link to its playbook here. When the user asks to "run bootstrap" (or "run the bootstrap action," or "do a verify pass"), the agent follows the link and executes the playbook.

No agent-specific slash commands, rule files, or prompts. One source of truth per action.

## Kicking off an entire unit of work

For the end-to-end lifecycle on a single task, say **"run task on `<ticket-id>`"** (or "run the task workflow"). The agent reads [`task.md`](task.md), which walks `spec → orient → implementation → check-drift → verify → review` and an optional `learn` pass, pausing for confirmation between phases.

## The actions

| Action | Playbook | When |
|---|---|---|
| **task** | [`task.md`](task.md) | The kickoff — chains spec → orient → implementation → check-drift → verify → review. |
| **bootstrap** | [`bootstrap.md`](bootstrap.md) | Once per project — detect stack, seed state, classify sensors. |
| **spec** | [`spec.md`](spec.md) | First action on any task — capture intent + acceptance criteria. |
| **orient** | [`orient.md`](orient.md) | After spec — load `CODEBASE_STATE.md` and idioms for the touched region; sketch a plan. |
| **check-drift** | [`check-drift.md`](check-drift.md) | Between implementation and verify — fast drift check on the diff. |
| **verify** | [`verify.md`](verify.md) | Pre-commit — lint / type-check / test / build / drift / commit-message sensors. |
| **review** | [`review.md`](review.md) | Post-verify — functional / security / risk / deployment review + spec-adherence. |
| **learn** | [`learn.md`](learn.md) | Any time something surprising happens — capture to `learning/inbox/`. |
| **audit** | [`audit.md`](audit.md) | Periodic — dual flywheel (Learning + Pruning). |
| **synthesize** | [`synthesize.md`](synthesize.md) | After inbox has accumulated — promote into corpus / guides. |
| **mode** | [`mode.md`](mode.md) | When pacing needs to change — paired / solo / autopilot. |

Each playbook is short (~20–40 lines): the agent reads it, executes the listed activities, and reports back. Deeper context lives in `harness/guides/process/`, `harness/sensors/`, and `harness/learning/`; playbooks forward-link to those when they apply.
