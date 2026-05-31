# Project Harness

This is the project's harness — a body of engineering knowledge organized into five layers, plus per-agent bindings. Each layer answers a different kind of question about how to write code in this project. The harness is project-owned: every file under `harness/` is versioned with the code and may be edited freely by the team.

## Three surfaces

- **Corpus** — `harness/principles/`, `harness/idioms/`, `harness/domain/`, `harness/state/`, `harness/process/`, plus `harness/learning/` and `harness/archive/`. Project-owned, agent-agnostic. Versioned with the code.
- **Adapters** — `harness/adapters/<agent>/`. Per-agent bindings: how each abstract lifecycle action maps to that agent's invocation surface (slash command, rules file, CLI, etc.). Ships with the corpus.
- **Agent runtime** — agent-specific (`.claude/`, `.cursor/`, `.aider.conf.yml`, etc.). Tool-owned.

## The five layers

| Layer | What it answers | Activation | Authorship |
|---|---|---|---|
| [`principles/`](principles/README.md) | What does good engineering look like, regardless of stack? | Ambient (always) | Literature + discipline |
| [`idioms/`](idioms/README.md) | How does *this* stack express those principles? | Ambient (lazy by region) | Lead engineer + agent |
| [`domain/`](domain/README.md) | What business rules constrain this codebase? | Ambient (always) | Domain expert + engineer |
| [`state/`](state/README.md) | What is true about the codebase right now? | Ambient (always) | Agent + human |
| [`process/`](process/README.md) | What happens at each phase of the workflow? | Loaded when entering a phase | Lead engineer + agent |

Plus three supporting directories:

- [`learning/`](learning/README.md) — staging area for the Learning flywheel.
- [`archive/`](archive/README.md) — pruned content, preserved with reasoning.
- [`adapters/`](adapters/README.md) — per-agent bindings for lifecycle actions and sensor execution.

## Assumptions

The corpus is written for projects that already have these in place. Missing one is not a hard failure — the corresponding phase degrades to "ask the user / surface the gap" — but the harness is most useful when all four exist:

| Assumption | Used by |
|---|---|
| **A way to track work** — anywhere on the spectrum from a full issue tracker (Jira / Linear / GitHub Issues / Asana) to a `TODO.md` to a sticky note | The **spec** phase opens on a unit of work. A tracker card ID lets the agent fetch the description automatically; without one, the agent authors the spec inline from a conversation. Either works. |
| **Adequate sensors** — lint, type-check, test runner, build command, optionally coverage | The **verify** phase gates every commit on these. Their commands are recorded in `state/CODEBASE_STATE.md`. |
| **Pull request process** (GitHub PRs, GitLab MRs, etc.) | The **review** phase spawns review agents on a diff; comment-driven verification re-runs sensors. The **release** phase opens the PR with the tracker link in the body. |
| **CI pipeline, ideally with CD** | The **release** phase assumes CI runs on the PR (sensors as a backstop) and that merge triggers a deploy. If CD isn't wired up yet, the harness still works — CI gates the merge; deploy stays manual. |

## Roles

The corpus is exercised through four roles:

- **Guides** — ambient prose loaded into context.
- **Sensors** — automated checks and instruments that read code / state.
- **Flywheels** — Learning (additive) and Pruning (subtractive).
- **Discipline** — the act of auditing code, corpus, and flywheel loops.

Workflows are folded in — actions and review agents are how each phase executes, not a separate role.

## Activation

- **Ambient** — loaded by context. No invocation needed.
- **Invoked** — a lifecycle action. Either agent-invoked inside a process phase, or user-invoked for heavyweight operations.

No event hooks. Process discipline drives the lifecycle.

## Lifecycle actions

Invoked by the agent inside process phases:

| Moment | Action | Phase |
|---|---|---|
| Pre-task | **orient** | planning entry |
| Post-edit | **check-drift** | implementation exit |
| Pre-commit | **verify** | verification gate |
| Post-commit | **learn** | release / capture |

## Heavyweight actions

Invoked by the user:

- **bootstrap** — initial audit of a fresh install. Detects stack, scaffolds `idioms/<stack>/`, initializes `state/`.
- **audit** — full dual audit (Learning + Pruning flywheels).
- **synthesize** — promote inbox items into the right corpus layer.
- **mode** — set pacing mode.

How each action is actually invoked in your agent (slash command, rules-file trigger, CLI, etc.) lives in `adapters/<your-agent>/lifecycle.md`. If your agent is not listed, see `adapters/_generic/`.

## Pacing modes

**mode** action with `<paired|solo|autopilot>`. Paired by default.

- **paired** — ask at every phase boundary; the user drives.
- **solo** — work independently; stop on hard problems.
- **autopilot** — maximally autonomous; assumption log at the end.

Agents without a notion of autonomy levels collapse to a single mode; the phases still run.

## Writing conventions

Carried across every layer:

- **IRON LAW** — non-negotiable. `## IRON LAW` heading.
- **GOLDEN RULES** — ideals; deviation requires reasoning. `## GOLDEN RULES` heading.
- Idiom files end with a **Traces to:** footer pointing at the principle they instantiate.
- Files ship with real content. Placeholders are filled in by the **bootstrap** action when the corpus is first installed.

## Two flywheels

**Learning — additive:**
1. Agent encounters a gap → writes a candidate to `learning/inbox/`.
2. Human gates by confidence.
3. **synthesize** promotes into the right corpus layer.

**Pruning — subtractive:**
1. **audit** detects discrepancies between corpus claims and codebase reality.
2. Staleness classified: factually wrong / aspirationally stale / domain-stale / process-stale.
3. Content moves to `archive/` with reasoning recorded.

## Reload after corpus writes

Ambient corpus content is loaded **once per session** in most agents. When **synthesize** promotes an inbox item into a layer, or **audit** archives a stale file, the active session still has the old content — the new rules are not in context yet.

Every flywheel-writing action ends with a **reload prompt**:

1. Save anything in the current session that the corpus change should not blow away.
2. Reset the agent's context (the agent's context-clear primitive — see `adapters/<your-agent>/activation.md`).
3. Re-prompt. The next turn's ambient load reads the updated corpus.

**learn** does **not** require a reload — it writes to `learning/inbox/`, which is not ambient.

---

> Installed via [keystone](https://github.com/tacoda/keystone). The corpus is yours after install — keystone is not a runtime dependency.
