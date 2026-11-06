# Cursor — Activation (rules binding)

How the corpus's ambient content (principles, idioms, process docs) gets loaded into Cursor.

## The menu

Cursor's primary instruction surface is `.cursor/rules/*.mdc`. The installer drops a `keystone.mdc` rule with `alwaysApply: true` and a short body pointing at `charter/`:

```mdc
---
description: Project charter pointer
alwaysApply: true
---

This project uses a project charter. The corpus at charter/ defines
principles, idioms, domain rules, codebase state, and the six-phase
workflow. Read charter/README.md before starting work. Cursor
bindings live in charter/adapters/cursor/.
```

That single always-applied rule is the menu. The full corpus is read on demand via the chat, not auto-loaded.

Legacy `.cursorrules` (single-file format) is **not** used — the modern `.cursor/rules/*.mdc` form is required for the lazy-by-region story below.

## Where runtime config lives

```
.cursor/
└── rules/
    ├── keystone.mdc                            # always-applied menu pointer (shipped by installer)
    └── keystone-<topic>-<name>.mdc             # one per guide with `globs:` (written by bootstrap)
```

Lifecycle actions live in `charter/actions/<action>.md` and are invoked via natural language — no per-action `.mdc` files. The single always-applied rule lists every action with a pointer to its playbook, and the agent follows that pointer when the user asks to run an action.

## Ambient loading

Rules with `alwaysApply: true` enter every chat. Rules with a `globs:` pattern enter the chat **when the user edits a file matching the glob** — this is Cursor's native lazy-by-region mechanism, and it is the architectural reason to reach for Cursor over alternatives that lack the same primitive.

Convention this charter follows:

- **Menu pointer** → `alwaysApply: true`, no glob. (One file: `keystone.mdc`.)
- **Per-guide globs projection** → for every guide whose frontmatter declares `globs:`, the **bootstrap** action writes a `.cursor/rules/keystone-<topic>-<name>.mdc` whose `globs:` mirror the guide's. The body is a one-line pointer to the source guide — Cursor auto-attaches the rule when the user edits a matching file, and the agent follows the pointer to read the actual rules. Guides without `globs:` get no `.mdc`; they activate via the menu pointer's chat-on-demand reading.

## Lazy-by-region — native

Cursor's biggest structural advantage over the other adapters in this charter: it can express "load this rule when editing this region" natively, via `globs:` frontmatter on each `.mdc`. The **bootstrap** action exploits this — for every guide that declares `globs:` (idioms, scoped domain rules, etc.), it writes a matching `.cursor/rules/keystone-<topic>-<name>.mdc`, so the right rules enter context only when the right files are being edited. Pointer-style adapters (Claude Code, Codex, Aider) achieve the same effect by reading `corpus/state/GLOBS_INDEX.md` inside the **orient** playbook; Cursor gets it for free from the framework.

## Context-reset primitive

- **New chat** (toolbar button or `Cmd+Shift+L`) — starts a fresh context. Use after any **synthesize** or **audit** writes changed the corpus.
- **Cursor does not have an equivalent of `/compact`** — there is no in-place context compression. Resetting fully is the only path.

After a flywheel write (synthesize/audit), the user must start a new chat so the next turn re-reads the updated `.mdc` files.

## Domain / state / process loading

- `domain/` — always relevant at task start; loaded via chat read at the start of each session.
- `state/CODEBASE_STATE.md` — loaded by the **orient** action playbook.
- `actions/<action>.md` — loaded when the user invokes that action by name.
- `guides/process/<phase>.md` — loaded when the agent enters that phase from inside an action playbook.

## Capability gaps

- **No autonomous tracker integration.** Cursor does not have MCP equivalents per-tracker; for Atlassian/Linear/GitHub Issues the user pastes the card content into the chat or asks the agent to fetch via web.
- **No sub-agent parallelism.** The `review` rule runs sequentially.
- **Single autonomy mode.** The user confirms each tool call in agent mode regardless of the charter's pacing mode.
