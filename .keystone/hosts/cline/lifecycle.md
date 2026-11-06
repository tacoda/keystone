# Cline / Roo Code — Lifecycle binding

How each abstract lifecycle action is invoked in Cline (and its Roo Code fork).

## Invocation

Cline and Roo Code are VS Code extensions that present a chat-driven agent with autonomous tool use. They read **two** charter pointers:

- The **custom instructions** field configured in the extension settings (`cline.customInstructions` / `rooCode.customInstructions`). This is the always-loaded menu.
- **`.clinerules`** (Cline) or **`.roorules`** (Roo Code) at the repo root — auto-loaded from the workspace. Newer Cline versions read `.clinerules` natively; older ones rely entirely on the settings field. The charter ships both.

Every action is invoked via natural language: "run task on TICKET-123," "run verify," "do a review pass." The agent finds the action in the menu's bulleted list, follows the link to `.charter/actions/<action>.md`, and executes the playbook. Users who want a one-click trigger can save the canonical phrase as a Cline workflow (saved prompts triggered from the side panel).

The canonical kickoff phrase is **"run task on `<ticket-id>`"** (or "run the task workflow") — `.charter/actions/task.md` orchestrates `spec → orient → implementation → check-drift → verify → review`.

## Cline's Plan / Act modes vs. charter modes

Cline ships a **Plan mode** (chat-only, no tool use) and an **Act mode** (full tool use). These are orthogonal to the charter's `paired`/`solo`/`autopilot` pacing:

| Charter mode | Suggested Cline toggle set |
|---|---|
| **paired** | Plan mode for planning/spec/review; Act mode for implementation. Auto-approve everything OFF. |
| **solo** | Act mode throughout. Auto-approve read-only operations + file edits; manual approval for shell commands. |
| **autopilot** | Act mode throughout. Auto-approve everything in the project's `.clineignore` whitelist. |

The user toggles auto-approve in the extension UI per category (read files, edit files, execute commands, use browser, use MCP). The charter recommends auto-approving file reads and `.git`-scoped commands at minimum so sensors and drift checks don't stall on approval prompts.

## Optional: saved workflows for one-click invocation

Cline supports **workflows** — saved prompts you trigger from the side panel. Power users who want per-action shortcuts can define one workflow per action, each invoking the canonical playbook:

```
Keystone: task        → "Run task on $INPUT"
Keystone: bootstrap   → "Run bootstrap"
Keystone: verify      → "Run verify"
Keystone: review      → "Run review"
```

Workflows aren't required — every action is reachable via natural language against the menu file alone. They're a UX convenience for repeated invocation.

## Sub-agent support

None. Cline runs a single agent loop per task. The **review** action runs each review concern sequentially in the same task.

Cline does support **subtasks** (a parent task spawning a focused child task), which can approximate review parallelism by running each concern as its own subtask — but the subtasks run sequentially, not in parallel.

## Capability matrix

| Capability | Supported? |
|---|---|
| Autonomous shell execution | ✓ (per-command approval, or auto-approve if enabled) |
| Sub-agent parallelism | ✗ — subtasks are sequential |
| Autonomy levels | ✓ (auto-approve toggles map roughly to charter modes) |
| Lazy-by-region | ✗ — Cline reads files the model decides to read; no glob-based auto-attach |
| Context-reset primitive | New Task button in the side panel |
| Tracker integration | ✓ (built-in MCP support; configure an Atlassian / Linear / GitHub server) |
| GitHub integration | ✓ (via MCP server or shell `gh`) |
| Roo Code fork compatibility | ✓ (same binding shape; substitute `.roorules` for `.clinerules` and `rooCode.customInstructions` for the settings key) |
