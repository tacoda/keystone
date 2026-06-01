# Cursor — Activation (rules binding)

How the corpus's ambient content (principles, idioms, process docs) gets loaded into Cursor.

## The menu

Cursor's primary instruction surface is `.cursor/rules/*.mdc`. The installer drops a `keystone.mdc` rule with `alwaysApply: true` and a short body pointing at `harness/`:

```mdc
---
description: Project harness pointer
alwaysApply: true
---

This project uses a project harness. The corpus at harness/ defines
principles, idioms, domain rules, codebase state, and the six-phase
workflow. Read harness/README.md before starting work. Cursor
bindings live in harness/adapters/cursor/.
```

That single always-applied rule is the menu. The full corpus is read on demand via the chat, not auto-loaded.

Legacy `.cursorrules` (single-file format) is **not** used — the modern `.cursor/rules/*.mdc` form is required for the lazy-by-region story below.

## Where runtime config lives

```
.cursor/
└── rules/
    ├── keystone.mdc          # always-applied menu pointer
    ├── orient.mdc            # manual — invoked at task start
    ├── check-drift.mdc       # manual — invoked after edits
    ├── verify.mdc            # manual — invoked before commit
    ├── review.mdc            # manual — invoked after verification
    └── learn.mdc             # manual — invoked post-merge
```

The installer ships these as **starter scaffolds**. Teams refine them as they discover what each phase should emphasize for their stack.

## Ambient loading

Rules with `alwaysApply: true` enter every chat. Rules with a `globs:` pattern enter the chat **when the user edits a file matching the glob** — this is Cursor's native lazy-by-region mechanism, and it is the architectural reason to reach for Cursor over alternatives that lack the same primitive.

Convention this harness follows:

- **Menu pointer** → `alwaysApply: true`, no glob.
- **Lifecycle action rules** → manual (no `alwaysApply`, no `globs`). The user invokes via `@<action>` when entering that phase.
- **Region-scoped idioms** → `globs: "<path-pattern>"`. *Bootstrap* writes these; they auto-attach when the user opens matching files.

## Lazy-by-region — native

Cursor's biggest structural advantage over the other adapters in this harness: it can express "load this rule when editing this region" natively, via `globs:` frontmatter on each `.mdc`. The **bootstrap** action exploits this — for each detected stack, it writes a `.cursor/rules/idiom-<stack>.mdc` with the matching glob, so the right idioms enter context only when the right files are being edited.

No equivalent mechanism exists in Aider, GitHub Copilot, or Claude Code (which implements lazy-by-region inside the **orient** action instead). Cursor's framework does the work for free.

## Context-reset primitive

- **New chat** (toolbar button or `Cmd+Shift+L`) — starts a fresh context. Use after any **synthesize** or **audit** writes changed the corpus.
- **Cursor does not have an equivalent of `/compact`** — there is no in-place context compression. Resetting fully is the only path.

After a flywheel write (synthesize/audit), the user must start a new chat so the next turn re-reads the updated `.mdc` files.

## Domain / state / process loading

- `domain/` — always relevant at task start; loaded via chat read at the start of each session.
- `state/CODEBASE_STATE.md` — loaded by the **orient** rule.
- `process/<phase>.md` — loaded by the matching action rule when the agent enters that phase.

## Capability gaps

- **No autonomous tracker integration.** Cursor does not have MCP equivalents per-tracker; for Atlassian/Linear/GitHub Issues the user pastes the card content into the chat or asks the agent to fetch via web.
- **No sub-agent parallelism.** The `review` rule runs sequentially.
- **Single autonomy mode.** The user confirms each tool call in agent mode regardless of the harness's pacing mode.
