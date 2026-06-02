# Guides

Rules — what the agent must *do* (and not do). Three tiers: regular **RULES** (the default), **GOLDEN RULES** (aspirational but explicit; stronger than regular), and **IRON LAWS** (non-negotiable; rare by design). Loaded ambient; enforced by the [drift sensor](../sensors/drift.md). Per-agent adapters lift these into the agent's rules surface (`.cursor/rules/*.mdc`, the directive block of `CLAUDE.md`, etc.).

For the full reasoning and references behind each rule, see the paired file in [`../corpus/`](../corpus/README.md).

## Kind

Every guide declares a `kind:` in its frontmatter. The kind says *how* the guidance is delivered:

**Inferential** — natural-language rules an agent reasons about. Markdown files under `principles/`, `idioms/`, `domain/`, and `process/`. This is what guides have been historically; it remains the default kind.

**Computational** — deterministic ambient enforcement that does not depend on agent reasoning. Examples: a language server giving live type/error feedback, an editor formatter, a pre-save linter rule. These live under [`computational/`](computational/README.md) and are inventoried by the bootstrap action based on what the project's stack supports.

Kind classifies the guide, not the thing the guide is about. An inferential rule about *how to write good TypeScript types* lives under `idioms/typescript/`, not `computational/` — even though the underlying enforcement (a TS LSP) is computational. The folder reflects the guide's mechanism; the LSP entry lives separately under `computational/`.

## Sub-directories

| Directory | Kind | What lives here | Activation |
|---|---|---|---|
| `principles/` | inferential | Universal engineering rules (rules extracted from `corpus/principles/`). | Ambient (always) |
| [`idioms/`](idioms/README.md) | inferential | Stack-specific rules (rules extracted from `corpus/idioms/<stack>/`). | Ambient (lazy by region) |
| [`domain/`](domain/README.md) | inferential | Business-rule constraints (rules extracted from `corpus/domain/`). | Ambient (always) |
| [`process/`](process/README.md) | inferential | What happens at each phase of the workflow. | Loaded when entering a phase |
| [`computational/`](computational/README.md) | computational | Deterministic ambient enforcement — language servers, formatters, editor checks the stack supports. | Ambient (in-editor / on-save) |

## File format

Each guide file is short — it carries only the rules, not the reasoning. Sections appear in order of strength; **only `## RULES` is mandatory**. Omit the special tiers when no rule warrants them — that's the common case.

```markdown
# <Topic> — rules

The rules from [`corpus/<layer>/<name>.md`](../../corpus/<layer>/<name>.md).
Loaded ambient; enforced by the drift sensor.

## IRON LAW(S)

Non-negotiable. Violation causes real damage. Rare by design — omit this section if nothing here qualifies.

## GOLDEN RULES

Strong, explicit standards. Stronger than regular rules; deviation requires reasoning. May be concrete prescriptions or aspirational ideals. Omit if nothing here qualifies.

## RULES

Regular rules. The default tier — most directives live here.

---

Traces to: [`corpus/<layer>/<name>.md`](../../corpus/<layer>/<name>.md).
```

**Tier discipline.** The special tiers (IRON LAW, GOLDEN RULE) derive their force from being rare. Default new rules to `## RULES`; promote upward only when the user confirms during **synthesize**.

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
- `guides/idioms/`, `guides/domain/`, and `guides/computational/` ship empty; the **bootstrap** action and Learning flywheel populate them.

Anything a computational guide needs at install time (a particular editor config file, an LSP binary, an agent setting) is exposed as an option on `keystone init` rather than shipped by default.

## Activation

Ambient (always loaded for principles, domain, process; lazy-by-region for idioms). The agent operates under guide rules at all times. Enforced by the [drift sensor](../sensors/drift.md) inside the **check-drift**, **verify**, and **audit** actions.

When the agent needs the reasoning *behind* a rule, it follows the forward-link from the guide to the paired corpus file. Corpus is on-demand; guides are not.
