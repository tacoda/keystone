# Hook

<!-- voice-register: informational -->
<!-- voice-english: ste -->

A hook is a check that runs on an event and can block the agent. It's a constraint, and the strongest one keystone has, because the agent can't skip it on a host that runs hooks. A computational hook runs a shell command, while an inferential hook asks a model for a verdict. The cost is the binary: each host bridge calls the `keystone` on your `PATH`, so every machine that runs agents needs the 5.0 binary.

## Example

```markdown
---
id: build
description: The project's build step.
mode: computational
event: Stop
run: go build ./...
---

# build

A non-zero exit means the build is broken. Read the compiler error and fix it before you claim done.
```

Save it as `.agents/hooks/build.md`. Then run `keystone index && keystone project`.

## Where it lives

```
.agents/hooks/<id>.md
```

Keystone infers `kind: hook` from the `hooks/` folder.

## Frontmatter

| Key | Required | Meaning |
|---|---|---|
| `id` | yes | Stable id |
| `description` | yes | One sentence on what the check catches |
| `mode` | yes, in practice | `computational` or `inferential`. Lint treats a missing `mode` as inferential |
| `event` | yes, to fire | A host event or a signal. Without it, the hook never fires |
| `run` | computational | The shell command. Lint fails without it |
| `returns` | inferential | The verdict schema the agent must return. Lint fails without it |
| `agent` | no | The subagent to dispatch. The default is the hook's own id |
| `tags` | no | Kebab-case labels for `keystone list --tag` |

The 4.x spellings still parse: `kind: sensor` and `on:`. `keystone lint` warns, and `keystone lint --fix` rewrites them.

## Events

- **A host event** is one the host fires: `PreToolUse`, `PostToolUse`, `UserPromptSubmit`, `Notification`, `Stop`, `SubagentStop`, `PreCompact`, `SessionStart`, or `SessionEnd`.
- **A signal** is any other name. For example, keystone fires `pre-verify` and `post-verify` around `keystone verify`. You can fire any signal with `keystone signal fire <name>`. `keystone signal list` shows the built-ins and the ones `keystone.json` `signals:` declares.

## How a computational hook runs

`keystone project` writes one bridge per host event: `keystone hook fire <event> --host <host>`. When the host fires the event, the bridge does these steps:

1. It searches upward from the current folder for `.agents/`.
2. It selects every primitive with this `event:` and a `run:`.
3. It runs each `run:` command in parallel.
4. It sets `KEYSTONE_EVENT`, `KEYSTONE_PHASE`, `KEYSTONE_COMMAND`, `KEYSTONE_TYPE`, and `KEYSTONE_HOST` for each command.
5. On a host event, it also passes the host's event JSON on stdin.

If any command exits non-zero, the bridge exits 2, which the host treats as a block. Also, on a host event, a command that runs past 50 seconds counts as a failure. A missing binary or a bad flag also exits 2, so the bridge fails closed.

There is no per-hook matcher. A `PreToolUse` hook runs before every tool call, so read the stdin payload to decide which calls to check.

## How an inferential hook runs

The hook's body is the review prompt, but keystone doesn't call a model itself. On the event, `keystone hook fire` prints a dispatch list: each agent to spawn and the `returns:` schema it must fill. Then the host's agent spawns them.

On Claude Code, `keystone project` also writes the hook as a subagent in `.claude/agents/`. An inferential hook on a host event can't block. So lint warns about that case. Use a `run:` check when the hook must gate.

## How each host gets it

- **Claude Code**: bridges merged into `.claude/settings.json`, next to your own hooks.
- **Cursor**: bridges merged into `.cursor/hooks.json`, with `failClosed: true`.
- **GitHub Copilot**: bridges in `.github/hooks/keystone.json`.
- **opencode**: a keystone-owned plugin.
- **Every other host**: no bridge. `keystone project` reports it, and `AGENTS.md` tells the agent to run the checks in `.agents/hooks/` itself.

[`docs/hosts.md`](../hosts.md) lists the open verification items per host.

## What a hook does not do

- A hook doesn't block on a host without a bridge. There, it's an instruction the agent can skip.
- An inferential hook doesn't produce a verdict by itself. Instead, the host's agent must run the dispatched review.
- A signal never reaches the host. It comes only from keystone or from your own script.
- `paths:` doesn't filter when a hook fires.

## Failure modes

- **The hook never runs.** It has no `event:`, or the host has no bridge. Then check the `keystone project` report.
- **Lint says an inferential hook is missing `returns:`.** The hook has `run:` but no `mode: computational`. Add the mode.
- **Checks stopped after an upgrade.** A 4.x binary is still on `PATH`. Install 5.0 on that machine.
- **The hook fires twice.** You enabled both Claude Code and Copilot, and Copilot also reads `.claude/settings.json`. `keystone project` warns about this.

## Authoring

```
keystone new hook <id>
```

This writes `.agents/hooks/<id>.md` with `mode: computational`, `event: Stop`, and a placeholder `run: "true"`. Replace the placeholder, because it always passes. `keystone new sensor` exits with an error that names `keystone new hook`.
