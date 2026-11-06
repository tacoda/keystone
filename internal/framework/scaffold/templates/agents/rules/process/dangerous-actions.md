---
kind: guide
id: process/dangerous-actions
description: 'Operations the agent must confirm with the user before executing — irreversible, broadly-visible, or outside the local workspace.'
---
# Dangerous Actions

Operations the agent must confirm with the user before executing — irreversible, broadly-visible, or outside the local workspace.

## IRON LAW

**NEVER PERFORM A DANGEROUS ACTION WITHOUT EXPLICIT, IN-TURN CONFIRMATION.**

Authorization granted once does not extend to the next invocation. Each occurrence is a separate confirmation. A mode (see [`modes.md`](modes.md)) never loosens this.

## What counts as dangerous

| Category | Examples |
|---|---|
| **Destructive filesystem** | `rm -rf`, `find -delete`, overwriting uncommitted work, dropping a worktree |
| **History rewriting** | `git push --force`, `git reset --hard`, `git rebase -i`, `git branch -D`, `git filter-branch`, `git clean -fd` |
| **Production data** | `DROP TABLE`, `TRUNCATE`, destructive `UPDATE` without `WHERE`, manual prod DB writes |
| **External communication** | Posting to Slack/Discord, sending email, commenting on someone else's repo, opening tickets in shared trackers |
| **System changes** | `brew install`, `apt install`, `sudo`, editing files outside the repo, modifying `~/.zshrc`, changing global git config |
| **Network exposure** | Opening ports, deploying, rotating secrets, changing DNS, modifying IAM |
| **Account state** | OAuth grants, API key creation, paying invoices, accepting ToS |

The **bootstrap** action specializes this list: it records which integrations are wired (Slack MCP, Linear, GitHub PR write, infra tooling) and which dangerous paths are reachable from this project.

## RULES

- **Show, then confirm.** State the exact command and its blast radius in one sentence. Wait for the user.
- **Reversibility downgrades risk, but doesn't remove it.** A `git revert` is cheap; a `git push --force` to `main` is not.
- **Default to the safest equivalent.** `git revert` over `reset --hard`. `UPDATE ... WHERE id = ...` over a bulk update. A draft PR before the real one.
- **Refuse to chain dangerous actions.** One confirmation, one action. Compound commands hide the dangerous step.
- **Investigate unfamiliar state before deleting it.** An unexpected file or branch may be the user's in-progress work.

## GOLDEN RULE

- **Aim to keep a recovery path open.** Before any destructive op, ask "if this is wrong, how do I undo it in the next minute?" If the answer is "I can't" — stop and confirm.
- **Aim to never act on shared state without a written audit trail.** A Slack message is a side effect on the team; treat it like a deploy.

## Pacing modes

- **paired** — confirm every dangerous action.
- **solo** — confirm every dangerous action. Mode does not loosen this.
- **autopilot** — confirm every dangerous action. Mode does not loosen this. See [`modes.md`](modes.md) IRON LAW.

## Anti-patterns

- `&& rm -rf …` tacked onto the end of a "normal" command.
- Confirming "yes, proceed" once and assuming that covers future runs.
- Running `git push --force` to "fix" a hook failure. Hook failures are signal — fix the cause.
- Deleting a file the agent does not recognize. Investigate first.

---

See also: [[sensitive-files]], [[ci-failure]], [`modes.md`](modes.md).
