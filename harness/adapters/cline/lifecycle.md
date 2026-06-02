# Cline / Roo Code — Lifecycle binding

How each abstract lifecycle action is invoked in Cline (and its Roo Code fork).

## Action → invocation

Cline and Roo Code are VS Code extensions that present a chat-driven agent with autonomous tool use. They read **two** harness pointers:

- The **custom instructions** field configured in the extension settings (`cline.customInstructions` / `rooCode.customInstructions`). This is the always-loaded menu.
- **`.clinerules`** (Cline) or **`.roorules`** (Roo Code) at the repo root — auto-loaded from the workspace. Newer Cline versions read `.clinerules` natively; older ones rely entirely on the settings field. The harness ships both.

Lifecycle actions are invoked by **asking in chat**. Cline / Roo Code have no slash-command primitive in the sense Claude Code does, but Cline supports **custom workflows** (saved prompts triggered from the side panel) that the harness uses to formalize lifecycle invocation.

| Action | Invocation | What happens |
|---|---|---|
| **spec** | "Start the spec phase for `<task>`." (or saved workflow `Keystone: spec`) | Cline reads `harness/guides/process/spec.md` and follows its activities. Tracker fetch via MCP server (Cline has built-in MCP support) or shell. |
| **orient** | "Orient for work in `<region>`." | Cline reads `harness/corpus/state/CODEBASE_STATE.md` and the matching idioms; sketches a plan. |
| **check-drift** | "Check the diff for drift." | Cline runs `git diff` and compares against loaded guide rules. |
| **verify** | "Run the verify action." | Cline executes sensor commands directly via its shell tool. |
| **review** | "Run the review action." | Cline walks the diff against spec AC, then runs functional and security review concerns **sequentially**. |
| **learn** | "Capture learnings from this work." | Cline writes a candidate to `harness/learning/inbox/<timestamp>-<slug>.md`. |
| **bootstrap** | "Bootstrap the harness." | One-time; seeds corpus (idioms/<stack>/, state/), paired guides (idioms/<stack>/), and confirms sensor commands. |
| **audit** | "Audit the harness." | Full Learning + Pruning flywheel pass. |
| **synthesize** | "Synthesize the inbox." | Promotes inbox items into the right corpus and/or guide. |
| **mode** | Edit `harness/guides/process/modes.md` directly. | Cline's autonomy is configured via auto-approve toggles, not the harness mode file; see below. |

## Cline's Plan / Act modes vs. harness modes

Cline ships a **Plan mode** (chat-only, no tool use) and an **Act mode** (full tool use). These are orthogonal to the harness's `paired`/`solo`/`autopilot` pacing:

| Harness mode | Suggested Cline toggle set |
|---|---|
| **paired** | Plan mode for planning/spec/review; Act mode for implementation. Auto-approve everything OFF. |
| **solo** | Act mode throughout. Auto-approve read-only operations + file edits; manual approval for shell commands. |
| **autopilot** | Act mode throughout. Auto-approve everything in the project's `.clineignore` whitelist. |

The user toggles auto-approve in the extension UI per category (read files, edit files, execute commands, use browser, use MCP). The harness recommends auto-approving file reads and `.git`-scoped commands at minimum so sensors and drift checks don't stall on approval prompts.

## Saved workflows

Cline supports **workflows** — saved prompts you trigger from the side panel. The harness defines one workflow per lifecycle action so users can fire actions consistently:

```
Keystone: spec        → "Read harness/guides/process/spec.md and run the spec action."
Keystone: orient      → "Read harness/guides/process/planning.md and orient."
Keystone: check-drift → "Read harness/sensors/drift.md and run the check-drift action."
Keystone: verify      → "Read harness/guides/process/verification.md and run verify."
Keystone: review      → "Read harness/guides/process/review.md and run review."
Keystone: learn       → "Read harness/learning/README.md and capture a learning candidate."
Keystone: bootstrap   → "Bootstrap the keystone harness."
Keystone: audit       → "Run the audit action."
Keystone: synthesize  → "Run the synthesize action."
```

The installer suggests these in `cline-instructions.md` (which is paste-into-settings text); users add them via the workflows menu.

## Sub-agent support

None. Cline runs a single agent loop per task. The **review** action runs each review concern sequentially in the same task.

Cline does support **subtasks** (a parent task spawning a focused child task), which can approximate review parallelism by running each concern as its own subtask — but the subtasks run sequentially, not in parallel.

## Capability matrix

| Capability | Supported? |
|---|---|
| Autonomous shell execution | ✓ (per-command approval, or auto-approve if enabled) |
| Sub-agent parallelism | ✗ — subtasks are sequential |
| Autonomy levels | ✓ (auto-approve toggles map roughly to harness modes) |
| Lazy-by-region | ✗ — Cline reads files the model decides to read; no glob-based auto-attach |
| Context-reset primitive | New Task button in the side panel |
| Tracker integration | ✓ (built-in MCP support; configure an Atlassian / Linear / GitHub server) |
| GitHub integration | ✓ (via MCP server or shell `gh`) |
| Roo Code fork compatibility | ✓ (same binding shape; substitute `.roorules` for `.clinerules` and `rooCode.customInstructions` for the settings key) |
