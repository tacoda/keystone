# Glossary

<!-- voice-register: informational -->
<!-- voice-english: ste -->

Keystone's vocabulary, and the one distinction it rests on.

## Charter vs. harness

Split everything around a coding agent by **one question: did you _author_ it
to specify behavior, or is it the _engine_ that applies what's specified?**

### Charter

Everything you author to specify how the agent should behave. Mostly prose,
sometimes code. **You write it.**

- `CLAUDE.md`, `AGENTS.md`, and equivalent instruction files
- rules and the references behind them
- hooks, permissions, and evals you author
- skills, commands, subagents, and tools
- policy

Keystone **manages the charter.** In a Keystone repo the charter lives at
`.agents/` and is indexed at `.keystone/INDEX.json`.

### Harness

The engine that interprets and applies the charter against the model. **You run
it; you don't author it.**

- the coding agent: Claude Code, Cursor, Codex, opencode, …
- agent frameworks: LangChain, LangGraph, CrewAI, …
- the orchestrator, the loop, the runner that fires your hooks

The lineage is the ML training-harness / eval test-harness sense: runnable
scaffolding that exercises the thing.

## The test (load-bearing)

**Authorship, not executability.** A pre-commit hook *executes* and is still
**charter**, because you authored it to state a standard. The runner that fires
that hook is the **harness**.

> Author the spec → charter. Be the engine → harness.

Say it that sharply or "charter" collapses back into "harness."

The orchestrator resolves cleanly: it *runs*, so it's **harness** (machinery);
the process it enforces is *authored*, so that's **charter**. It's the split
between mechanism and policy that puts Claude Code in the harness and your
`CLAUDE.md` in the charter.

## Keystone

The agent charter framework: a CLI, an MCP server, and a dashboard that
author, check, project, and maintain the charter. Keystone is **not** a
harness. It manages the charter that constrains whatever harness you run.

## Constraint and capability

The two halves of a charter.

- A **constraint** limits, gates, or checks what the agent does: `rule`,
  `reference`, `hook`, `permission`, `eval`.
- A **capability** adds something the agent can do: `skill`, `command`,
  `subagent`, `tool`, `document`.

## Primitive

A typed unit of the charter: one file (or, for a skill, one folder) under
`.agents/`, with frontmatter and a body. Its kind comes from its folder.

| Kind         | Category   | What it is                                                        |
| ------------ | ---------- | ----------------------------------------------------------------- |
| `rule`       | constraint | A directive. `paths:` scopes it to matching files.                |
| `reference`  | constraint | The reasoning behind a rule, opened on demand. Was `corpus`.      |
| `hook`       | constraint | A check that runs on an event and can block. Was `sensor`.        |
| `permission` | constraint | Allow, ask, and deny rules for tools. Was `posture`.              |
| `eval`       | constraint | A test that the charter changes agent behavior.                   |
| `skill`      | capability | An Agent Skills `SKILL.md`. Absorbed `playbook` and `pattern`.    |
| `command`    | capability | A unit of work the user asks for by name.                         |
| `subagent`   | capability | A role the agent delegates to. Was `agent`.                       |
| `tool`       | capability | A callable you write, served over MCP, as a command, or as a host plugin. |
| `document`   | capability | A template for governed output, such as a plan or an ADR.         |

## Agent and subagent

The **agent** is the model you talk to, running in its host. A **subagent**
is a role that agent delegates to, with its own instructions and tools.
Keystone's kind is `subagent`. I kept `agents/` for its folder because every
host with subagents names its folder that way: `.claude/agents/`,
`.cursor/agents/`, `.github/agents/`, `.opencode/agents/`.

## Host

A coding agent keystone projects to: Claude Code, Cursor, GitHub Copilot,
Continue, opencode, Codex, Cline, Goose, Pi, or Aider. `keystone.json`
`adapters` lists the ones a repo uses.

## Event

What a hook runs on, set by its `event:` field.

- A **host event** (`PreToolUse`, `Stop`, …) is one the host sees. Keystone
  bridges it into the host's hook config, and a failing hook blocks there.
- A **signal** is any other event: one keystone fires itself, like
  `pre-verify`, or one your project defines in `keystone.json` `signals:`.
  Fire it with `keystone signal fire <name>`.

## Projection

Delivering `.agents/` to a host with `keystone project`: natively where the
host already reads the path, as a link where it reads the format elsewhere,
as a translated file where it needs another format, or as an `AGENTS.md`
entry where it has no slot. Keystone records each file it writes in the
lockfile. Edit the source in `.agents/`, never the projected file.
