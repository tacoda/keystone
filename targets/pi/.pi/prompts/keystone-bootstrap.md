---
description: One-time initial harness scaffold — detect stack, seed state, classify sensors
---

Run the **bootstrap** action of the project harness.

Read `harness/adapters/pi/lifecycle.md` for the action table and `harness/README.md` for the components.

## Activities

1. **Detect the stack** — inspect package.json, go.mod, pyproject.toml, Gemfile, Cargo.toml, build.gradle, etc.
2. **Seed `harness/corpus/state/CODEBASE_STATE.md`** with detected stacks, real tool commands (lint / type-check / test / build / coverage), region map, and CI platform.
3. **Seed stack idioms** — scaffold `harness/corpus/idioms/<stack>/` and `harness/guides/idioms/<stack>/` for each detected stack.
4. **Inventory computational guides** (LSP, formatter, editor enforcement) under `harness/guides/computational/`.
5. **Classify sensors** — record in `CODEBASE_STATE.md` which sensors this adapter can run.

## Iron law

**No silent overwrites.** Propose every state-file diff before applying.
