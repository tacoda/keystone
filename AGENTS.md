<!-- keystone:start -->
# Charter

This repo is governed by a **keystone charter**: the standards that shape how any coding agent works here. The charter has two halves. **Constraints** limit, gate, or check what you do. **Capabilities** add things you can do.

The sources live in `.agents/`; keystone's index and state live in `.keystone/`. Host files (`.claude/`, `.cursor/`, the region of this file between the keystone markers) are generated from `.agents/`: edit the sources, never the generated files.

## Read first

`.keystone/INDEX.lite.json` lists every primitive (category, kind, id, description). Browse it, then open a body only when it applies to the task.

## Constraints

- **rule** (`.agents/rules/`) — applies when a touched file matches its `paths:`; a rule without `paths:` applies everywhere.
- **reference** (`.agents/references/`) — the reasoning behind a rule; open it when you need the why.
- **hook** (`.agents/hooks/`) — a check that runs on an event and can block; where your host cannot run it, run it yourself before claiming done.
- **permission** (`.agents/permissions/`) — allow / ask / deny rules for tools.
- **eval** (`.agents/evals/`) — tests that the charter changes agent behavior.

## Capabilities

- **skill** (`.agents/skills/`) — activates when its description matches the task.
- **command** (`.agents/commands/`) — a unit of work the user asks for by name.
- **subagent** (`.agents/agents/`) — a role you delegate to.
- **tool** (`.agents/tools/`) — a callable, served by the keystone MCP server.
- **document** (`.agents/documents/`) — a template for governed output; instances live in `.keystone/work/`.

## Rules that apply everywhere

Open the one that fits the phase you are in.

- `.agents/rules/process/ci-failure.md` — What to do when CI fails — on a PR, after a merge, or post-deploy.
- `.agents/rules/process/context-budget.md` — How much the agent reads, runs, or spawns before making a change.
- `.agents/rules/process/dangerous-actions.md` — Operations the agent must confirm with the user before executing — irreversible, broadly-visible, or outside the local workspace.
- `.agents/rules/process/escalation.md` — When the agent is stuck, when to stop and ask the user — versus when to keep going.
- `.agents/rules/process/grounding.md` — The discipline of verifying that what the agent references — functions, APIs, packages, files, flags — actually exists before invoking it.
- `.agents/rules/process/implementation.md` — The phase where code gets written, test-first.
- `.agents/rules/process/modes.md` — Active mode: paired
- `.agents/rules/process/planning.md` — The phase that turns an approved spec into an approved plan, before code is written.
- `.agents/rules/process/pushback.md` — The discipline of disagreeing with the user when the user is wrong.
- `.agents/rules/process/release.md` — The phase where the change lands and is announced.
- `.agents/rules/process/review.md` — The phase that confirms a mechanically-correct change is also the right change.
- `.agents/rules/process/runtime-resolution.md` — How the agent resolves a scenario — staged escalation from rules to corpus to external sources, asking before applying.
- `.agents/rules/process/scoping.md` — How big any single commit, PR, or implementation pass is allowed to get before it must split.
- `.agents/rules/process/self-validation.md` — The discipline of refusing to count the agent's own claim as evidence.
- `.agents/rules/process/sensitive-files.md` — Files the agent must never read into context, write to, or commit.
- `.agents/rules/process/spec.md` — The phase that captures what needs to happen, before deciding how.
- `.agents/rules/process/subagent-trust.md` — The discipline of treating subagent reports as claims to verify, not as evidence to accept.
- `.agents/rules/process/surgical-edits.md` — Touch only what serves the task.
- `.agents/rules/process/verification.md` — The phase where claims are paired with evidence before commit.

## Iron laws (every phase)

- No proceeding without explicit acceptance criteria.
- No completion claims without fresh verification: checks run this turn, against post-edit state, with cited output.
- No commits with failing checks. Never `--no-verify`.
- No AI attribution in commits, PRs, or tracker comments.
- No silent overwrites of state files.
- No reading or writing sensitive files (`.env*`, `*.pem`, `*.key`, `credentials.json`, secrets dirs); ask out-of-band.
- No dangerous action without explicit in-turn confirmation (`rm -rf`, force-push, `reset --hard`, prod DB writes, external comms, installs).
- No invented imports, methods, config keys, or CLI flags: grep, read the manifest, or check `--help` first.
- No "while I'm here" cleanups: every changed line traces to the request.
- No accepting a subagent's "done" report as evidence: read the diff and re-run checks yourself.

## Lifecycle

To start a unit of work, say **"run task on `<ticket-id>`"**; it runs the `task` skill. For one command, ask in plain words ("run verify"); its body is at its INDEX `path`.

## Override

Project files in `.agents/` win over installed policies (`.keystone/policies/`). Deeper policies refine outer ones; a `strict:` item is absolute.

## On each host

- **Claude Code** — loads rules, hooks (inferential), skills, commands, subagents by itself.
- **Cursor** — loads rules, skills, commands, subagents by itself.
- **Continue** — loads rules, skills, commands by itself; open subagents from `.agents/` yourself.
- **Aider** — open rules, skills, commands, subagents from `.agents/` yourself.
<!-- keystone:end -->
