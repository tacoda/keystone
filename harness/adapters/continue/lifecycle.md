# Continue — Lifecycle binding

How each abstract lifecycle action is invoked in Continue.

## Action → invocation

Continue reads `.continuerules` (legacy) and `config.yaml` / `config.json` (current) at the repo root. Lifecycle actions are invoked either by **typing the action name in chat** or via **custom slash commands** declared in `config.yaml`. Slash commands are preferred — they make the invocation explicit and let the harness ship a known set.

| Action | Invocation | What happens |
|---|---|---|
| **spec** | `/keystone:spec <task>` (custom slash) or "Start the spec phase for `<task>`." | Continue reads `harness/guides/process/spec.md` and follows its activities. Tracker fetch via MCP server (if configured) or `cmd` step (`gh issue view`). |
| **orient** | `/keystone:orient` or "Orient for work in `<region>`." | Continue reads `harness/corpus/state/CODEBASE_STATE.md` and the matching idioms; sketches a plan. |
| **check-drift** | `/keystone:check-drift` or "Check the diff for drift." | Continue compares `git diff` (via `cmd` step) against loaded guide rules. |
| **verify** | `/keystone:verify` or "Run the verify action." | Continue invokes sensors via `cmd` steps; reports results. |
| **review** | `/keystone:review` or "Run the review action." | Continue walks the diff against spec AC, then runs functional and security review concerns **sequentially** (no sub-agent parallelism). |
| **learn** | `/keystone:learn` or "Capture learnings from this work." | Continue writes a candidate to `harness/learning/inbox/<timestamp>-<slug>.md`. |
| **bootstrap** | `/keystone:bootstrap` or "Bootstrap the harness." | One-time; detects stack, frameworks, and libraries; seeds corpus (idioms/<stack>/, state/), paired guides (idioms/<stack>/); confirms sensor commands; inventories computational guides (LSPs, formatters, editor enforcement) into `guides/computational/`; classifies sensors by kind. Post-bootstrap, every applicable guide and sensor is recorded in `corpus/state/CODEBASE_STATE.md`. |
| **audit** | `/keystone:audit` or "Audit the harness." | Full Learning + Pruning flywheel pass. |
| **synthesize** | `/keystone:synthesize` or "Synthesize the inbox." | Promotes inbox items into the right corpus and/or guide. |
| **mode** | Edit `harness/guides/process/modes.md` directly. | Continue has no autonomy levers; the file is informational. |

## Suggested `config.yaml` for slash commands

```yaml
name: keystone-harness
version: 0.3.0
schema: v1

prompts:
  - name: "keystone:spec"
    description: "Spec phase — capture acceptance criteria"
    prompt: |
      Read harness/guides/process/spec.md and run the spec action
      for: {{{ input }}}
  - name: "keystone:orient"
    description: "Planning phase — orient on touched region"
    prompt: |
      Read harness/guides/process/planning.md. Orient for: {{{ input }}}
  - name: "keystone:check-drift"
    description: "Compare current diff against loaded guides"
    prompt: |
      Read harness/sensors/drift.md and run the check-drift action
      on the current diff.
  - name: "keystone:verify"
    description: "Run every sensor against the current change"
    prompt: |
      Read harness/guides/process/verification.md and run the verify
      action. Sensor commands live in harness/corpus/state/CODEBASE_STATE.md.
  - name: "keystone:review"
    description: "Walk spec AC + run review concerns over the diff"
    prompt: |
      Read harness/guides/process/review.md and run the review action.
  - name: "keystone:learn"
    description: "Capture a learning candidate"
    prompt: |
      Read harness/learning/README.md. Write a candidate to
      harness/learning/inbox/<timestamp>-<slug>.md classifying it
      as rule (→ guides) or information (→ corpus).
  - name: "keystone:bootstrap"
    description: "One-time bootstrap"
    prompt: |
      Read harness/guides/process/README.md and run the bootstrap action.
  - name: "keystone:audit"
    description: "Audit the harness (Learning + Pruning flywheels)"
    prompt: |
      Read harness/archive/README.md and run the audit action.
  - name: "keystone:synthesize"
    description: "Promote inbox items into guides and/or corpus"
    prompt: |
      Read harness/learning/README.md and run the synthesize action.
```

Continue picks these up as `/keystone:spec`, `/keystone:orient`, etc., and they appear in the slash-command picker.

## Context providers

Continue supports custom context providers in `config.yaml`. The harness benefits from two built-ins:

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

A `folder` provider scoped to `harness/guides/` keeps guides reliably in the model's reach without making the user `@`-mention them.

## Sub-agent support

None. Continue runs a single conversation. The **review** action runs each review concern sequentially in the same session.

## Modes

Continue has no autonomy levels in the harness sense. The user can toggle **agent mode** (full tool use) vs. **chat mode** (Q&A only) in the side panel. The harness's `paired`/`solo`/`autopilot` collapse to a single mode in practice.

## Capability matrix

| Capability | Supported? |
|---|---|
| Autonomous shell execution | ✓ (`cmd` step type and built-in terminal tool in agent mode) |
| Sub-agent parallelism | ✗ |
| Autonomy levels (paired/solo/autopilot) | ✗ — effectively paired |
| Lazy-by-region | partial — `folder` and `codebase` context providers reach into `harness/corpus/idioms/<stack>/` on demand, but no auto-attach by file glob |
| Context-reset primitive | new chat in the side panel; or "Clear" in the chat options menu |
| Tracker integration | via MCP server (Atlassian/Linear) if configured, or `cmd` step (`gh`, `curl`) |
| GitHub integration | partial (via `cmd` step) |
