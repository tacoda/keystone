# Port: Guide

**Activation:** Ambient by default — every guide in the resolved cascade is loaded into the agent's context at session start. The `<topic>` directory sets the default activation (see [Default activation by topic](#default-activation-by-topic)); an optional per-guide `globs:` field narrows it (see [Globs](#globs)).
**Purpose:** The rules the agent must follow. The *what* and *what not*. Reasoning lives in the paired corpus file.

## Path convention

```
harness/guides/<topic>/<name>.md                              # project-owned
harness/plugins/<plugin>/guides/<topic>/<name>.md             # plugin-owned (read-only)
```

`<topic>` groups related guides (`process`, `principles`, `idioms`, `domain`, `computational`). Topic directories are open-ended — adding a new topic is just creating a new directory.

## Required shape

```markdown
---
globs:               # optional; see "Globs" below
  - "src/billing/**"
  - "!src/billing/legacy/**"
---
# <Name> — rules

<one-sentence framing of what this guide governs>

## IRON LAW(S)

<non-negotiable rules; omit this section when nothing here qualifies>

## GOLDEN PATH

<strong, explicit standards; omit when nothing here qualifies>

## RULES

<regular rules — the default tier; most directives live here>

For reasoning, see [`corpus/<topic>/<name>.md`](corpus/<topic>/<name>.md).
```

- **H1 title** — required. Format: `# <Name> — rules` (the `— rules` suffix is convention, not enforced).
- **Frontmatter** — optional. Recognized keys: `kind:` (declared on computational guides; documented in their topic README), `globs:` (see below). Unrecognized keys are ignored. Bare-content guides remain valid.
- **Forward-link to corpus** — required when a paired corpus file exists. Harness-root-relative path (no `../` segments; `keystone doctor` enforces).
- **Length** — short. A guide is rules; long-form belongs in corpus. Rough ceiling: one screen.

### Rules tiers

Rules in a guide are organized by strength. Only `## RULES` is mandatory — omit the special tiers when no rule warrants them (the common case).

| Tier              | When to use it                                                                                                         |
| ----------------- | ---------------------------------------------------------------------------------------------------------------------- |
| `## IRON LAW(S)`  | Non-negotiable. Violation causes real damage (data loss, security breach, contract break). Rare by design.             |
| `## GOLDEN PATH` | Stronger than regular rules; **deviation requires reasoning**. May be concrete prescriptions or aspirational ideals.   |
| `## RULES`        | Regular rules. The default tier — most directives live here.                                                           |

The tier framing is part of the agent's reading discipline: iron laws short-circuit any conflicting instruction; golden path can be overridden only with explicit reasoning; regular rules can be overridden when a stronger rule applies.

## Default activation by topic

Activation has two gates: the topic's default behavior, then the optional `globs:` filter. The topic gate runs first.

| Topic                    | Default activation                                                          |
| ------------------------ | --------------------------------------------------------------------------- |
| `guides/domain/`         | Ambient — every action.                                                     |
| `guides/idioms/<stack>/` | Ambient, lazy by region — only when the touched region's stack matches.     |
| `guides/process/`        | On phase entry — only when the agent enters the matching phase.             |
| `guides/computational/`  | Editor / LSP / on-save — driven by the underlying tool, not the harness.    |
| `guides/principles/`     | Ambient — every action. (Universal principles ship under `policies/`.)      |

New topic directories are open-ended; their default activation must be documented in the topic's `README.md`.

## Globs

`globs:` is an optional frontmatter list of glob patterns that **narrows** a guide's activation. It can only remove a guide from activation; it never expands it.

Globs reflect the project's real code structure — they reference paths that already exist in the codebase, recorded in the guide's frontmatter. The **bootstrap** action seeds initial globs from the region map in `corpus/state/CODEBASE_STATE.md`; idiom guides pick up the regions of the stacks they live under. Hand-written globs are valid, but they should still target real code regions — globs that match no files are a learning signal, not a feature.

```
activates ⇔ (topic default fires) ∧ (globs match, or no globs declared)
```

Worked examples:

