---
name: keystone-check-drift
description: 'Run the drift sensor on the current diff — fast pre-verify check that loaded charter rules still match what the code is doing. Triggers: keystone check-drift; keystone:check-drift; /keystone:check-drift; check drift; run drift sensor; drift check.'
---

# keystone:check-drift — fast pre-verify drift check

Compares the in-progress diff against loaded charter rules. Fires
before the full computational sensor suite — catches the obvious case
where the agent is implementing something the rules explicitly
disallow.

Canonical playbook: `.agents/commands/check-drift.md`. Sensor
contract at `.agents/hooks/drift.md`. Full discipline at
`.agents/rules/process/implementation.md`.

## Run

Open `.agents/commands/check-drift.md` and execute every
activity.

## When to trigger

- Mid-implementation, before running `/keystone:verify`.
- After a large refactor that touched many regions — make sure the
  loaded idiom rules still apply.
- When `/keystone:verify` keeps failing on the same rule — drift is
  often the upstream cause.
