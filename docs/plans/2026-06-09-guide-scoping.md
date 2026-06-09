---
status: approved
captured: 2026-06-09
approved: 2026-06-09
topic: guide scoping (path-narrowed activation)
---

# Plan — Guide scoping (path-narrowed activation)

Add a first-class `globs:` field to guide frontmatter so the rules that load are constrained by path. **Globs narrow existing category-based activation; they never expand it.** A guide without `globs:` keeps today's defaults — the change is fully opt-in.

## 1. Problem

Activation today is category-driven and described in prose:

| Category | Today's activation | Source |
|---|---|---|
| `guides/domain/` | Ambient, always loaded | `guides/domain/README.md` |
| `guides/idioms/<stack>/` | Ambient, lazy-by-stack (region) | `guides/idioms/README.md` |
| `guides/process/` | Loaded on entering the phase | `guides/process/README.md` |
| `guides/computational/` | Editor / LSP / on-save | `guides/computational/README.md` |
| `policies/<policy>/guides/principles/` | Ambient, always loaded | `policies/universal/...` |

Three problems with prose-only activation:

1. **Path scoping is implicit and per-adapter.** Cursor uses native `globs:` on the generated `.cursor/rules/*.mdc`. Claude Code rederives the same idea inside the **orient** playbook by walking `corpus/state/CODEBASE_STATE.md`. Every adapter rebuilds its own version of the truth.
2. **There is no machine-readable contract.** The drift sensor's `## Inputs` line already says "the file paths each rule applies to" — but those paths are inferred from the directory the rule lives in, not declared by the rule. A sensor cannot honor globs it cannot read.
3. **All-or-nothing for domain / principles.** A domain rule about billing or a principle about migrations is loaded everywhere. There is no way to say "this only fires under `src/billing/**`" short of inventing a new category.

## 2. Decision

Add `globs:` to guide frontmatter. Globs are an **intersection filter** on top of the category default — never a replacement, never an expansion.

```
activates ⇔ (category default fires) ∧ (globs match, or no globs declared)
```

Worked examples:

| Guide | Frontmatter | When it activates |
|---|---|---|
| `guides/domain/orders.md` | *(no `globs:`)* | Always (today's behavior). |
| `guides/domain/billing.md` | `globs: ["src/billing/**"]` | Only when a touched file matches. |
| `guides/idioms/typescript/hooks.md` | *(no `globs:`)* | When the touched region is the TS stack (today). |
| `guides/idioms/typescript/hooks.md` | `globs: ["src/web/**"]` | When the region is TS **and** a touched file matches `src/web/**`. |
| `guides/process/implementation.md` | `globs: ["src/**"]` | When entering implementation **and** a touched file matches. (Probably not useful for `process/`; the option exists for symmetry.) |
| `policies/universal/guides/principles/migrations.md` | `globs: ["db/**", "migrations/**"]` | Always-loaded *as today*, but only when those paths are touched. |

The contract is intentionally one direction: **globs can only remove a guide from activation, never add it.** A `process/` rule's globs cannot make it ambient. An `idioms/typescript/` rule's globs cannot make it activate in the Go region.

## 3. Schema

Add an optional `globs:` field to the guide frontmatter:

```markdown
---
kind: inferential
globs:
  - "src/billing/**"
  - "tests/billing/**"
---
```

- **Type** — list of glob strings. Order is irrelevant; any match counts.
- **Glob syntax** — `bmatcuk/doublestar/v4` (Go), gitignore-style with `**`. Already in the Go ecosystem; the framework binary uses it for stack detection. *(Open question — see §10.)*
- **Negation** — support `!`-prefixed entries (`"!src/legacy/**"`) for excluding sub-trees. A file matches if it satisfies any positive pattern and no negative pattern. Keeps the common case simple while letting an idiom say "this stack, except the legacy folder."
- **Empty list** — `globs: []` is a parse error. Either omit the key (default activation) or list patterns.
- **Absence** — omitted key = no narrowing. The category default fires unchanged.

Paths in the field are repo-relative POSIX. Adapters that need OS-specific forms convert at the projection boundary.

## 4. Activation semantics by category

Globs are consulted only when the category default already fires. The category gate runs first; globs are the second filter.

### 4.1 `guides/domain/`

- **Default fire** — every action.
- **With `globs:`** — fires only on actions whose **touched-files set** match the globs.
- **Rationale** — domain rules with a clear regional boundary (billing, auth, search) should not bloat context outside their region.

### 4.2 `guides/idioms/<stack>/`

- **Default fire** — when the touched region's stack matches `<stack>`. ("Lazy-by-region", today's behavior.)
- **With `globs:`** — fires when the stack matches **and** touched files match the globs. Globs *intersect* with the existing stack region, not used as a replacement.
- **Rationale** — most idioms cover an entire stack; some cover only a sub-tree (`apps/admin/**` vs `apps/marketing/**` within the same TS stack). Globs express that without inventing pseudo-stacks.

