---
kind: command
id: bootstrap
description: 'One-time initial harness scaffold.'
---
# bootstrap

**One-time initial harness scaffold.** Detect the project's stack, seed `corpus/state/`, scaffold idiom directories with code-grounded globs, inventory computational guides, classify sensors, and generate the globs index. Run once per project.

## Activities

Every activity below produces a concrete file write. Detecting, narrating, or summarizing is not enough — call the edit primitive and land the change before moving on.

1. **Detect the stack.** Inspect the repo (`package.json`, `go.mod`, `pyproject.toml`, `Gemfile`, `Cargo.toml`, `requirements.txt`, `build.gradle`, etc.) and list every primary language, framework, and notable library.
2. **Seed `harness/corpus/state/CODEBASE_STATE.md`.** Propose an edit that replaces every template placeholder with real values:
   - Detected stacks
   - Tool commands (lint / type-check / test / build / coverage) — actual commands, not placeholders
   - Region map (top-level directories → which stacks they hold) — this becomes the source of truth for the globs in step 3
   - CI platform (GitHub Actions / GitLab CI / etc.)
3. **Seed stack idioms with globs.** For each detected stack, scaffold `harness/corpus/idioms/<stack>/` and the paired `harness/guides/idioms/<stack>/`. The guides directory gets a `README.md` describing the stack, and each seeded guide's frontmatter declares `globs:` derived from the region map written in step 2 — the globs reflect real code paths, not invented patterns. Populate idiom files as patterns emerge; new files inherit the region's globs unless they cover a sub-tree (in which case narrow them).
4. **Inventory computational guides.** Record LSP, formatter, and editor enforcement (`.editorconfig`, pre-commit hooks, etc.) under `harness/guides/computational/`. Each entry records `globs:` set to the paths the tool actually covers (read from its config file) — this lets the **stack-drift** sensor compare documented vs. effective configuration.
5. **Classify sensors.** For each sensor in `harness/sensors/`, propose an edit to `CODEBASE_STATE.md` recording whether this adapter can run it. Inferential sensors (`review-functional`, `review-security`, `review-risk`, `review-deployment`, `spec-adherence`) and computational sensors get separate sections.
6. **Generate `harness/corpus/state/GLOBS_INDEX.md`.** Walk every guide under `harness/guides/` and `harness/policies/*/guides/`, read each guide's `globs:` frontmatter, and write the reverse-index (glob pattern → list of guides claiming it). Touch only the `## Index` table; preserve everything else. The index ships with an empty `<pattern>` placeholder row — bootstrap replaces it with real data. Pointer-style adapters (Claude Code, Codex, Aider) read this index to gate idiom loading on the touched-files set.
7. **Project to per-adapter rule surfaces.** For each guide with `globs:` declared, produce the agent-specific projection:
   - **Cursor** (if `.cursor/rules/` exists in the install) — write `.cursor/rules/keystone-<topic>-<name>.mdc` for every guide that declares `globs:`. The `<topic>-<name>` slug comes from the guide's path under `harness/guides/` (e.g. `guides/idioms/typescript/hooks.md` → `keystone-idioms-typescript-hooks.mdc`). The `.mdc` frontmatter mirrors the guide's `globs:`; the body is a single-line pointer at the source guide. Guides without `globs:` get no `.mdc`.
   - **Pointer-style adapters** (Claude Code, Codex, Aider, Cline, Continue, Goose, Pi) — `GLOBS_INDEX.md` from step 6 *is* the projection. No per-guide file is written; each adapter's `orient` playbook reads the index.
   - **`_generic`** — skip. This adapter does not honor `globs:` and falls back to topic defaults.

   Each `.cursor/rules/keystone-<topic>-<name>.mdc` looks like:

   ```mdc
   ---
   description: <guide topic> — <guide name>
   globs: <copied verbatim from the guide's globs:>
   ---

   Rules from `harness/guides/<topic>/<name>.md`. Read the guide for the full ruleset.
   ```

## Iron law

**No silent overwrites of state files.** Propose every diff via the agent's edit primitive; let the user accept or reject each. **Narration is not a write** — bootstrap is incomplete until each file change has actually landed on disk.

## Completion check

Before claiming bootstrap is done, verify empirically:

- `grep -E '<[^>]+>' harness/corpus/state/CODEBASE_STATE.md` returns no template placeholders.
- The `last_reconciled` front-matter value in `CODEBASE_STATE.md` is today's date.
- `harness/corpus/idioms/<stack>/` exists for each detected stack.
- `harness/guides/idioms/<stack>/` exists, paired with each corpus folder above.
- Every seeded `harness/guides/idioms/<stack>/*.md` declares `globs:` in its frontmatter, and those globs reference paths that actually exist in the repo.
- `harness/corpus/state/GLOBS_INDEX.md` exists and its `## Index` table lists at least one row per seeded idiom guide (or is explicitly empty if no guides declare `globs:` yet).
- If `.cursor/rules/` exists in the install, every guide listed in `GLOBS_INDEX.md` has a matching `.cursor/rules/keystone-<topic>-<name>.mdc`. (No orphans either direction.)

If any check fails, the action is not done — return to the corresponding activity.

## Context-only questions

`keystone init` at 2.0 only asks for the agent target. Everything else
(language, framework, CI, region map, test approach, compliance scope,
starter packs) is detected by **bootstrap** from the actual codebase.

A small set of questions the codebase **can't** answer should be asked
of the user during bootstrap:

- **Aspirational patterns** — what code style / architecture is the
  team moving *toward*, separate from what the code currently does?
- **Methodology** — TDD, BDD, none? Code-review posture? Pacing
  preference (paired / solo / autopilot)?
- **In-flight migrations** — anything mid-port (framework upgrade,
  monolith → services, language change)? Record in
  `corpus/state/CODEBASE_STATE.md` so the agent doesn't surprise the
  user with rules that contradict the migration plan.
- **Code-state assessment** — known debt, deprecated subsystems,
  do-not-touch areas. Different from "what exists" — closer to "what
  do you wish wasn't there."
- **Compliance scope** — SOX / HIPAA / SOC 2 / etc.? Determines
  which starter rules apply.

These are the only items worth a prompt. Detect everything else.

## Index refresh

After every state-file write + every idiom-guide / corpus seed, run `keystone index` so `.keystone/INDEX.json` reflects the new primitives. The `keystone:index` skill wraps the CLI invocation. If skills, subagents, or commands were added, also run `keystone project` to regenerate `.claude/` host projections.
