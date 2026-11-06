# Idiom rules

Stack-specific rules. Each file carries the directives for one idiom. Its paired reference in `.agents/references/idioms/<stack>/` holds the reasoning and examples.

## Layout

One folder per stack, mirrored under `references/idioms/`:

```
rules/idioms/
└── <stack>/
    └── <idiom>.md     # directives only
```

This folder ships empty. The `bootstrap` command detects the stack and seeds it, or you add one with `keystone new rule idioms/<stack>/<idiom>`.

## Scope with paths

Give each idiom rule a `paths:` list for the files its stack owns. Then the rule applies only when the agent touches those files:

```markdown
---
id: idioms/typescript/admin-forms
description: Form validation rules for the admin app.
paths:
  - "apps/admin/**/*.ts"
corpus:
  - idioms/typescript/admin-forms
---

# Admin form validation

## Rules

- One directive per line.
```

Without `paths:`, the rule applies everywhere and takes a line in the `AGENTS.md` managed block.

## Changes when

The stack changes. A new framework, a major upgrade, or a newly adopted pattern adds or changes a rule and its reference together.
