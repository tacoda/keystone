# Idioms

Stack-specific patterns that express the principles in this stack's vocabulary. Each idiom file traces back to the principle it instantiates.

## Layout

One folder per stack:

```
idioms/
├── README.md
├── <stack-1>/
│   ├── <pattern>.md
│   └── anti-patterns.md
└── <stack-2>/
    └── ...
```

Stacks are added by the **bootstrap** action (detects the project's stack and scaffolds the matching folder) or by hand. Common stack names: `php-laravel`, `typescript-react`, `python-django`, `ruby-rails`, `go`, `rust`, `elixir-phoenix`, `claude-code-plugin` (for projects authoring a Claude Code plugin).

## Empty by default

This directory ships **empty** in a fresh install. The **bootstrap** action populates it on first use. Until then, only the principles layer drives the agent.

## Activation

Ambient, **lazy by region**. When the agent enters a code region matching a stack, the corresponding `idioms/<stack>/` folder activates. The agent does not load every idiom for every edit.

How "lazy by region" is enforced is agent-specific — see `harness/adapters/<your-agent>/activation.md`.

## Authorship

Lead engineer drafts; agent refines through Learning flywheel cycles. Idioms accumulate over the project's lifetime.

## Format

An idiom is split across two files — informational in `corpus/`, prescriptive in `guides/`:

**`corpus/idioms/<stack>/<idiom>.md`** — the explanation:

```markdown
# <Idiom Name>

One-paragraph statement of the pattern in this stack.

> **Rules extracted:** [`guides/idioms/<stack>/<idiom>.md`](guides/idioms/<stack>/<idiom>.md).

## How to apply

Concrete steps, code shape, anti-patterns to avoid.

## Review checklist

Bullet list a reviewer can run through.

**Traces to:** [Principle Name](principles/<file>.md)
```

**`guides/idioms/<stack>/<idiom>.md`** — the rules:

```markdown
# <Idiom Name> — rules

The rules from [`corpus/idioms/<stack>/<idiom>.md`](corpus/idioms/<stack>/<idiom>.md).

## IRON LAW

Non-negotiable for this stack. (Optional — only when there is a true must.)

## GOLDEN PATH

Ideals. Deviation requires reasoning.
```

Idiom files ship with real content. Stack-specific facts are inferred by the **bootstrap** action from the project's manifest, build files, and code — not parked behind placeholders.

## Changes when

The stack does. Adding a new framework, upgrading a major version, adopting a new pattern → an idiom file changes or appears.

## Anti-patterns for this layer

- An idiom that does not cite a principle.
- A "principle" pretending to be an idiom because the team likes a particular way of doing things.
- An idiom that names a specific function or file from the codebase — that is state, not knowledge.
