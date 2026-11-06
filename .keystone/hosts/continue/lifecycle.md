# Continue — Lifecycle binding

How each abstract lifecycle action is invoked in Continue.

## Invocation

Every action is invoked via natural language: "run task on TICKET-123," "run verify," "do a review pass." Continue reads `.continuerules` at session start, finds the action in the bulleted list, follows the link to `.charter/actions/<action>.md`, and executes the playbook. The charter no longer ships per-action `config.yaml` slash commands — natural-language invocation against the agent-agnostic playbooks is the single path. Users who *want* slash-command ergonomics can author their own `prompts:` entries in `config.yaml` that just say "Read `.charter/actions/<action>.md` and follow it."

The canonical kickoff phrase is **"run task on `<ticket-id>`"** (or "run the task workflow") — `.charter/actions/task.md` orchestrates `spec → orient → implementation → check-drift → verify → review`.

## Context providers

Continue supports custom context providers in `config.yaml`. The charter benefits from two built-ins:

```yaml
context:
  - provider: file
  - provider: codebase
  - provider: diff
  - provider: terminal
```

- **`diff`** — auto-attaches the current diff for **check-drift** and **review**.
- **`codebase`** — semantic search over the repo; useful when **orient** needs to find similar regions.
- **`terminal`** — surfaces recent terminal output (e.g., test failures) into the model's context.

A `folder` provider scoped to `.charter/guides/` keeps guides reliably in the model's reach without making the user `@`-mention them.

## Sub-agent support

None. Continue runs a single conversation. The **review** action runs each review concern sequentially in the same session.

## Modes

Continue has no autonomy levels in the charter sense. The user can toggle **agent mode** (full tool use) vs. **chat mode** (Q&A only) in the side panel. The charter's `paired`/`solo`/`autopilot` collapse to a single mode in practice.

## Capability matrix

| Capability | Supported? |
|---|---|
| Autonomous shell execution | ✓ (`cmd` step type and built-in terminal tool in agent mode) |
| Sub-agent parallelism | ✗ |
| Autonomy levels (paired/solo/autopilot) | ✗ — effectively paired |
| Lazy-by-region | partial — `folder` and `codebase` context providers reach into `.charter/corpus/idioms/<stack>/` on demand, but no auto-attach by file glob |
| Context-reset primitive | new chat in the side panel; or "Clear" in the chat options menu |
| Tracker integration | via MCP server (Atlassian/Linear) if configured, or `cmd` step (`gh`, `curl`) |
| GitHub integration | partial (via `cmd` step) |