### 4.3 `guides/process/`

- **Default fire** — when the agent enters the phase.
- **With `globs:`** — fires when the phase is entered **and** touched files match the globs.
- **Rationale** — usually unnecessary; phase docs are universal. Supporting globs keeps the model orthogonal and lets a project carve out (e.g.) a stricter implementation rule for `infra/**`.

### 4.4 `guides/computational/`

- **Default fire** — editor/LSP/on-save, per the tool's own configuration.
- **With `globs:`** — surfaced for documentation and the audit sensor only. The actual tool (LSP, formatter) doesn't read our frontmatter — its own config file determines what files it touches. Globs here describe intent; if the recorded globs diverge from the tool's real config, the **stack-drift** sensor flags it.
- **Rationale** — keeps the schema uniform across categories without overpromising on enforcement.

### 4.5 `policies/<policy>/guides/principles/` (and any future policy layer)

- **Default fire** — ambient, always loaded (today).
- **With `globs:`** — fires when touched files match the globs.
- **Cascade interaction** — see §6.

## 5. The "touched-files set"

Globs match against an action-scoped set of paths. The set is computed once per action invocation.

| Action | Touched-files set |
|---|---|
| `orient` | The paths the agent plans to read or edit, derived from the task description plus any user-supplied paths. Best-effort. |
| `check-drift` | The files in the current diff. |
| `verify` | The files in the current diff. |
| `audit` | The full repo file set (audit is by definition unbounded; globs still narrow but most rules will match somewhere). |
| `review` | The files in the diff under review. |
| Editor-time activation (Cursor `globs:`, Claude Code via `orient`) | The currently-open buffer plus paths the agent has read or written in the active turn. |

The set is computed by the harness layer that drives the action, not by each rule. Rules consume the globs; the runtime computes touched-files.

## 6. Cascade interaction

The cascade (project over plugin, with `strict` / `required` semantics) is unchanged. Globs compose on top:

1. **Resolution wins by path, as today.** `harness/guides/<port>/<name>.md` in the project beats the same path in a plugin. Only the winner's `globs:` is consulted.
2. **The winner's globs are the only ones.** An override does not inherit the base's globs. This keeps the resolver pure — one file wins, period.
3. **Narrow-only is enforced at the file level, not the cascade level.** If a project override drops `globs:` while the policy had one, the override loads *more broadly* than the policy did. That's "expansion" relative to the policy, but it's the user's deliberate choice in their own file — not a cascade operation. Document it; don't block it.
4. **`strict` does not constrain globs.** `strict` blocks descendants from overriding; it does not freeze the winning file's `globs:` value.

## 7. Adapter projection

Each adapter declares how it projects globs into its rules surface. The projection is the only place adapters diverge.

### 7.1 Cursor (native)

Globs project directly to `globs:` on the generated `.cursor/rules/*.mdc`. Today's `bootstrap` already emits per-stack `.mdc` files with globs; the change is to thread the per-guide `globs:` through.

