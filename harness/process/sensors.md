# Sensors

Automated checks and instruments that read code or state. One of the four roles (Guides / **Sensors** / Flywheels / Discipline).

## Two kinds

Sensors split by what they do with their output:

**Verification sensors** — gate the commit. Pass/fail. Run during the verification phase via the **verify** action. Failure blocks the commit.

**State sensors** — produce empirical updates to the State layer. Propose diffs against `state/` files; the user accepts or edits per the scaffolding safety contract.

A third group is read-only — they consult State but write nothing. The **state-region sensor** is the only one of this kind in the default set.

## How sensors fire

Sensors do not run on event hooks. They run inside lifecycle actions invoked at the right phase boundary:

| Action | Phase | Sensors that run |
|---|---|---|
| **spec** | spec | tracker-card-fetcher (if a card ID was provided) |
| **orient** | planning | state-region |
| **check-drift** | implementation | drift |
| **verify** | verification | lint, type-check, test, build, drift, commit-message; proposes state updates from coverage |
| **review** | review | review-functional, review-security (agents); spec-adherence check |
| **audit** | discipline | drift, coverage, risk-fingerprint, traffic-topology |

How an action is actually invoked, and whether sensors run autonomously or require human prompting, is agent-specific — see `harness/adapters/<your-agent>/sensors.md`. The most common degradation: an agent that cannot run shell commands during a turn surfaces the sensor commands for the human to execute instead.

## Contract shape

Each sensor declares:

- **Trigger** — when it runs.
- **Inputs** — what it reads.
- **Exit condition** — what counts as pass.
- **Output** — pass/fail + structured findings.
- **State writes** — which `state/` file(s), if any. State writes follow the scaffolding safety contract: propose a diff, never silently overwrite.

Tool commands (lint, test, build, type-check, coverage) live in `state/CODEBASE_STATE.md`, populated by the **bootstrap** action. Sensors read them from there.

---

## Sensor: lint

Surface-level style and pattern checks.

- **Trigger** — implementation phase (continuous, fast feedback) and verification phase (gate).
- **Inputs** — the project's lint command from `state/CODEBASE_STATE.md`.
- **Exit condition** — exit code 0; no errors.
- **Output** — pass/fail. On fail: the linter's structured output passed through.
- **State writes** — none.

## Sensor: type-check

Signature and contract consistency.

- **Trigger** — implementation phase (continuous), verification phase (gate).
- **Inputs** — the project's type-check command from `state/CODEBASE_STATE.md`. Skipped if the project has no type checker.
- **Exit condition** — exit code 0; no type errors.
- **Output** — pass/fail. On fail: errors as the type checker emits them.
- **State writes** — none.

## Sensor: test

The project's test suite.

- **Trigger** — implementation phase (after each green during TDD), verification phase (gate).
- **Inputs** — the project's test command from `state/CODEBASE_STATE.md`.
- **Exit condition** — exit code 0; 0 failures.
- **Output** — pass/fail with failure summary. Stale evidence does not count — see `verification.md`'s IRON LAW.
- **State writes** — none directly; the coverage sensor reads test artifacts.

## Sensor: build

The project's build / compile / package step.

- **Trigger** — verification phase (gate).
- **Inputs** — the project's build command from `state/CODEBASE_STATE.md`.
- **Exit condition** — exit code 0; artifacts produced where expected.
- **Output** — pass/fail.
- **State writes** — none.

## Sensor: drift

Compares the diff (or full codebase, during audit) against loaded corpus rules. Finds violations of IRON LAWs and GOLDEN RULES.

- **Trigger** — **check-drift** (implementation), **verify** (verification), **audit** (discipline).
- **Inputs** — current diff (or file set for audit), loaded corpus rules, the file paths each rule applies to.
- **Exit condition** — no IRON LAW violations.
- **Output** — pass/fail. Findings list each violation with rule reference, file, and reason. GOLDEN RULE violations surface as warnings, do not fail.
- **State writes** — discrepancies discovered during the **audit** action may become Pruning flywheel candidates (archive proposals). Promotions go through the audit flow, not silent writes.

## Sensor: coverage

Reads test coverage and updates the State layer.

- **Trigger** — verification phase (proposes state update), **audit**.
- **Inputs** — the project's coverage command from `state/CODEBASE_STATE.md`. Skipped if no coverage tool is configured.
- **Exit condition** — coverage report produced. No minimum threshold by default — projects may set one in `state/CODEBASE_STATE.md`.
- **Output** — coverage stats per region.
- **State writes** — proposes a diff to `state/CODEBASE_STATE.md` updating coverage per region. User accepts or edits.

