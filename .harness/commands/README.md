# Actions

Each file here is one **action** — a single unit of lifecycle work. Actions are agent-agnostic: read by Claude Code, Cursor, Codex, Aider, GitHub Copilot, Continue, Cline, Goose, pi, or any other coding agent the project's menu file points here.

For ordered sets of actions, see [`harness/playbooks/`](playbooks/README.md). For example, the end-to-end task workflow lives at [`harness/playbooks/task.md`](playbooks/task.md).

## How invocation works

The agent reads its menu file (`CLAUDE.md`, `AGENTS.md`, `CONVENTIONS.md`, `.continuerules`, `.cursor/rules/keystone.mdc`, etc.) on session start. The menu file lists every action with a one-line description and a link. When the user asks to "run bootstrap" (or "run the bootstrap action," or "do a verify pass"), the agent follows the link and executes the action.

No agent-specific slash commands, rule files, or prompts. One source of truth per action.

## The actions

| Action | File | When |
|---|---|---|
| **bootstrap** | [`bootstrap.md`](bootstrap.md) | Once per project — detect stack, seed state, classify sensors. |
| **spec** | [`spec.md`](spec.md) | First action on any task — capture intent + acceptance criteria. |
| **orient** | [`orient.md`](orient.md) | After spec — load `CODEBASE_STATE.md` and idioms for the touched region; sketch a plan. |
| **check-drift** | [`check-drift.md`](check-drift.md) | Between implementation and verify — fast drift check on the diff. |
| **verify** | [`verify.md`](verify.md) | Pre-commit — lint / type-check / test / build / drift / commit-message sensors. |
| **review** | [`review.md`](review.md) | Post-verify — functional / security / risk / deployment review + spec-adherence. |
| **learn** | [`learn.md`](learn.md) | Any time something surprising happens — capture to `learning/inbox/`. |
| **audit** | [`audit.md`](audit.md) | Periodic — dual flywheel (Learning + Pruning); Pruning writes to `corpus/state/harness-debt.md`. |
| **debt-review** | [`debt-review.md`](debt-review.md) | Periodic — triage `corpus/state/code-debt.md`. |
| **policy-audit** | [`policy-audit.md`](policy-audit.md) | Before release or per audit cycle — check codebase against installed policy guides; verify org `strict` items aren't being overridden. |
| **synthesize** | [`synthesize.md`](synthesize.md) | After inbox has accumulated — promote into corpus / guides. |
| **mode** | [`mode.md`](mode.md) | When pacing needs to change — paired / solo / autopilot. |

Each action is short (~20–40 lines): the agent reads it, executes the listed activities, and reports back. Deeper context lives in `harness/guides/process/`, `harness/sensors/`, and `harness/learning/`; actions forward-link to those when they apply.

## Override cascade

For any `<name>.md`, the project's `harness/actions/<name>.md` always wins by default. Among policies, policies nested deeper in `keystone.json` refine the outer policies they're nested in. A policy can mark an item `strict` to make it absolute — nothing else can override a strict item, not the project, not any other policy. `keystone verify` reports a violation if any layer attempts to shadow a strict item.
