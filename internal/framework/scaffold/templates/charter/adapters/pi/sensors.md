# pi.dev — Sensor binding

How sensors fire inside pi.

## Execution model

Pi runs shell commands directly. **All sensors run autonomously** — the agent reads tool commands from `charter/corpus/state/CODEBASE_STATE.md`, invokes them via pi's shell tool, and consumes the output. No human paste-and-report loop required.

## Per-sensor binding

| Sensor | Fires inside | Implementation |
|---|---|---|
| lint | **verify** prompt template | Shell with the project's lint command. |
| type-check | **verify** prompt template | Shell with type-check command. Skipped when no type checker. |
| test | **verify** prompt template | Shell with test command. Fresh run per turn. |
| build | **verify** prompt template | Shell with build command. |
| drift | **check-drift** + **verify** + **audit** prompt templates | Agent reads loaded corpus rules + the diff; reports findings. |
| coverage | **verify** + **audit** | Shell runs the coverage command; agent proposes a diff to `state/CODEBASE_STATE.md`. |
| risk-fingerprint | **audit** | Complexity metrics + `git log`; agent writes `state/risk-fingerprints.md` diff. |
| traffic-topology | **audit** | `git log` + `state/CODEBASE_STATE.md` criticality → `state/traffic-topology.md` diff. |
| state-region | **orient** | Agent reads `state/CODEBASE_STATE.md` and active migrations for touched paths. |
| commit-message | **release** phase, pre-commit | Agent inspects proposed message. |
| tracker-card-fetcher | **spec** prompt template | If pi has a tracker integration (extension or shell-based `gh issue view`, `jira` CLI, etc.), use it. Otherwise the human pastes the card content. |
| spec-adherence | **review** | Agent reads spec + diff and walks AC. |

## State writes

Sensors that update state files propose a diff via pi's file-edit capability. The user accepts or edits per pi's standard confirmation flow.

## Stale evidence guard

Same IRON LAW as Claude Code: claims need fresh evidence. Each sensor must re-run in the current turn after any edit. The shipped `.pi/prompts/verify.md` template enforces this by instructing pi to invoke every sensor at the top of the template body.

## Sensor failure handling

1. The **verify** template surfaces structured output from the failing sensor.
2. The agent returns to the **implementation** phase. Do not "fix forward" inside verification.
3. After the fix, re-invoke `/keystone-verify`. Each retry is a fresh evidence cycle.
