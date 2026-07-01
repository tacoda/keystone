# Glossary

Keystone's vocabulary, and the one distinction it rests on.

## Charter vs. harness

Split everything around a coding agent by **one question: did you _author_ it
to specify behavior, or is it the _engine_ that applies what's specified?**

### Charter

Everything you author to specify how the agent should behave. Mostly prose,
sometimes code. **You write it.**

- `CLAUDE.md`, `AGENTS.md`, and equivalent instruction files
- rules / guides, corpus, patterns
- sensors, hooks, and gates you author
- skills, playbooks, personas
- policy

Keystone **manages the charter.** In a Keystone repo the charter lives at
`.charter/` and is indexed at `.charter/INDEX.json`.

### Harness

The engine that interprets and applies the charter against the model. **You run
it; you don't author it.**

- the coding agent — Claude Code, Cursor, Codex, opencode, …
- agent frameworks — LangChain, LangGraph, CrewAI, …
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
the process / SOP it enforces is *authored*, so that's **charter**. Mechanism
vs. policy — the same split that puts Claude Code in the harness but your
`CLAUDE.md` in the charter.

## Keystone

The agent charter framework: a CLI + MCP server + dashboard that authors,
validates, projects, and maintains the charter. Keystone is **not** a harness —
it manages the charter that constrains whatever harness you run.

## Primitive

A typed unit of the charter — one of 13 kinds (`guide`, `sensor`,
`agent`, `command`, `skill`, `playbook`, `pattern`, `corpus`, `document`,
`concern`, `posture`, `tool`, `eval`). Each carries canonical frontmatter and a
canonical path under `.charter/`.

## Signal

A keystone **framework event** — the extensible, higher-level counterpart
to a host hook phase. A primitive subscribes to one via `on:` (like a
skill declares `triggers:`); when the signal fires, the subscriber reacts:

- **sensor** — a check → verdict (exit/HTTP status); gates.
- **tool** — an external callable; on-demand, or a side-effect with `on:`.
- **agent** — an inferential review → structured `returns:`.

Host phases (`PreToolUse`, `Stop`, …) are a closed set bridged into the
host; *any other* `on:` value is a signal, so projects define their own
(`keystone.json signals:`, `keystone signal fire|list`). The `hook`
primitive kind is retired — reactions self-subscribe.

## Projection

Rendering the charter into a host agent's native paths (`.claude/rules/`,
`.claude/skills/`, `.cursor/rules/`, `AGENTS.md`, …) via `keystone project`.
Canonical source is `.charter/`; projected files are generated — don't
hand-edit them.