```mdc
---
description: TypeScript hooks rules
globs: "src/web/**"
---
<rule body>
```

For guides with no globs, the existing category-derived glob (or `alwaysApply: true`) is used as before.

### 7.2 Claude Code (via `orient`)

Claude Code has no native scoping. The **orient** action playbook already walks `CODEBASE_STATE.md` to find idioms for the touched region. The change:

- Extend `orient` to read `globs:` on every candidate guide and only load guides whose globs match the touched-files set (or guides with no globs).
- A short generated index at `harness/corpus/state/GLOBS_INDEX.md` (path → list of guides) keeps `orient` from re-walking the tree every turn. The index is regenerated on `bootstrap`, `synthesize`, and `audit` writes.

### 7.3 Codex / AGENTS.md, Aider, Cline, Continue, GitHub Copilot, Goose, Pi

Each of these adapters currently follows one of two patterns:

- **Pointer-only** (Codex, Aider, Continue, etc.) — the agent reads `harness/` on demand. These adopt the Claude Code model: globs are honored inside the action playbook, with the `GLOBS_INDEX.md` shared across adapters.
- **Native glob-aware** (currently only Cursor) — use native glob frontmatter when supported.

`harness/adapters/<agent>/activation.md` documents which path the adapter takes. A capability matrix row "Lazy-by-region (path-scoped)" makes the difference visible.

### 7.4 `_generic`

The fallback `_generic` adapter does *not* honor `globs:` — it falls back to category defaults. Document the gap; it's the price of supporting arbitrary agents.

## 8. Sensor behavior

The **drift** sensor's contract already names "the file paths each rule applies to" as an input. Globs make that input concrete:

- For each loaded guide, drift compares findings only against files matching its globs (or all files in the diff, if no globs declared).
- Findings outside the globs are dropped before tier classification. A rule cannot flag a violation in a file it does not claim.
- The **audit** action — which is by definition cross-cutting — still walks every file, but per-rule findings are filtered by per-rule globs. Audit reports include a "glob coverage" section that lists files in the diff that no rule's globs matched (i.e., possibly under-covered regions).

The **stack-drift** sensor gets a new check: for computational guides, if the recorded `globs:` diverges from the underlying tool's effective configuration (e.g., the formatter's `include`/`exclude`), the sensor flags it.

No other sensor changes shape.

## 9. Flywheel (learn / synthesize / bootstrap)

### 9.1 `bootstrap`

When seeding `guides/idioms/<stack>/`, bootstrap writes the stack's region globs into the generated stub's `globs:` — same value it uses for the Cursor adapter glob. A computational guide entry records the tool's configured paths as `globs:` so the stack-drift sensor has something to compare against.

### 9.2 `learn`

The learn candidate frontmatter gains a `proposed-globs:` field (optional). The agent records the touched paths from the surprising interaction so synthesize has signal to work from. No promotion happens here.

### 9.3 `synthesize`

When promoting a candidate, synthesize:

1. Inspects `proposed-globs:` if present.
2. Proposes a `globs:` value derived from the candidate's evidence.
3. **Always shows the proposed globs to the user before writing.** Per the harness "no silent overwrites" iron law, the globs are part of the diff.
4. Defaults to no `globs:` if the candidate's evidence is genuinely cross-cutting.

Synthesize also regenerates `harness/corpus/state/GLOBS_INDEX.md` after any guide write.

## 10. Decisions locked / open questions

**Locked (2026-06-09):**

- **Granularity is per-action.** Each action computes its own touched-files set; globs are re-evaluated per action. No "session-wide" globs.
- **Cascade — winner's globs are the only ones.** An override does not inherit the base's `globs:`. The resolver stays pure: one file wins, its globs are consulted, the rest are ignored. Documented behavior in §6.
- **`GLOBS_INDEX.md` is part of the design.** Generated at `harness/corpus/state/GLOBS_INDEX.md` by `bootstrap` and `synthesize`; consumed by pointer-style adapters (Claude Code, Codex, Aider, etc.) inside their action playbooks.

