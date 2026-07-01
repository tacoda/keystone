# pi.dev — Lifecycle binding

How pi.dev runs the charter's lifecycle actions.

## Invocation

Every action is invoked via natural language: "run task on TICKET-123," "run verify," "do a review pass." The agent reads `AGENTS.md` at session start, finds the action in the bulleted list, follows the link to `charter/actions/<action>.md`, and executes the playbook. No prompt templates, no `/keystone-<action>` shortcuts — the playbooks are agent-agnostic.

The canonical kickoff phrase is **"run task on `<ticket-id>`"** (or "run the task workflow") — `charter/actions/task.md` orchestrates `spec → orient → implementation → check-drift → verify → review`.

## Sensor execution

Pi runs shell commands directly. Sensors execute autonomously — see `charter/adapters/pi/sensors.md` for the per-sensor binding.

## Sub-agent degradation

The **review** action would ideally spawn `review-functional`, `review-security`, `review-risk`, and `review-deployment` in parallel. Pi doesn't support sub-agent parallelism natively. Two workable approaches:

1. **Sequential** (default) — pi runs each reviewer pass in turn against the same diff and combines findings.
2. **tmux fan-out** — power users can launch parallel pi instances in tmux panes, each running one reviewer pass, then merge results.

`charter/actions/review.md` defaults to sequential.

## Session and branching

Pi's tree-structured session history with branching fits the `spec → orient → implementation → verification → review` flow well: each phase can be a branch point, and the user can return to a prior branch if a downstream phase reveals the plan was wrong.

## Capability matrix

| Capability | Supported? |
|---|---|
| Autonomous shell execution | ✓ (pi runs shell commands directly) |
| Sub-agent parallelism | ✗ built-in (workaround: spawn pi instances via tmux, or extensions) |
| Autonomy levels (paired/solo/autopilot) | partial — pi doesn't model autonomy, but the corpus' `process/modes.md` is still readable; the agent reads it and adjusts behavior |
| Lazy-by-region | ✗ (no glob-based loading; **orient** reads `state/CODEBASE_STATE.md` and loads matching idioms by hand) |
| Context-reset primitive | `/compact` and session branching |
| Project-scoped settings | ✓ (`.pi/settings.json`) |
