# Tool

<!-- voice-register: informational -->
<!-- voice-english: ste -->

A tool is a callable the charter defines: a shell `run:` handler with declared arguments, which keystone runs when the agent calls it. It is not a host built-in tool like `Bash`, `Read`, or `Grep`. Those belong to the host, and a subagent's `tools:` list names them. The cost is trust: the handler is a shell script with your full environment, and keystone doesn't sandbox it.

## Example

```markdown
---
kind: tool
id: open-pr-count
description: Count the open pull requests for a GitHub label.
transport: mcp
run: gh pr list --label "$KEYSTONE_ARG_LABEL" --state open --json number --jq length
args:
  - name: label
    type: string
    required: true
    description: The label to filter on.
---

# open-pr-count

Use this before you plan a release, to see how much work is still open.
```

Save it as `.agents/tools/open-pr-count.md`. The keystone MCP server then serves it as `keystone_tool_open_pr_count`. To test it from a shell, run `keystone tool run open-pr-count --label bug`.

## Where it lives

```
.agents/tools/<id>.md
```

The body explains what the tool does and when the agent should call it instead of a plain shell command. Keystone doesn't send the body to the handler.

## Frontmatter

| Key | Required | Meaning |
|---|---|---|
| `kind` | no | `tool`. Keystone infers it from the `tools/` folder. |
| `id` | yes | The name the agent calls. |
| `description` | yes | One line. The MCP server and `AGENTS.md` show it to the agent. |
| `run` | yes | The shell command. Keystone runs it with `bash -c` in the project root. |
| `transport` | no | `cli` (the default), `mcp`, or `plugin`. |
| `args` | no | A list of `name`, `type`, `required`, and `description`. A bare string is short for `name`. |
| `event` | no | A signal that also fires the tool, as a side effect. |

`keystone lint` reports an error when `run` is missing or when `transport` has another value. Lint also accepts `http`, but no code serves an `http` tool, so don't use it.

## How the handler runs

Every transport goes through the same runner, `internal/framework/toolrun`:

1. Keystone checks the arguments. If a required argument is missing, or the tool doesn't declare an argument, the call fails before the handler runs.
2. Keystone exports each declared argument as `KEYSTONE_ARG_<NAME>`, with the name in upper case. For an argument the caller doesn't give, the variable is empty.
3. Keystone runs `run:` with `bash -c` in the project root.
4. Keystone returns stdout and stderr together. A non-zero exit is a failed call.

Every argument reaches the handler as a string, whatever its `type` says.

## Transports

| Transport | How the agent reaches it |
|---|---|
| `mcp` | `keystone mcp serve` serves it as `keystone_tool_<id>`, with `:`, `/`, and `-` turned into `_`. `AGENTS.md` lists that name. |
| `cli` | `AGENTS.md` lists it with its command, `keystone tool run <id> --<arg> <value>`. The agent runs that command through the host's shell tool. |
| `plugin` | As `cli`, plus a native host tool where the host has one. Today that's opencode only: `keystone project` writes `.opencode/tools/<id>.ts`, which calls `keystone tool run`. |

`keystone tool run` takes `--<arg> <value>` or `--<arg>=<value>`. It reserves `--dir` for the project folder, so don't name an argument `dir`. Run `keystone tool list` to see every tool with its command.

## What a tool does not do

- It doesn't replace the host's built-in tools, and it doesn't appear in a subagent's `tools:` list.
- It doesn't run in a sandbox. The handler has the same rights as the user who runs keystone.
- It doesn't get the arguments when an `event:` fires it. On a signal, keystone runs it like a hook, with `KEYSTONE_EVENT` and the other fire variables instead.
- It doesn't reach a running MCP server by itself. The server reads the tools when it starts.

## Failure modes

- **The MCP tool doesn't appear.** You added it after the server started. Restart `keystone mcp serve`, or the host session that runs it.
- **The call fails with `missing required argument`.** The caller left out a required argument. The MCP schema marks each `required: true` argument, and the `keystone tool run` line in `AGENTS.md` lists every argument.
- **The opencode tool fails to load.** The generated file imports `@opencode-ai/plugin`. Whether opencode installs that package for project tools by itself is unverified.
- **The handler works only in your shell.** Keystone runs it in the project root under `bash`, not in your current folder.

## Authoring

```
keystone new tool <id>
```

This writes `.agents/tools/<id>.md` with `transport: cli`, `run: ./scripts/<id>.sh`, and one required sample argument. Write that script, or change `run:` to point at the command you want. Then run `keystone index && keystone project`.
