---
kind: command
id: bootstrap
description: 'One-time initial charter scaffold.'
---
# bootstrap

**One-time initial charter scaffold.** Detect the project's stack, seed `.agents/references/state/`, scaffold idiom directories with code-grounded `paths:`, inventory computational rules, classify hooks, and generate the globs index. Run once per project.

## Activities

Every activity below produces a concrete file write. Detecting, narrating, or summarizing is not enough — call the edit primitive and land the change before moving on.

1. **Detect the stack.** Inspect the repo (`package.json`, `go.mod`, `pyproject.toml`, `Gemfile`, `Cargo.toml`, `requirements.txt`, `build.gradle`, etc.) and list every primary language, framework, and notable library.
2. **Seed `.agents/references/state/CODEBASE_STATE.md`.** Propose an edit that replaces every template placeholder with real values:
   - Detected stacks
   - Tool commands (lint / type-check / test / build / coverage) — actual commands, not placeholders
   - Region map (top-level directories → which stacks they hold). This becomes the source of truth for the `paths:` in step 3
   - CI platform (GitHub Actions / GitLab CI / etc.)
3. **Seed stack idioms with `paths:`.** For each detected stack, scaffold `.agents/references/idioms/<stack>/` and the paired `.agents/rules/idioms/<stack>/`. The rules directory gets a `README.md` describing the stack. Each seeded rule's frontmatter declares `paths:` derived from the region map written in step 2. The paths reflect real code paths, not invented patterns. Populate idiom files as patterns emerge; new files inherit the region's `paths:` unless they cover a sub-tree (in which case narrow them).
4. **Inventory computational rules.** Record LSP, formatter, and editor enforcement (`.editorconfig`, pre-commit hooks, etc.) under `.agents/rules/computational/`. Each entry records `paths:` set to the paths the tool actually covers (read from its config file). The **stack-drift** hook uses these to compare documented vs. effective configuration.
5. **Classify hooks.** For each hook in `.agents/hooks/`, propose an edit to `CODEBASE_STATE.md` recording whether this host can run it. Inferential hooks (`review-functional`, `review-security`, `review-risk`, `review-deployment`, `spec-adherence`) and computational hooks get separate sections.
6. **Generate `.agents/references/state/GLOBS_INDEX.md`.** Walk every rule under `.agents/rules/` and `.keystone/policies/*/rules/`. Read each rule's `paths:` frontmatter, and write the reverse-index (path pattern → list of rules claiming it). Touch only the `## Index` table; preserve everything else. The index ships with an empty `<pattern>` placeholder row, and bootstrap replaces it with real data.
7. **Project to each host.** Run `keystone index && keystone project`. The project step writes each host's rule files from each rule's `paths:`, such as `.cursor/rules/*.mdc` and `.claude/rules/`. On a host with no path-scoped rule format, the agent finds each rule in `.keystone/INDEX.lite.json`.

## Iron law

**No silent overwrites of state files.** Propose every diff via the agent's edit primitive; let the user accept or reject each. **Narration is not a write** — bootstrap is incomplete until each file change has actually landed on disk.

## Completion check

Before claiming bootstrap is done, verify empirically:

- `grep -E '<[^>]+>' .agents/references/state/CODEBASE_STATE.md` returns no template placeholders.
- The `last_reconciled` front-matter value in `CODEBASE_STATE.md` is today's date.
- `.agents/references/idioms/<stack>/` exists for each detected stack.
- `.agents/rules/idioms/<stack>/` exists, paired with each reference folder above.
- Every seeded `.agents/rules/idioms/<stack>/*.md` declares `paths:` in its frontmatter, and those patterns match paths that actually exist in the repo.
- `.agents/references/state/GLOBS_INDEX.md` exists and its `## Index` table lists at least one row per seeded idiom rule (or is explicitly empty if no rules declare `paths:` yet).
- `keystone project` ran after the last rule write, and `keystone verify` passes.

If any check fails, the command is not done. Return to the corresponding activity.

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
  `.agents/references/state/CODEBASE_STATE.md` so the agent doesn't surprise the
  user with rules that contradict the migration plan.
- **Code-state assessment** — known debt, deprecated subsystems,
  do-not-touch areas. Different from "what exists" — closer to "what
  do you wish wasn't there."
- **Compliance scope** — SOX / HIPAA / SOC 2 / etc.? Determines
  which starter rules apply.

These are the only items worth a prompt. Detect everything else.

## Index refresh

After every state-file write + every idiom rule / reference seed, run `keystone index` so `.keystone/INDEX.json` reflects the new primitives. The `keystone-index` skill wraps the CLI invocation. If skills, subagents, or commands were added, also run `keystone project` to regenerate `.claude/` host projections.