**Still open:**

1. **Glob syntax.** `bmatcuk/doublestar/v4` is the cheap default — already a transitive dep via stack detection. Confirm before we lock the contract; gitignore semantics (with negation) are user-friendly but introduce two flavors of `!` (negation vs literal). Pick one and document.
2. **Glob coverage warnings.** When `audit` finds files matched by no rule's globs, is that a warning, an `info`, or silent? Probably `info` — it's a learning candidate, not a violation.
3. **Backward-compat for existing installs.** Today's installs have no `globs:` in any guide. On upgrade, behavior must be byte-identical. The migration is purely additive (add the field where helpful) and is the user's call to make.
4. **`policies/universal/guides/principles/` cascade.** Should the shipped principles have `globs:` set? Most don't — they really are universal. A few (`migrations.md`, `concurrency.md`) might benefit. Defer to a follow-up pass after the schema lands.

## 11. Acceptance criteria

This change is done when:

- [ ] `docs/ports/guide.md` defines the `globs:` frontmatter field — glob syntax, semantics, and the narrow-only contract. (`docs/schemas/` is JSON-only; the markdown schema lives in the port contract.)
- [ ] Each `guides/<category>/README.md` describes how globs narrow that category's default activation, with one worked example per category.
- [ ] The drift sensor `Inputs` line names `globs:` as a per-rule input and the sensor honors it.
- [ ] The Cursor adapter projects `globs:` to `globs:` on generated `.mdc` files.
- [ ] The Claude Code adapter's **orient** playbook reads `globs:` and the generated `GLOBS_INDEX.md` to gate idiom loading.
- [ ] At least one other text-pointer adapter (Codex or Aider) documents the same `GLOBS_INDEX.md` flow.
- [ ] `bootstrap` writes `globs:` into seeded idiom stubs based on detected region globs.
- [ ] `synthesize` shows a proposed `globs:` value in every promotion diff.
- [ ] The `_generic` adapter documents that globs are not honored.
- [ ] Existing installs (no `globs:` in any file) behave identically — covered by the end-to-end golden test for `keystone init`.
- [ ] A new end-to-end test covers a guide with `globs:` and asserts it is loaded only for matching paths and ignored for non-matching paths.

## 12. Out of scope

- **Re-organizing categories.** Scoping makes `domain/`-vs-`idioms/<stack>/` boundaries less load-bearing, but this plan does not collapse or rename categories.
- **Cross-rule composition.** No "rule A overrides rule B for these paths." Cascade still operates at file granularity.
- **Time-based scoping.** No "fires only during release phase after 5pm." If we want it, that's a separate `phase:` or `mode:` field.
- **Per-file authority levels.** Tier (IRON LAW / GOLDEN / RULES) is per-section inside the file, not in frontmatter. Unchanged.
- **Globs-aware learning suggestions.** The Learning flywheel could *suggest* `globs:` for an existing rule based on where violations cluster. Worth doing — later.

## 13. Phased rollout

| Phase | Deliverable | Status |
|---|---|---|
| A | Schema + activation semantics doc; READMEs updated; **no runtime behavior change** | shipped (1.0.4) |
| B | Drift sensor honors `globs:`; check-drift / audit playbooks updated; existing guides keep working with no globs | shipped (1.0.4) |
| C | `bootstrap` seeds code-grounded `globs:` into idiom and computational guides from the region map; generates initial `GLOBS_INDEX.md` | shipped (1.0.4) |
| D | `bootstrap` projects each guide's `globs:` to a per-guide `.cursor/rules/keystone-<topic>-<name>.mdc`; Cursor adapter doc updated | shipped (1.0.4) |
| E | `orient` reads `GLOBS_INDEX.md` to gate per-guide loading; all 8 pointer-style adapters (Claude Code, Codex, Aider, Cline, Continue, Goose, Copilot, Pi) updated to describe the new flow; Codex AGENTS.md cascade description corrected | shipped (1.0.4) |
| F | `learn` candidate gains `proposed-globs:`; `synthesize` proposes `globs:` from evidence, regenerates `GLOBS_INDEX.md` and the Cursor `.mdc` projections after every promotion | shipped (1.0.4) |

