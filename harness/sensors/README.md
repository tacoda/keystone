# Sensors

Automated checks and instruments that read code or state. One of the four harness components (Corpus / Guides / **Sensors** / Flywheels).

## Two kinds

Sensors split by what they do with their output:

**Verification sensors** — gate the commit. Pass/fail. Run during the verification phase via the **verify** action. Failure blocks the commit.

**State sensors** — produce empirical updates to the State layer. Propose diffs against `corpus/state/` files; the user accepts or edits per the scaffolding safety contract.

A third group is read-only — they consult State but write nothing. The **state-region sensor** is the only one of this kind in the default set.

## How sensors fire

Sensors do not run on event hooks. They run inside lifecycle actions invoked at the right phase boundary:

| Action | Phase | Sensors that run |
|---|---|---|
| **spec** | spec | [tracker-card-fetcher](tracker-card-fetcher.md) (if a card ID was provided) |
| **orient** | planning | [state-region](state-region.md) |
| **check-drift** | implementation | [drift](drift.md) |
| **verify** | verification | [lint](lint.md), [type-check](type-check.md), [test](test.md), [build](build.md), [drift](drift.md), [commit-message](commit-message.md); proposes state updates from [coverage](coverage.md) |
| **review** | review | review-functional, review-security (agents); [spec-adherence](spec-adherence.md) check |
| **audit** | discipline | [drift](drift.md), [coverage](coverage.md), [risk-fingerprint](risk-fingerprint.md), [traffic-topology](traffic-topology.md) |

How an action is actually invoked, and whether sensors run autonomously or require human prompting, is agent-specific — see `harness/adapters/<your-agent>/sensors.md`. The most common degradation: an agent that cannot run shell commands during a turn surfaces the sensor commands for the human to execute instead.

## Contract shape

Each sensor declares:

- **Trigger** — when it runs.
- **Inputs** — what it reads.
- **Exit condition** — what counts as pass.
- **Output** — pass/fail + structured findings.
- **State writes** — which `corpus/state/` file(s), if any. State writes follow the scaffolding safety contract: propose a diff, never silently overwrite.

Tool commands (lint, test, build, type-check, coverage) live in `corpus/state/CODEBASE_STATE.md`, populated by the **bootstrap** action. Sensors read them from there.

## Sensor index

| Sensor | Purpose |
|---|---|
| [lint](lint.md) | Surface-level style and pattern checks |
| [type-check](type-check.md) | Signature and contract consistency |
| [test](test.md) | The project's test suite |
| [build](build.md) | Build / compile / package step |
| [drift](drift.md) | Diff vs. loaded corpus and guide rules |
| [coverage](coverage.md) | Test coverage; proposes state updates |
| [risk-fingerprint](risk-fingerprint.md) | Complexity + coupling + coverage per region |
| [traffic-topology](traffic-topology.md) | Git churn + recency + criticality map |
| [state-region](state-region.md) | Read-only: what State says about a touched region |
| [commit-message](commit-message.md) | Conventional-commit format + no AI attribution |
| [tracker-card-fetcher](tracker-card-fetcher.md) | Fetches tracker card from Jira / Linear / etc. |
| [spec-adherence](spec-adherence.md) | Walks spec ACs against the diff |

## How contracts are wired

Sensors are documented here; they are implemented inside the lifecycle action that invokes them. The **verify** action knows how to run lint and test sensors; the **audit** action knows how to run risk-fingerprint and traffic-topology. Each action reads tool commands from `corpus/state/CODEBASE_STATE.md` (populated by the **bootstrap** action).

The actual wiring — what command, file, or rule trigger embodies each action for a given agent — lives in `harness/adapters/<your-agent>/sensors.md`.

## Adding a sensor

1. Define its contract in a new file using the section shape above.
2. Update the relevant lifecycle action(s) to invoke it, in each adapter where the action exists.
3. If it writes state, ensure the action honors the scaffolding safety contract — propose a diff, do not overwrite.

## Removing a sensor

1. Archive its file (move to `archive/sensors/<name>.md` with an `archived_at` header).
2. Update each adapter to stop invoking it.
3. If the sensor wrote to State, decide whether the historical data stays or moves to `corpus/state/archive/`.

## Anti-patterns

- A sensor that silently overwrites State.
- A sensor whose "pass" condition is "ran without error" — passing should mean *useful evidence was produced*.
- A sensor that runs on every conversational turn. Sensors fire at phase boundaries, not as side effects of conversation.
- Two sensors that overlap (e.g., two "test" sensors). One concern, one sensor.
