# Context budget

<!-- voice-register: informational -->
<!-- voice-english: ste -->

The context budget measures how much charter text an agent loads, per host. It splits the load into two tiers: what loads every session and what loads on demand. Caps in `keystone.json` turn a regression into a failed `keystone verify`. The count is an estimate, not a tokenizer count. Use it to compare files and to set caps, but don't use it to predict cost.

The code lives in `internal/framework/budget/` (`load.go`, `caps.go`, `budget.go`) and `cmd/keystone/budget.go`.

## The two tiers

**Always loaded** is per host, because each host reads different files by itself. The tier has three parts:

- **Root files.** Every host reads `AGENTS.md`. Claude Code also reads `CLAUDE.md`. Aider also reads `.keystone/INDEX.lite.json` through `.aider.conf.yml`.
- **Listings.** For each skill, command, and subagent that the host loads by itself, the host lists its name and description. The budget counts the id and the description, not the body.
- **Pathless rules.** A rule with no `paths:` that the host loads by itself costs its whole body.

The host profile in `internal/framework/projection` decides which kinds a host loads by itself. If the profile gives a kind the `index` or `report` strategy, the agent opens that primitive from `.agents/` when it applies. So that primitive doesn't count toward the host's always-loaded tier.

**On demand** is the same for every host you enable, because every host can reach `.agents/`. It counts the full body of each rule, reference, skill, command, subagent, tool, and document. The report gives a total per kind and lists the largest files first.

## The estimate

`budget.Estimate` counts whitespace-separated words. It's fast and deterministic, and it needs no tokenizer. Real tokenizers count about 1.35 times more on English prose. Every surface prints this caveat. If you need a closer figure, multiply the count by 1.35.

## Caps

To set caps, add a `budgets` object to `keystone.json`:

```json
"budgets": {
  "always_loaded": { "max_tokens": 4000 },
  "reference":     { "max_tokens": 50000, "max_tokens_per_load": 3000 }
}
```

You can use three forms of key:

| Key | `max_tokens` caps | `max_tokens_per_load` caps |
|---|---|---|
| `always_loaded` | each host's always-loaded tier | not used |
| a kind (`rule`, `reference`, …) | that kind's on-demand total | each file of that kind |
| a 4.x key (`guides`, `corpus`, `sensors`, `actions`, `playbooks`) | the mapped kind's total | each file of the mapped kind |

A plural kind (`rules`) and a 4.x kind name (`guide`) map to the same kind. Keystone ignores the 4.x `adapters` key and any other key it can't map. If two keys map to one kind, the stricter value of each field wins. To leave a field uncapped, set it to 0 or leave it out.

## Where it runs

- **`keystone budget [--json] [--dir <path>]`** prints both tiers. For each host, it shows the three largest entries; for the on-demand tier, the five largest. If the charter goes over a cap, it names the scope, the file, the count, and the overage. Then it exits non-zero.
- **`keystone verify`** measures the same report. If the charter goes over any cap, verify fails.
- **`keystone init`** and **`keystone doctor`** print one line per host for the always-loaded tier, plus any exceeded cap. These two only report. They don't fail on a cap.
- **`keystone allocate`** prints each host's always-loaded tier against the `always_loaded` cap.
- **MCP `keystone_budget`** returns the report as JSON. Use it when you want an agent to check the budget.
- **The dashboard** shows it at `/charter/budget`: exceeded caps, the always-loaded tier per host, the on-demand tier per kind, and the largest files.

All of these call `budget.MeasureProject`. It measures every host in `keystone.json` `adapters`, plus Claude Code, which it always measures. Without a `keystone.json`, it measures Claude Code with no caps.

## When a cap fails

If the always-loaded tier is over its cap, run `keystone allocate`. It finds sections of `CLAUDE.md` and `AGENTS.md` that can move into primitives that load on demand. If a kind's total or one file is over its cap, split or trim the files the report lists first.
