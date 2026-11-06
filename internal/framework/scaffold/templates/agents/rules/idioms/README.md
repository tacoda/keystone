# Idiom rules

Stack-specific rules extracted from `corpus/idioms/<stack>/`. Each rule file traces to its corpus counterpart, which holds the reasoning, examples, and review checklist.

## Layout

One folder per stack, mirroring `corpus/idioms/`:

```
guides/idioms/
├── README.md
├── <stack-1>/
│   └── <pattern>.md       # rules only
└── <stack-2>/
    └── ...
```

## Empty by default

This directory ships **empty** in a fresh install. The **bootstrap** action populates it on first use, alongside `corpus/idioms/<stack>/`. Until then, only principle rules drive the agent.

## Activation

**Default** — ambient, **lazy by region**. When the agent enters a code region matching a stack, the corresponding `guides/idioms/<stack>/` folder activates. The agent does not load every rule for every edit.

**With `globs:`** — narrows that default. An idiom guide with `globs:` fires only when the stack region matches **and** a touched file matches at least one glob. Useful when an idiom covers only a sub-tree within a stack:

```markdown
---
globs:
  - "apps/admin/**"
---
# Admin form validation — rules
```

The above fires inside the TypeScript region only when files under `apps/admin/**` are touched. Globs can never make an idiom fire outside its stack region — see [`../README.md`](../README.md) for the narrow-only contract.

How "lazy by region" is enforced is agent-specific — see `charter/adapters/<your-agent>/activation.md`.

## Format

Each rule file:

```markdown
# <Idiom Name> — rules

The rules from [`corpus/idioms/<stack>/<idiom>.md`](corpus/idioms/<stack>/<idiom>.md).

## IRON LAW

Non-negotiable for this stack. (Optional — only when there is a true must.)

## GOLDEN RULE

Ideals. Deviation requires reasoning.
```

## Changes when

The stack does. Adding a new framework, upgrading a major version, adopting a new pattern → a rule file changes or appears (in both `corpus/` and `guides/`).
