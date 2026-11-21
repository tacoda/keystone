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

The agent context framework: a CLI, an MCP server, and a dashboard that
author, check, project, and maintain the charter. The **charter is the
codified context of the repository** — everything a coding agent should be
able to look up about how the codebase works and how to work in it.
Keystone is **not** a harness: it manages the context; it never runs the
model, serves author-defined callables, or acts as an engine. What must
execute deterministically belongs to hosts and CI, which the charter can
name but never replace. Codified context is in scope; ambient context
(MCP servers, CLIs, APIs, external services) is out.

## Primitive

A typed unit of the charter: one file (or, for a skill, one folder) under
`.agents/`, with frontmatter and a body. Its kind comes from its folder.

| Kind         | What it is                                                        |
| ------------ | ----------------------------------------------------------------- |
| `rule`       | A directive. `paths:` scopes it to matching files.                |
| `reference`  | The reasoning behind a rule, opened on demand. Was `corpus`.      |
| `hook`       | A check that runs on an event and can block. Was `sensor`.        |
| `permission` | Allow, ask, and deny rules for tools. Was `posture`.              |
| `eval`       | A test that the charter changes agent behavior.                   |
| `skill`      | An Agent Skills `SKILL.md`. Absorbed `playbook` and `pattern`.    |
| `command`    | A unit of work the user asks for by name.                         |
| `agent`      | A role the agent delegates to — a subagent. 5.0 spelled the kind `subagent`; 6.0 restored `agent`. |
| `document`   | A template for governed output, such as a plan or an ADR.         |

The `tool` kind is retired in 6.0: keystone manages context, not callables.

## Agent and subagent

The **agent** (the model) is what you talk to, running in its host. A
**subagent** is a role that agent delegates to, with its own instructions
and tools. Keystone's kind is `agent` (5.0 spelled it `subagent`). The
folder stayed `agents/` throughout, because every host with subagents names
its folder that way: `.claude/agents/`, `.cursor/agents/`,
`.github/agents/`, `.opencode/agents/`.

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