| Guide                                          | `globs:`                  | Fires when …                                                  |
| ---------------------------------------------- | ------------------------- | ------------------------------------------------------------- |
| `guides/domain/orders.md`                      | *(omitted)*               | Always — today's behavior.                                    |
| `guides/domain/billing.md`                     | `["src/billing/**"]`      | A touched file matches `src/billing/**`.                      |
| `guides/idioms/typescript/hooks.md`            | *(omitted)*               | The touched region is the TS stack — today's behavior.        |
| `guides/idioms/typescript/hooks.md`            | `["src/web/**"]`          | Region is TS **and** a touched file matches `src/web/**`.     |
| `guides/process/implementation.md`             | `["infra/**"]`            | Entering implementation **and** a touched file matches.       |

### Field

- **Type** — list of glob strings (`bmatcuk/doublestar/v4` semantics). Paths are repo-relative POSIX. Adapters that need OS-specific paths convert at the projection boundary.
- **Order** — irrelevant. A file matches if it satisfies any positive pattern and no negative pattern.
- **Negation** — entries prefixed with `!` exclude a sub-tree (`"!src/legacy/**"`).
- **Empty list** — `globs: []` is a parse error. Either omit the key or list patterns.
- **Absence** — omitted key = no narrowing. The topic default fires unchanged.

### The narrow-only invariant

Globs **cannot** make a guide fire when its topic default would not. A `process/` rule's globs cannot make it ambient. An `idioms/typescript/` rule's globs cannot make it activate in the Go region. If you need a rule to fire somewhere its topic does not reach, you need a different topic (or a different guide), not a wider set of globs.

### The touched-files set

Globs match against an action-scoped set of paths. Each action computes its own set; the globs are re-evaluated per action.

| Action          | Touched-files set                                                |
| --------------- | ---------------------------------------------------------------- |
| `orient`        | Paths the agent plans to read or edit (best-effort).             |
| `check-drift`   | Files in the current diff.                                       |
| `verify`        | Files in the current diff.                                       |
| `audit`         | The full repo file set (per-rule globs still apply).             |
| `review`        | Files in the diff under review.                                  |
| Editor-time     | The open buffer plus files read or written in the active turn.   |

### Computational guides

`globs:` on a `guides/computational/` entry is **documentation, not enforcement** — the actual tool (LSP, formatter, hook) reads its own config, not this frontmatter. If the recorded `globs:` diverges from the tool's effective configuration, the `stack-drift` sensor flags it.

## Cascade behavior

For a given `guides/<topic>/<name>.md`:

1. The project's `harness/guides/<topic>/<name>.md` always wins by default.
2. Otherwise, among plugins, plugins nested deeper in `keystone.json` refine the outer plugins they're nested in.
3. A `strict.guides: [<name>]` declaration on any tree node makes that item absolute — nothing else can override it, not the project, not any other plugin. `keystone verify` reports a violation if any layer attempts to shadow a strict item.

The framework never composes overlapping guides for the same name. Exactly one file loads.

**Globs under cascade.** The winner's `globs:` is the only one consulted. An override does not inherit the base's globs: if a project guide overrides a plugin guide and only the plugin had `globs:`, the override loads under its own (possibly absent) globs. The resolver is pure — one file wins, its frontmatter is what's read.

## Drift and corpus pairing

Every guide *should* have a paired corpus file at `corpus/<topic>/<name>.md`. The `drift` sensor flags rules in guides that have no corresponding corpus reasoning, and corpus entries that no guide references.

A guide without a corpus pair is permitted (some rules are self-evident), but `keystone doctor` reports them so authors can decide.

## Example

```markdown
# Surgical edits — rules

Touch only what the task requires. Don't refactor adjacent code, don't reformat
files you didn't change, don't "improve" comments you didn't introduce.

## RULES

- Every changed line traces directly to the user's request.
- Orphaned imports/variables your change created: remove.
- Pre-existing dead code: mention, don't delete (unless asked).

For reasoning, see [`corpus/process/surgical-edits.md`](corpus/process/surgical-edits.md).
```

## Authoring

```
keystone new guide <topic>/<name>
```

Scaffolds the guide and a paired corpus stub at the conventional paths.
