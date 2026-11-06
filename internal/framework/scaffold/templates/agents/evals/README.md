# Evals

Framework primitive. Each eval measures how the charter behaves
against a known scenario — useful when adding, pruning, or
re-severing rules to confirm the change actually does what you
expect.

## Shape

```
.charter/evals/<id>/
├── EVAL.md          # canonical primitive frontmatter + scenario prose
├── fixture/         # snapshot of code the eval runs against (optional)
└── expected.json    # assertions
```

Each EVAL.md declares canonical primitive frontmatter:

```yaml
---
kind: eval
id: rule-fires-on-billing-touch
description: When a touched file matches src/billing/**, the billing guide must activate.
level: static                # static | sensor | agent
deps:
  - guide/idioms/billing/billing-rules
---
```

## Levels

| Level   | What it checks                                                            | Cost     | Deterministic |
| ------- | -------------------------------------------------------------------------- | -------- | ------------- |
| static  | INDEX.json + cascade + glob match given a fixture file set.               | seconds  | yes           |
| sensor  | Computational sensors against the fixture diff; exit codes + output.      | seconds  | yes           |
| agent   | LLM-driven scenario w/ charter loaded; judge-graded outputs. (Phase 2.1.) | $ per run | no            |

## Running

```
keystone eval run                  # all evals at static + sensor levels
keystone eval run --filter <pat>   # subset
keystone eval run --report md      # human report (default: json)
keystone new eval <id>             # scaffold a new eval dir
```

The dashboard at `/evals` provides one-click runs + SSE-live result
streams.

## Committing evals + the baseline contract

`EVAL.md` and `expected.json` are the spec — small, version-controlled,
always committed. `keystone eval run --baseline <git-ref>` checks out
the ref into a worktree; if EVAL.md / expected.json aren't in that
ref's tree, the baseline diff reports those evals as `removed` (not a
regression).

`fixture/` is the negotiable part. If it carries heavy code snapshots
or regenerable artifacts, you can gitignore it:

```gitignore
# in your project's .gitignore
.charter/evals/*/fixture/
```

Tradeoff: baseline diffs against a ref where the fixture is missing
will skip sensor-level assertions that read fixture files. Static
assertions still run because they only need `touched_files` from
`expected.json`. If fidelity matters for sensor evals, leave the
fixture committed.

## Packaging in policies

Evals are a framework abstraction; policies (vendored charter
fragments) can ship them. Org policies often bundle evals alongside
their guides so consumers can verify the policy's contract is met
locally.

See [`docs/ports/primitive.md`](../../../docs/ports/primitive.md) for
the canonical descriptor shape.
