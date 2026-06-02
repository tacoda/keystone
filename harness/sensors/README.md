# Sensors

Automated checks and instruments that read code or state. One of the four harness components (Corpus / Guides / **Sensors** / Flywheels).

## Kind

Every sensor declares a `kind:` in its frontmatter. The kind says *how* the check is produced:

**Computational** — deterministic execution. A binary, script, or library returns a result that does not depend on agent reasoning. Examples: a unit test runner, a type checker, a lint pass, a build command, a git-history diff.

**Inferential** — produced by an agent reasoning over the code or spec. Examples: an agent doing a functional review, a security review, or walking spec acceptance criteria against the diff.

The bootstrap action inventories both kinds for the project and records them under `corpus/state/CODEBASE_STATE.md`.

## Role

Sensors also split by what they do with their output (orthogonal to kind):

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
| **review** | review | [review-functional](review-functional.md), [review-security](review-security.md), [spec-adherence](spec-adherence.md) |
| **audit** | discipline | [drift](drift.md), [coverage](coverage.md), [risk-fingerprint](risk-fingerprint.md), [traffic-topology](traffic-topology.md) |

How an action is actually invoked, and whether sensors run autonomously or require human prompting, is agent-specific — see `harness/adapters/<your-agent>/sensors.md`. The most common degradation: an agent that cannot run shell commands during a turn surfaces the sensor commands for the human to execute instead.

## Contract shape

Each sensor file begins with YAML frontmatter declaring its kind:

```markdown
---
kind: computational  # or: inferential
---
```

Each sensor declares:

- **Kind** — `computational` (deterministic execution) or `inferential` (agent reasoning).
- **Trigger** — when it runs.
- **Inputs** — what it reads.
- **Exit condition** — what counts as pass.
- **Output** — pass/fail + structured findings.
- **State writes** — which `corpus/state/` file(s), if any. State writes follow the scaffolding safety contract: propose a diff, never silently overwrite.

Tool commands (lint, test, build, type-check, coverage) live in `corpus/state/CODEBASE_STATE.md`, populated by the **bootstrap** action. Sensors read them from there.

## Sensor index

| Sensor | Kind | Purpose |
|---|---|---|
| [lint](lint.md) | computational | Surface-level style and pattern checks |
| [type-check](type-check.md) | computational | Signature and contract consistency |
| [test](test.md) | computational | The project's test suite |
| [build](build.md) | computational | Build / compile / package step |
| [drift](drift.md) | computational | Diff vs. loaded corpus and guide rules |
| [coverage](coverage.md) | computational | Test coverage; proposes state updates |
| [risk-fingerprint](risk-fingerprint.md) | computational | Complexity + coupling + coverage per region |
| [traffic-topology](traffic-topology.md) | computational | Git churn + recency + criticality map |
| [state-region](state-region.md) | computational | Read-only: what State says about a touched region |
| [commit-message](commit-message.md) | computational | Conventional-commit format + no AI attribution |
| [tracker-card-fetcher](tracker-card-fetcher.md) | computational | Fetches tracker card from Jira / Linear / etc. |
| [spec-adherence](spec-adherence.md) | inferential | Walks spec ACs against the diff |
| [review-functional](review-functional.md) | inferential | Agent reviews the diff for logic / behavior bugs |
| [review-security](review-security.md) | inferential | Agent reviews the diff for security concerns |

The bootstrap action confirms which sensors are wired up for this project (e.g., projects without a tracker skip `tracker-card-fetcher`; adapters without sub-agent support skip the review sensors) and records the result in `corpus/state/CODEBASE_STATE.md`.

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