Each phase is independently shippable. Phase A is the only one that locks the contract; the rest are additive. **All six phases shipped in 1.0.4** as a single coherent release (markdown-only changes, agent-side behavior; the Go runtime is unchanged).

## 14. Follow-ups (considered, deferred)

Two narrowing axes were considered alongside `globs:` and explicitly deferred. They share the same narrow-only contract — they only ever remove a guide from activation, never add it — and would compose as `activates ⇔ topic-default ∧ scope ∧ <axis>`.

### 14.1 Phase scoping (`phase:`)

A frontmatter list of phase names that gates a non-`process/` guide to certain workflow phases:

```markdown
---
phase: [implementation, review]
---
# Billing invariants — rules
```

The above keeps `guides/domain/billing.md` from loading during `spec`, `planning`, or `release` actions. Same narrow-only invariant: a `process/spec.md` guide cannot use `phase:` to make itself fire during release.

**Why deferred** — `globs:` alone may cover most of the context savings; adding `phase:` simultaneously expands the schema and the mental model. Land `globs:` first, measure, then decide.

### 14.2 Tags (`tags:`)

A frontmatter list of orthogonal topic facets:

```markdown
---
tags: [security, performance]
---
```

Tags don't shrink ambient load — they enable targeted retrieval. A "security review" action could pull every guide tagged `security` regardless of directory, instead of relying on the directory layout to hold the only grouping. Useful for cross-cutting concerns that span `idioms/`, `domain/`, and `process/`.

**Why deferred** — tags solve a different problem than `globs:` (retrieval vs. activation). Without a concrete retrieval flow that needs them, they're speculative.

### 14.3 The compounding-narrow risk

Each additional narrowing axis raises the chance a rule silently fails to fire because the intersection of all axes is empty. When `phase:` or `tags:` land, the `audit` action will need an explicit "coverage" section that flags files matched by no rule (today's open question §10.2) — same primitive both axes depend on. Land the coverage signal at the same time as the second axis, not later.

---

## Appendix — files that change

Inventory of likely edits. Use as a checklist during implementation; not exhaustive.

**Docs / contracts**
- `docs/ports/guide.md` (port contract for guides — already exists; extend with `globs:` schema)
- `docs/conventions.md` (note `globs:` on the Guide row)

**Templates — root harness**
- `internal/framework/scaffold/templates/harness/guides/README.md`
- `internal/framework/scaffold/templates/harness/guides/{domain,idioms,process,computational}/README.md`
- `internal/framework/scaffold/templates/harness/sensors/drift.md`
- `internal/framework/scaffold/templates/harness/sensors/stack-drift.md`
- `internal/framework/scaffold/templates/harness/actions/{bootstrap,orient,learn,synthesize,audit}.md`

**Adapters**
- `internal/framework/scaffold/templates/harness/adapters/cursor/activation.md`
- `internal/framework/scaffold/templates/harness/adapters/claude-code/activation.md`
- `internal/framework/scaffold/templates/harness/adapters/{codex,aider,cline,continue,github-copilot,goose,pi}/activation.md`
- `internal/framework/scaffold/templates/harness/adapters/_generic/activation.md`

**Plugins shipped under `optional/`**
- Sample principle / idiom files that benefit from a default `globs:` — handled in Phase F, not now.

**Go runtime**
- Frontmatter parser (likely already YAML-front-matter aware) — add `globs` field.
- `GLOBS_INDEX.md` generator — new, called from `bootstrap` and `synthesize`.
- Touched-files-set computation per action — new helper consumed by the action playbooks.

**Tests**
- End-to-end golden test for `keystone init` — assert behavior unchanged when no `globs:` is set.
- New end-to-end test for `globs:` matching and filtering.
