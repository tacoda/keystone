---
kind: rule
id: process/iron-laws
description: 'Iron laws (every phase) + the cascade override rule — moved here from the generated AGENTS.md block at 6.0'
---

# Iron laws (every phase)

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

## Override

Project files in `.agents/` win over installed policies (`.keystone/policies/`). Deeper policies refine outer ones; a `strict:` item is absolute.
