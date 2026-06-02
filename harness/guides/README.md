# Guides

Rules — what the agent must *do* (and not do). IRON LAWs (non-negotiable) and GOLDEN RULES (strongly preferred). Loaded ambient; enforced by the [drift sensor](../sensors/drift.md). Per-agent adapters lift these into the agent's rules surface (`.cursor/rules/*.mdc`, the directive block of `CLAUDE.md`, etc.).

For the full reasoning and references behind each rule, see the paired file in [`../corpus/`](../corpus/README.md).

## Sub-directories

| Directory | What lives here | Activation |
|---|---|---|
| `principles/` | Universal engineering rules (rules extracted from `corpus/principles/`). | Ambient (always) |
| [`idioms/`](idioms/README.md) | Stack-specific rules (rules extracted from `corpus/idioms/<stack>/`). | Ambient (lazy by region) |
| [`domain/`](domain/README.md) | Business-rule constraints (rules extracted from `corpus/domain/`). | Ambient (always) |
| [`process/`](process/README.md) | What happens at each phase of the workflow. | Loaded when entering a phase |

## File format

Each guide file is short — it carries only the rules, not the reasoning:

```markdown
# <Topic> — rules

The rules from [`corpus/<layer>/<name>.md`](../../corpus/<layer>/<name>.md).
Loaded ambient; enforced by the drift sensor.

## IRON LAW(S)

Non-negotiable.

## GOLDEN RULES

Ideals. Deviation requires reasoning.

---

Traces to: [`corpus/<layer>/<name>.md`](../../corpus/<layer>/<name>.md).
```

Process files are exceptions — they are entirely prescriptive (phase-by-phase rules) and do not pair with a corpus file. They live directly under `process/`.

## How adapters use guides

A per-agent adapter at `harness/adapters/<agent>/activation.md` describes how this directory is projected into the agent's rules surface. Examples:

- **Cursor** — every `guides/principles/*.md` becomes a `.cursor/rules/<name>.mdc` file.
- **Claude Code** — guides are referenced from `CLAUDE.md`'s directive block; the agent reads them on demand.
- **Codex** — referenced from `AGENTS.md`.

The drift sensor reads guides directly from `guides/` regardless of how the agent surface is wired.

## Empty vs. populated

- `guides/principles/` ships populated with 29 files extracted from the corpus principles.
- `guides/process/` ships populated with the workflow phase files.
- `guides/idioms/` and `guides/domain/` ship empty; the **bootstrap** action and Learning flywheel populate them.

## Activation

Ambient (always loaded for principles, domain, process; lazy-by-region for idioms). The agent operates under guide rules at all times. Enforced by the [drift sensor](../sensors/drift.md) inside the **check-drift**, **verify**, and **audit** actions.

When the agent needs the reasoning *behind* a rule, it follows the forward-link from the guide to the paired corpus file. Corpus is on-demand; guides are not.
