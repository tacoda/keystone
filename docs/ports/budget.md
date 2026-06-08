# Port: Budget

**Activation:** Read at session start (the agent's ambient load is the sum of every port's tokens) and on demand by `keystone doctor --budget`.
**Purpose:** Make context-window consumption legible. Per-port caps in `keystone.json` make budget regressions visible the moment they happen.

## Where budgets live

Budgets are declared in `keystone.json` under an optional `budgets`
block:

```json
{
  "version": "1",
  "harness_root": "harness",
  "plugins": [],
  "budgets": {
    "guides":    { "max_tokens": 10000 },
    "corpus":    { "max_tokens": 50000 },
    "sensors":   { "max_tokens": 5000 },
    "adapters":  { "max_tokens": 20000 }
  }
}
```

Each entry is a `BudgetSpec`:

| Field                  | Meaning                                                                                 |
| ---------------------- | --------------------------------------------------------------------------------------- |
| `max_tokens`           | Cap on total tokens loaded from this port. 0 (or omitted) means no cap.                 |
| `max_tokens_per_load`  | Per-load cap for on-demand content (corpus). Currently informational; Phase 6 enforces. |

Budgets are advisory: doctor never exits non-zero just because a port
is over its cap. Projects that want strict enforcement wrap the doctor
output in CI scripts, or upgrade specific warnings to errors via their
own tooling.

## Estimator

The 1.0 estimator is **whitespace-approximate**: token count =
`len(strings.Fields(content))`. Fast, deterministic, no external
dependencies. Sufficient for relative comparisons between files and
ports, and for trend lines as the harness grows.

The whitespace count under-counts compared to real model tokenizers
(GPT-style BPE, Claude's tokenizer, etc.) by roughly **25–35% on
English prose**. Multiply by ~1.35 if you want a conservative
"real-tokenizer" estimate.

A future `--tokenizer=tiktoken` opt-in at install time (PLAN §6 open
question 4) will swap in a more accurate count without changing the
budget package's API.

## CLI surfaces

### `keystone doctor --budget`

Walks every `.md` file under `<harness-root>/`, classifies each by
port, and renders a per-port breakdown:

```
budget: per-port token estimate (whitespace-approximate)
  • actions    2955 / no cap
        505  harness/actions/policy-audit.md
        441  harness/actions/README.md
        ...
  • corpus     30715 / 50000  (61% used)
        ...
  ! guides     18893 / 10000
        867  harness/guides/process/release.md
        ...
  ⚠ 1 port(s) over their declared budget — top contributors above
```

Markers:

- `•` port within its declared cap, or no cap declared.
- `!` port over its declared cap; top 5 contributors printed.
- `⚠` closing line — at least one port over budget.
- `✓` closing line — every declared budget within cap.
- `ℹ` closing line — no budgets declared in keystone.json.

Files skipped: README at any depth, `learning/`, `archive/`, anything
under `<harness-root>/plugins/`.

### `keystone init`

After scaffolding, init prints the ambient load:

```
▸ ambient load: 39209 tokens across 6 port(s)
    actions    2955
    adapters   15685
    corpus     3548
    guides     11767
    playbooks  634
    sensors    4620
  (run `keystone doctor --budget` later for top contributors per port)
```

If `keystone.json` already has a budgets block (e.g. when re-running
init in an existing project), per-line `(cap N, X% used)` or
`(over budget by N)` suffixes appear so the user catches budget
regressions on the first run.

## What counts toward the budget

Files under `<harness-root>/<port>/...`, where `<port>` is one of:

- `guides`
- `corpus`
- `sensors`
- `actions`
- `playbooks`
- `adapters`

Files **excluded** from the count:

- `README.md` at any depth — orientation, not loaded by the agent.
- `<harness-root>/learning/...` — flywheel state, not policy.
- `<harness-root>/archive/...` — same.
- `<harness-root>/plugins/...` — vendored content; tracked separately in
  the plugin tree (the cascade picks them up at resolve time).

## Recommendations

- **Guides + adapters are loaded ambient** (every session). Keep their
  caps tight; over-budget here directly hits every session's context
  window.
- **Corpus is on-demand**. Total can be large; what matters is
  `max_tokens_per_load`, which Phase 6 will enforce per-paired-corpus.
- **Sensors are invoked**. Their markdown describes how to invoke;
  it's read but not constantly resident. Looser caps are fine.
- **Set budgets early.** Pick numbers that match your model's window
  and budget-per-prompt, then tune. The budget block is a forcing
  function — without it, harness content grows unboundedly.

## Future evolution

- **Real tokenizer opt-in** (`--tokenizer=tiktoken` at init) — PLAN §6
  open question 4. The budget package's API stays stable; only the
  estimator implementation swaps.
- **Per-plugin attribution** — the report currently lumps plugin
  content under its port. A future split would let users see how much
  budget each individual plugin consumes.
- **CI-friendly output mode** — `--format=json` so projects can pipe
  the breakdown into custom enforcement logic.

## Read by

- `keystone init` (`reportAmbientLoad` in `cmd/keystone/init.go`).
- `keystone doctor --budget` (`runBudgetReport` in
  `cmd/keystone/doctor.go`).
- `internal/framework/budget/` — the estimator and Allocator.
