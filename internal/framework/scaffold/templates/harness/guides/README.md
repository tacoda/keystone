# Guides

Rules — what the agent must *do* (and not do). Three tiers: regular **RULES** (the default), **GOLDEN PATH** (aspirational but explicit; stronger than regular), and **IRON LAWS** (non-negotiable; rare by design). Loaded ambient; enforced by the [drift sensor](sensors/drift.md). Per-agent adapters lift these into the agent's rules surface (`.cursor/rules/*.mdc`, the directive block of `CLAUDE.md`, etc.).

For the full reasoning and references behind each rule, see the paired file in [`../corpus/`](corpus/README.md).

## Kind

Every guide declares a `kind:` in its frontmatter. The kind says *how* the guidance is delivered:

**Inferential** — natural-language rules an agent reasons about. Markdown files under `idioms/`, `domain/`, and `process/`. This is what guides have been historically; it remains the default kind.

**Computational** — deterministic ambient enforcement that does not depend on agent reasoning. Examples: a language server giving live type/error feedback, an editor formatter, a pre-save linter rule. These live under [`computational/`](computational/README.md) and are inventoried by the bootstrap action based on what the project's stack supports.

Kind classifies the guide, not the thing the guide is about. An inferential rule about *how to write good TypeScript types* lives under `idioms/typescript/`, not `computational/` — even though the underlying enforcement (a TS LSP) is computational. The folder reflects the guide's mechanism; the LSP entry lives separately under `computational/`.

## Globs

A guide's sub-directory sets a **default** activation (see the table below). An optional `globs:` field in the guide's frontmatter **narrows** that default to a set of paths:

```
activates ⇔ (sub-directory default fires) ∧ (globs match, or no globs declared)
```

Globs reflect the project's real code structure — the **bootstrap** action seeds initial globs from the region map in `corpus/state/CODEBASE_STATE.md`, and idiom guides pick up the regions of the stacks they live under. They are not invented patterns; they are paths that already exist in the codebase, recorded here in the guide's frontmatter.

Globs can only remove a guide from activation — they never expand it. A `process/` rule's globs cannot make it ambient; an `idioms/typescript/` rule's globs cannot fire it inside the Go region. If a guide needs to reach paths its sub-directory does not cover, the answer is a different sub-directory, not a wider set of globs.

```markdown
---
globs:
  - "src/billing/**"
  - "!src/billing/legacy/**"
---
# Billing — rules
...
```

`globs:` is a list of gitignore-style patterns (`**` supported; `!`-prefix excludes). Paths are repo-relative POSIX. Omitting the field keeps today's default behavior — the change is opt-in.

Full contract, including per-action match semantics and cascade interaction: the **Guide port contract** in the Keystone docs.

## Sub-directories

| Directory | Kind | What lives here | Default activation |
|---|---|---|---|
| [`idioms/`](idioms/README.md) | inferential | Stack-specific rules (rules extracted from `corpus/idioms/<stack>/`). | Ambient (lazy by region) |
| [`domain/`](domain/README.md) | inferential | Business-rule constraints (rules extracted from `corpus/domain/`). | Ambient (always) |
| [`process/`](process/README.md) | inferential | What happens at each phase of the workflow. | Loaded when entering a phase |
| [`computational/`](computational/README.md) | computational | Deterministic ambient enforcement — language servers, formatters, editor checks the stack supports. | Ambient (in-editor / on-save) |

Any of these defaults can be narrowed per-guide via `globs:` (see above).

**Universal engineering rules** (the principles rule extracts that used to live in `guides/principles/`) now ship inside the default policy at [`../policies/universal/guides/principles/`](policies/universal). They are still ambient and still part of the always-loaded rule set — they just live under the policies layer rather than under project-owned guides.

## File format

Each guide file is short — it carries only the rules, not the reasoning. Sections appear in order of strength; **only `## RULES` is mandatory**. Omit the special tiers when no rule warrants them — that's the common case.

```markdown
# <Topic> — rules

The rules from [`corpus/<layer>/<name>.md`](corpus/<layer>/<name>.md).
Loaded ambient; enforced by the drift sensor.

## IRON LAW(S)

Non-negotiable. Violation causes real damage. Rare by design — omit this section if nothing here qualifies.

## GOLDEN PATH

Strong, explicit standards. Stronger than regular rules; deviation requires reasoning. May be concrete prescriptions or aspirational ideals. Omit if nothing here qualifies.

## RULES

Regular rules. The default tier — most directives live here.

---

Traces to: [`corpus/<layer>/<name>.md`](corpus/<layer>/<name>.md).
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

- `guides/process/` ships populated with the workflow phase files.
- `guides/idioms/`, `guides/domain/`, and `guides/computational/` ship empty; the **bootstrap** action and Learning flywheel populate them.
- Universal engineering rules ship under [`../policies/universal/`](policies) (the default policy), not under project-owned `guides/`.

Anything a computational guide needs at install time (a particular editor config file, an LSP binary, an agent setting) is exposed as an option on `keystone init` rather than shipped by default.

## Activation

Sub-directory default first, then optional `globs:` narrows it (see **Globs** above):

- `domain/` — ambient, every action.
- `process/` — loaded on phase entry.
- `idioms/<stack>/` — ambient, lazy-by-region (when the touched region matches the stack).
- `computational/` — driven by the underlying tool (LSP, formatter, hook).
- Universal-policy principles — ambient.

The agent operates under guide rules at all times within their activation window. Enforced by the [drift sensor](sensors/drift.md) inside the **check-drift**, **verify**, and **audit** actions.

When the agent needs the reasoning *behind* a rule, it follows the forward-link from the guide to the paired corpus file. Corpus is on-demand; guides are not.
