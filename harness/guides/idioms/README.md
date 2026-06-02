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

Ambient, **lazy by region**. When the agent enters a code region matching a stack, the corresponding `guides/idioms/<stack>/` folder activates. The agent does not load every rule for every edit.

How "lazy by region" is enforced is agent-specific — see `harness/adapters/<your-agent>/activation.md`.

## Format

Each rule file:

```markdown
# <Idiom Name> — rules

The rules from [`corpus/idioms/<stack>/<idiom>.md`](../../../corpus/idioms/<stack>/<idiom>.md).

## IRON LAW

Non-negotiable for this stack. (Optional — only when there is a true must.)

## GOLDEN RULES

Ideals. Deviation requires reasoning.
```

## Changes when

The stack does. Adding a new framework, upgrading a major version, adopting a new pattern → a rule file changes or appears (in both `corpus/` and `guides/`).
