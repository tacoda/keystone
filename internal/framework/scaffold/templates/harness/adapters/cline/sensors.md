# Cline / Roo Code — Sensor binding

How sensors actually fire inside Cline.

## Execution model

Cline has a built-in **execute_command** tool that runs shell commands in the workspace. By default, Cline prompts for approval before each command; the user can auto-approve broad categories (read-only ops, specific commands, command prefixes) via the extension settings.

**Sensors run autonomously** in agent mode, gated only by the user's auto-approve policy. The conversation transcript records each invocation and output, so the **verify** action's "fresh evidence" requirement is naturally satisfied.

## Per-sensor binding

Sensor commands come from `harness/corpus/state/CODEBASE_STATE.md`, populated by the **bootstrap** action.

| Sensor | Fires inside | Implementation |
|---|---|---|
| [lint](../../sensors/lint.md) | **verify** action | `execute_command` running `<lint command>` |
| [type-check](../../sensors/type-check.md) | **verify** action | `execute_command` running `<type-check command>`. Skipped when no type checker. |
| [test](../../sensors/test.md) | **verify** action | `execute_command` running `<test command>`. Fresh run per turn. |
| [build](../../sensors/build.md) | **verify** action | `execute_command` running `<build command>` |
| [drift](../../sensors/drift.md) | **check-drift** + **verify** + **audit** | `execute_command` for `git diff`; agent reads loaded guides and compares. No state writes. |
| [coverage](../../sensors/coverage.md) | **verify** + **audit** | `execute_command` running `<coverage command>`; agent proposes a diff to `corpus/state/CODEBASE_STATE.md` via the file-edit tool. |
| [risk-fingerprint](../../sensors/risk-fingerprint.md) | **audit** | `execute_command` for complexity metrics + `git log --stat`; agent writes table to `corpus/state/risk-fingerprints.md`. |
| [traffic-topology](../../sensors/traffic-topology.md) | **audit** | `git log` + criticality from `corpus/state/CODEBASE_STATE.md` → `corpus/state/traffic-topology.md`. |
| [state-region](../../sensors/state-region.md) | **orient** action | Agent reads `corpus/state/CODEBASE_STATE.md` and active migrations via `read_file`. No shell needed. |
| [commit-message](../../sensors/commit-message.md) | **release** phase, pre-commit | Agent inspects the proposed message; runs `git commit` via `execute_command` only after the message passes. |
| [tracker-card-fetcher](../../sensors/tracker-card-fetcher.md) | **spec** action | MCP tracker server (if configured) — Atlassian, Linear, GitHub — or `execute_command` with `gh issue view`. Fallback: user pastes. |
| [spec-adherence](../../sensors/spec-adherence.md) | **review** action | Agent reads spec + diff and walks AC. No shell needed. |

## State writes

Sensors that update state files propose edits via Cline's standard `write_to_file` / `replace_in_file` tools. Cline shows the diff in the editor; the user approves or modifies the change before it lands.

The scaffolding safety contract (no silent overwrites) maps cleanly onto Cline's diff-then-approve flow.

## Auto-approve recommendations

To avoid the verify cycle becoming a clicker game, auto-approve in Cline's settings:

- **Read files** — needed for the agent to walk guides, corpus, and the diff.
- **Execute safe commands** — whitelist `npm test`, `pytest`, `<lint command>`, `git diff`, `git log`, `git status`. (Cline supports per-command and per-prefix auto-approve.)
- **MCP server use** — if the user wants tracker fetches to happen without an approval prompt.

Keep **write file** and **execute unsafe commands** on manual approval — that is where review pressure belongs.

## Stale evidence guard

The **verify** action's IRON LAW requires sensors to have run *this turn*. Cline's transcript shows every `execute_command` invocation with its output; the agent reports completion only after each sensor passes in the current task. When in doubt, ask Cline to "re-run the sensors" — fresh invocations.

## Sensor failure handling

When a sensor fails:

1. The **verify** action surfaces the structured output (linter findings, test failures) from the `execute_command` result.
2. The agent returns to the **implementation** phase. Do not "fix forward" inside verify.
3. After the fix, re-invoke the failing sensor. Fresh `execute_command` = fresh evidence.

## A useful pattern: consolidated verify command

Several sensors per verify cycle means several approval prompts (unless auto-approved). A `scripts/verify.sh` that runs every sensor and exits non-zero on any failure cuts the cycle to one `execute_command`. Add the script to the auto-approve whitelist; verify becomes one click (or zero, in solo mode).

## Differences from Claude Code

- Cline has **no sub-agent parallelism** — review runs sequentially (subtasks are sequential).
- Cline's **per-command approval** is the friction point; whitelist sensor commands to keep the cycle fast.
- Cline's **diff-then-approve** edit flow is a feature for state writes — it satisfies the scaffolding safety contract without the agent doing anything special.
- MCP support is **first-class**; tracker integration is real when an MCP server is configured.
