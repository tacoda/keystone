# Evals

Each folder here is an eval: a test that the charter behaves the way you expect. Use one when you add, scope, or prune a rule. An eval doesn't run on its own, so run it yourself or add `keystone eval run` to CI.

## Shape

```
.agents/evals/<id>/
├── EVAL.md          # frontmatter + the scenario in prose
├── expected.json    # the assertions
└── fixture/         # optional; the runner doesn't read it today
```

```yaml
---
id: billing-scope
description: The billing rule applies to billing code and stays out of the web code.
level: static          # static | sensor | agent; default static
---
```

## Levels

| Level | What it checks | Status |
|---|---|---|
| `static` | Each rule's `paths:` against `touched_files` in `expected.json`. | Runs |
| `sensor` | A hook's command: its exit code and output. | Runs |
| `agent` | An agent's behavior with the charter loaded. | Skipped: not built yet |

A rule without `paths:` never counts as fired in a static eval, so assert only scoped rules.

## Run

```
keystone eval run                    # every eval, Markdown report
keystone eval run --filter billing   # ids that contain "billing"
keystone eval run --baseline main    # compare against a git ref
keystone new eval <id>               # scaffold a new eval folder
```

With `--baseline`, an eval missing from the ref shows as new, not as a regression. Commit `EVAL.md` and `expected.json`.
