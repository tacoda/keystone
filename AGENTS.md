# Agent orientation

This project uses a **keystone harness** — an agent-agnostic framework
for guides, sensors, actions, playbooks, and personas. Every
host-native surface (Claude Code skills, Cursor rules, Aider
conventions) is projected from the canonical sources under
`.keystone/harness/`.

## Read first

`.keystone/INDEX.lite.json` — cheap discovery (kind + id +
description per primitive). Browse this to pick what you need; open
the full `.keystone/INDEX.json` only when you need a primitive's
path, globs, or triggers.

## Activate by kind

- **guide** — touched files match the entry's `globs` (or no globs declared).
- **corpus** — a guide's `traces:` (or a prose forward-link) points at it.
- **action** — user intent matches `description`; the body is the playbook.
- **playbook** — composed sequence of actions.
- **sensor** — fires under an action's phase, narrowed by `globs`.
- **persona** — spawned as a subagent for narrow review/scout work.

## Iron laws (non-negotiable)

- No proceeding without explicit acceptance criteria.
- No completion claims without fresh verification — sensors run this turn.
- No commits with failing sensors. Never `--no-verify`.
- No AI attribution in commits, PRs, or tracker comments.
- No silent overwrites of state files.
- No reading or writing sensitive files (`.env*`, `*.pem`, `credentials.json`, secrets dirs).
- No dangerous action without explicit in-turn confirmation (`rm -rf`, force-push, prod DB writes, external comms).
- No invented imports, methods, config keys, or CLI flags — grep first.
- No "while I'm here" cleanups — every changed line traces to the request.
- No accepting a subagent's "done" report as evidence — read the diff.

## Lifecycle

To start a unit of work, say **"run task on `<ticket-id>`"** — runs the
`task` playbook. For one action, ask in natural language ("run verify",
"do a review pass"). The action's body lives at its INDEX `path`.

## Override

Project files at `.keystone/harness/<kind>/<id>.md` always win. Among
installed policies, deeper nesting refines outer policies; `strict:` is
absolute.
