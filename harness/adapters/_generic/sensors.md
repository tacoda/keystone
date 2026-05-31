# Generic — Sensor binding

How sensors fire when the agent cannot run shell commands directly.

## Execution model

The agent surfaces the sensor command; the human runs it; the human pastes the output back. The agent reads the output and proceeds.

This is the slowest and most error-prone path. Every adapter that *can* run sensors autonomously should do so. Use this only when the agent has no shell execution capability.

## Per-sensor binding

| Sensor | Behavior |
|---|---|
| lint | Agent says: "Please run `<lint command from CODEBASE_STATE.md>` and paste the output." |
| type-check | Same pattern. Skipped if no type checker is configured. |
| test | Same pattern. Stale evidence does not count — the human must re-run after any edit. |
| build | Same pattern. |
| drift | Agent reads loaded corpus rules + diff; runs the comparison itself; reports findings. (No shell needed.) |
| coverage | Agent asks the human to run the coverage command and paste the report; agent proposes a diff to `state/CODEBASE_STATE.md`. |
| risk-fingerprint | Agent asks for complexity metrics + `git log` output if it can't read them itself; otherwise computes from what it has access to. |
| traffic-topology | Same — requires `git log` output. |
| state-region | Agent reads `state/CODEBASE_STATE.md` directly. No shell needed. |
| commit-message | Agent inspects proposed message before the human commits. No shell needed. |
| tracker-card-fetcher | If the agent has no tracker integration, the human pastes the card content. |
| spec-adherence | Agent reads the spec + diff and walks AC. No shell needed. |

## Stale evidence guard

The same IRON LAW applies: claims need fresh evidence. In the surface-and-paste model, "fresh" means the human ran the command **after the most recent edit and before the claim**. The agent must explicitly ask for the re-run after every edit cycle — there is no automation to catch a forgotten step.

## When this hurts most

The **verify** action takes 4-6 sensor runs per cycle. In the surface-and-paste model, that's 4-6 round-trips of "agent asks, human runs, human pastes." For projects with fast sensors (< 30s total), this is workable but tiring. For slow sensors (minutes), the human-in-the-loop cost dominates.

Recommendation: if you're stuck in this model, batch the sensors into a single script (`scripts/verify.sh`) and have the human run that one command per cycle.
