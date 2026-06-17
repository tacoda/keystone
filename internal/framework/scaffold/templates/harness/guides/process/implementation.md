---
kind: guide
id: process/implementation
description: 'The phase where code gets written, test-first.'
---
# Implementation

The phase where code gets written, test-first.

## Entry condition

An approved plan exists (`docs/plans/...`). The verification phase has not yet started.

## Activities

### 1. TDD loop

Red → Green → Refactor.

- **Red** — write failing tests first. Present the list of test descriptions for review before writing assertions. The smallest test that proves the new behavior.
- **Green** — implement the minimum to make the test pass. It is fine if one change resolves several failing tests; only make the smallest change for the fewest greens.
- **Refactor** — only after green. Do not change behavior; the tests passing before refactor must pass after.

### 2. Catch TDD anti-patterns

During the loop, watch for and stop on:

- **Tests that match the implementation, not the behavior.** A test that breaks every time the function name changes is testing structure, not output.
- **Mocking the unit under test.** If half the test setup is mocks of the code being tested, the test is theater.
- **Tests that pass without being run.** A test that does not fail when the implementation is removed is not a test.
- **Tests with no assertions.** An assertion-free test is a smoke test at best.
- **One test, many concerns.** Each test should have one logical assertion. Arrange-Act-Assert with one act.
- **Mocking your own modules.** Mock unmanaged dependencies — external APIs, time, the filesystem. Do not mock the code under test or its collaborators unless legacy forces it.

### 3. Task shape — bug

- Reproduce the bug first, in code. The reproduction is a failing test.
- Diagnose the *root cause*, not the surface symptom. A fix that suppresses the symptom while leaving the cause is a regression waiting to happen.
- Wrap the fix in the regression test that previously failed.

### 4. Task shape — refactor

- Do not change behavior in a refactor commit. Tests pass before; tests pass after.
- Rule of three — refactor on the third duplication, not the first.
- Smell-driven catalog of small steps; tests after each. Long methods, large classes, long parameter lists, duplicated code, dead code, shotgun surgery, feature envy, speculative generality.
- Tidy first — small structural improvements before adding features. Never tidy and add behavior in the same commit.

### 5. Task shape — performance work

- Pick a metric. Without one, "improvement" is hand-waving.
- Baseline first. Numbers before any change.
- One change at a time. Measure delta. If the delta is within noise, the change is not a perf improvement.
- No performance claims without before/after numbers.

### 6. Multi-task plans

If the plan has 3+ tasks, dispatch each in a fresh subagent (if your agent supports subagents; sequential if not). Two-stage review per task: first spec compliance, then code quality. The subagent's "done" report is not evidence — check the diff yourself.

### 7. Check drift

After meaningful edits, invoke the **check-drift** action. It compares the changes against loaded principles, idioms, and domain invariants. Findings are surfaced, not blocking; the verification phase is where they harden.

## Sensors

Implementation runs sensors continuously, not just at exit:

- **Lint** — catches surface-level violations early.
- **Type-check** — catches signature drift as it happens.
- **Test runner** — runs after each green to confirm the TDD loop is still on the rails.
- **Drift sensor** — invoked by the **check-drift** action.

These sensors do not gate the phase by themselves; the verification phase is where the gate hardens. They run during implementation as a fast feedback loop, *when the agent can drive them autonomously*. Agents that can only suggest commands fall back to surfacing the command for the human to run — see `harness/adapters/<your-agent>/sensors.md`.

## Gate condition

To exit implementation:

1. All tests written for the change pass.
2. The change matches the approved plan. Deviations are recorded in the plan as inline updates.
3. The **check-drift** action ran in this turn with no IRON LAW violations.

The verification phase will run the harder gates.

## Artifacts

| Kind | Location |
|---|---|
| Working code | Wherever the change belongs |
| Plan updates | `docs/plans/YYYY-MM-DD-<topic>.md` (inline edits as deviations occur) |

## Anti-patterns

- Writing the implementation before the test (skipping Red).
- Writing more implementation than the test demands (skipping the smallest-step rule).
- Refactoring and changing behavior in the same commit.
- Performance "improvements" with no measurement.
- Reporting subagent success without checking the diff.
