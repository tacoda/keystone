# Charter

This repo is governed by a **keystone charter** — the authored standards
that constrain whatever coding-agent harness runs here (Claude Code,
Cursor, Codex, opencode, …). A *harness* is the engine that runs the
model; the *charter* is what you author to make its output reliable.
Author the spec → charter. Be the engine → harness.

The charter is a tree of typed primitives under `.charter/`. Every
host-native surface (`.claude/`, `.cursor/`, this file) is projected
from it — do not hand-edit projected files.

## Read first

`.charter/INDEX.lite.json` — cheap discovery (kind + id + description
per primitive). Browse it to pick what you need; open the full
`.charter/INDEX.json` only when you need a primitive's path, globs, or
triggers, and open a body only when its activation condition matches.

## Activate by kind

- **guide** — touched files match the entry's `globs:` (or it declares none).
- **corpus** — a guide's `corpus:` (or a prose forward-link) points at it — the *why*, on demand.
- **command** — user intent matches `description` + `phase`; the body is a unit of work.
- **playbook** — a composed sequence of commands with human `gates:`.
- **sensor** — a phase-gated check; inferential → an agent review, computational → a host hook.
- **hook** — fires deterministically on an `event:` → `run:` shell or `agent:` dispatch.
- **skill** — auto-activates by `triggers:` match.
- **agent** — a role spawned as a subagent by `id`; the body is its system prompt.
- **pattern** — a reusable documentation pattern; apply when writing docs.

## Iron laws (non-negotiable, every phase)

- No proceeding without explicit acceptance criteria.
- No completion claims without fresh verification — checks run this turn, against post-edit state, with cited output.
- No commits with failing checks. Never `--no-verify`.
- No AI attribution in commits, PRs, or tracker comments.
- No silent overwrites of state files.
- No reading or writing sensitive files (`.env*`, `*.pem`, `*.key`, `credentials.json`, secrets dirs) — ask out-of-band.
- No dangerous action without explicit in-turn confirmation (`rm -rf`, force-push, `reset --hard`, prod DB writes, external comms, installs).
- No invented imports, methods, config keys, or CLI flags — grep / read the manifest / check `--help` first.
- No "while I'm here" cleanups — every changed line traces to the request.
- No accepting a subagent's "done" report as evidence — read the diff; re-run checks in the parent turn.

## Lifecycle

To start a unit of work, say **"run task on `<ticket-id>`"** — runs the
`task` playbook. For one command, ask in natural language ("run verify",
"do a review pass"); its body lives at its INDEX `path`.

## Override

Project files at `.charter/<kind>/<id>.md` always win. Among installed
policies, deeper nesting refines outer policies; a `strict:` item is
absolute — nothing else overrides it.
