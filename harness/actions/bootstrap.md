# bootstrap

**One-time initial harness scaffold.** Detect the project's stack, seed `corpus/state/`, scaffold idiom directories, inventory computational guides, and classify sensors. Run once per project.

## Activities

1. **Detect the stack.** Inspect the repo (`package.json`, `go.mod`, `pyproject.toml`, `Gemfile`, `Cargo.toml`, `requirements.txt`, `build.gradle`, etc.) and list every primary language, framework, and notable library.
2. **Seed `harness/corpus/state/CODEBASE_STATE.md`.** Record:
   - Detected stacks
   - Tool commands (lint / type-check / test / build / coverage) — actual commands, not placeholders
   - Region map (top-level directories → which stacks they hold)
   - CI platform (GitHub Actions / GitLab CI / etc.)
3. **Seed stack idioms.** For each detected stack, scaffold `harness/corpus/idioms/<stack>/` and the paired `harness/guides/idioms/<stack>/`. Start with a `README.md` describing the stack; populate idiom files as patterns emerge.
4. **Inventory computational guides.** Record LSP, formatter, and editor enforcement (`.editorconfig`, pre-commit hooks, etc.) under `harness/guides/computational/`.
5. **Classify sensors.** For each sensor in `harness/sensors/`, record in `CODEBASE_STATE.md` whether this adapter can run it. Inferential sensors (`review-functional`, `review-security`, `review-risk`, `review-deployment`, `spec-adherence`) and computational sensors get separate sections.

## Iron law

**No silent overwrites of state files.** Propose every diff via the agent's edit primitive; let the user accept or reject each.

## When this is done

`CODEBASE_STATE.md` lists every guide and sensor wired up, and the next session of any action can read it as ground truth.