## Sensor: risk-fingerprint

Computes complexity + coupling + coverage patterns per region.

- **Trigger** — **audit**, manually via **bootstrap**.
- **Inputs** — code metrics (cyclomatic complexity, fan-in/fan-out, churn from git), coverage data.
- **Exit condition** — fingerprints computed for all tracked regions.
- **Output** — risk score per region with the metric breakdown.
- **State writes** — proposes a diff to `state/risk-fingerprints.md`. User accepts or edits.

## Sensor: traffic-topology

Combines git churn + recency + business criticality into a map of where attention concentrates.

- **Trigger** — **audit**, manually via **bootstrap**.
- **Inputs** — git log per region (last N months), business criticality flags from `state/CODEBASE_STATE.md` if the consumer has marked any.
- **Exit condition** — topology computed for all tracked regions.
- **Output** — heat map: per-region churn count, last-touched date, criticality tag.
- **State writes** — proposes a diff to `state/traffic-topology.md`. User accepts or edits.

## Sensor: state-region

Read-only. Surfaces what is already in State for a touched region.

- **Trigger** — **orient** (planning).
- **Inputs** — current task's touched paths, `state/CODEBASE_STATE.md`, active migrations in `state/migrations/active/`.
- **Exit condition** — always succeeds; output is informational.
- **Output** — for the touched region: which idioms are loaded, coverage, last reconcile date, active migrations affecting this region.
- **State writes** — none.

## Sensor: commit-message

Validates conventional-commit format and absence of AI attribution.

- **Trigger** — release phase (final gate before `git commit`).
- **Inputs** — the staged commit message.
- **Exit condition** — message matches `<type>(<scope>): <subject>`, title under 70 chars, no mention of Claude / AI agents / co-authors / tool attribution.
- **Output** — pass/fail. On fail: the violated rule and a suggested fix.
- **State writes** — none.

## Sensor: tracker-card-fetcher

Fetches a tracker card from Jira / Linear / GitHub Issues / Asana via whatever tracker integration the agent has (MCP server, CLI, plugin), if one is referenced.

- **Trigger** — **spec** (when an ID or URL is provided), occasionally **learn** and **review** to surface card metadata in artifacts.
- **Inputs** — a card identifier (e.g., `PROJ-123`, a Linear URL, a GitHub Issue URL); the corresponding tracker integration.
- **Exit condition** — card fetched, or "card not reachable" message produced if the integration is offline or the user lacks access.
- **Output** — title, description, acceptance criteria (if present), labels, links. The agent never edits the card unless the user explicitly asks.
- **State writes** — none. Card metadata lands in `docs/specs/<file>.md` instead.

## Sensor: spec-adherence

Walks the spec's acceptance criteria against the current diff.

- **Trigger** — **review** (review phase).
- **Inputs** — the spec (`docs/specs/<file>.md`) and the diff.
- **Exit condition** — every criterion is met *with evidence* (a test, an output, a manual check).
- **Output** — per-criterion pass/fail with evidence link. A criterion missing evidence fails.
- **State writes** — none.

---

## How contracts are wired

Sensors are documented here; they are implemented inside the lifecycle action that invokes them. The **verify** action knows how to run lint and test sensors; the **audit** action knows how to run risk-fingerprint and traffic-topology. Each action reads tool commands from `state/CODEBASE_STATE.md` (populated by the **bootstrap** action).

The actual wiring — what command, file, or rule trigger embodies each action for a given agent — lives in `harness/adapters/<your-agent>/sensors.md`.

## Adding a sensor

1. Define its contract in this file using the section shape above.
2. Update the relevant lifecycle action(s) to invoke it, in each adapter where the action exists.
3. If it writes state, ensure the action honors the scaffolding safety contract — propose a diff, do not overwrite.

## Removing a sensor

1. Archive its contract section (move to `archive/process/sensors-<name>.md` with an `archived_at` header).
2. Update each adapter to stop invoking it.
3. If the sensor wrote to State, decide whether the historical data stays or moves to `state/archive/`.

## Anti-patterns

- A sensor that silently overwrites State.
- A sensor whose "pass" condition is "ran without error" — passing should mean *useful evidence was produced*.
- A sensor that runs on every conversational turn. Sensors fire at phase boundaries, not as side effects of conversation.
- Two sensors that overlap (e.g., two "test" sensors). One concern, one sensor.
