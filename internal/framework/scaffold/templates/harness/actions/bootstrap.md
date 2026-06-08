# bootstrap

**One-time initial harness scaffold.** Detect the project's stack, seed `corpus/state/`, scaffold idiom directories, inventory computational guides, and classify sensors. Run once per project.

## Activities

Every activity below produces a concrete file write. Detecting, narrating, or summarizing is not enough — call the edit primitive and land the change before moving on.

1. **Detect the stack.** Inspect the repo (`package.json`, `go.mod`, `pyproject.toml`, `Gemfile`, `Cargo.toml`, `requirements.txt`, `build.gradle`, etc.) and list every primary language, framework, and notable library.
2. **Seed `harness/corpus/state/CODEBASE_STATE.md`.** Propose an edit that replaces every template placeholder with real values:
   - Detected stacks
   - Tool commands (lint / type-check / test / build / coverage) — actual commands, not placeholders
   - Region map (top-level directories → which stacks they hold)
   - CI platform (GitHub Actions / GitLab CI / etc.)
3. **Seed stack idioms.** For each detected stack, scaffold `harness/corpus/idioms/<stack>/` and the paired `harness/guides/idioms/<stack>/`. Start with a `README.md` describing the stack; populate idiom files as patterns emerge.
4. **Inventory computational guides.** Record LSP, formatter, and editor enforcement (`.editorconfig`, pre-commit hooks, etc.) under `harness/guides/computational/`.
5. **Classify sensors.** For each sensor in `harness/sensors/`, propose an edit to `CODEBASE_STATE.md` recording whether this adapter can run it. Inferential sensors (`review-functional`, `review-security`, `review-risk`, `review-deployment`, `spec-adherence`) and computational sensors get separate sections.

## Iron law

**No silent overwrites of state files.** Propose every diff via the agent's edit primitive; let the user accept or reject each. **Narration is not a write** — bootstrap is incomplete until each file change has actually landed on disk.

## Completion check

Before claiming bootstrap is done, verify empirically:

- `grep -E '<[^>]+>' harness/corpus/state/CODEBASE_STATE.md` returns no template placeholders.
- The `last_reconciled` front-matter value in `CODEBASE_STATE.md` is today's date.
- `harness/corpus/idioms/<stack>/` exists for each detected stack.
- `harness/guides/idioms/<stack>/` exists, paired with each corpus folder above.

If any check fails, the action is not done — return to the corresponding activity.
