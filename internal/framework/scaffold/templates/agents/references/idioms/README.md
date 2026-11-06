# Idiom references

Stack-specific patterns: how this stack expresses good practice. Each file explains one idiom. Its paired rule in `.agents/rules/idioms/<stack>/` carries the directives.

## Layout

One folder per stack, mirrored under `rules/idioms/`:

```
references/idioms/
├── <stack>/
│   └── <idiom>.md     # the reasoning
rules/idioms/
├── <stack>/
│   └── <idiom>.md     # the directives; corpus: names the reference
```

This folder ships empty. The `bootstrap` command detects the stack and seeds it, or you add a stack by hand with `keystone new rule idioms/<stack>/<idiom>`.

## Format

```markdown
---
id: idioms/<stack>/<idiom>
description: One line on the reasoning this file holds.
---

# <Idiom name>

One paragraph that states the pattern in this stack.

## How to apply

Concrete steps and code shape.

## Anti-patterns

What to avoid.
```

## Anti-patterns for this folder

- An idiom that names a specific function or file in the codebase. That is state, not an idiom.
- A team preference written up as if it were a universal principle.
