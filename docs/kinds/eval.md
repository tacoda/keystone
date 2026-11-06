# Eval

<!-- voice-register: informational -->
<!-- voice-english: ste -->

An eval is a test that the charter behaves the way you expect. It's a constraint, and a keystone extension, because no host has a format for it. Use an eval when you add, scope, or prune a rule, to confirm the change does what you meant. The cost is reach: today an eval checks rule scoping and hook commands, and it can't yet check what an agent does.

## Example

`.agents/evals/billing-scope/EVAL.md`:

```markdown
---
id: billing-scope
description: The billing rule applies to billing code and stays out of the web code.
level: static
---

# billing-scope

A change to `src/billing/` must load the billing rule. A change to `src/web/` must not.
```

`.agents/evals/billing-scope/expected.json`:

```json
{
  "touched_files": ["src/billing/invoice.go"],
  "static": {
    "rules_fired": ["domain/billing"],
    "rules_silent": ["idioms/web/components"]
  }
}
```

Then run `keystone eval run`.

## Where it lives

```
.agents/evals/<id>/
├── EVAL.md          # frontmatter + the scenario in prose
├── expected.json    # the assertions
└── fixture/         # optional; the runner doesn't read it today
```

Keystone infers `kind: eval` from the `evals/` folder.

## Frontmatter

| Key | Required | Meaning |
|---|---|---|
| `id` | yes | Stable id. `--filter` matches on it |
| `description` | yes | One sentence on what the eval checks and what counts as a pass |
| `level` | no | `static`, `sensor`, or `agent`. The default is `static` |
| `levels` | no | A list of levels, to run one eval at more than one level |

## Levels

| Level | What it checks | Status |
|---|---|---|
| `static` | Each rule's `paths:` against `touched_files`. | Runs |
| `sensor` | A hook's command: its exit code and output. | Runs |
| `agent` | An agent's behavior with the charter loaded. | Skipped: not built yet |

### Static

For each id in `static.rules_fired`, at least one `paths:` pattern of that rule must match a touched file. For each id in `static.rules_silent`, no pattern may match. An id can be the rule's `id` or `rule/<id>`.

Static matching is simpler than the matching in the rest of keystone. A pattern with `**` matches any path that starts with the text before `**`. A `!` pattern never counts as a match, and it doesn't exclude. A rule without `paths:` never counts as fired.

### Sensor

Each entry in `sensors` names a hook id, an `expect_exit` code, and optional `must_contain` strings. The runner runs the command from the project root.

The runner reads the command from the **first fenced code block in the hook's body**, not from `run:`. A hook from `keystone new hook` has no code block, so its sensor eval fails with "has no shell command block". To test that hook, put its command in a fenced block in the body as well.

```json
{
  "sensors": [
    { "id": "build", "expect_exit": 0 }
  ]
}
```

## Running evals

```
keystone eval run                       # every eval, Markdown report
keystone eval run --filter billing      # ids that contain "billing"
keystone eval run --report json --out evals.json
keystone eval run --baseline main       # compare against a git ref
keystone eval list                      # every eval id
```

`keystone eval run` exits non-zero when any eval fails. With `--baseline <ref>`, it checks out the ref in a git worktree and runs both versions. It then reports each eval as a regression, a fix, new, removed, or stable. It exits non-zero only on a regression. An eval missing from the ref shows as new, not as a regression.

The dashboard has an Evals page at `/evals`.

## What an eval does not do

- An eval doesn't run on its own. Run it yourself, or add `keystone eval run` to CI.
- A static eval doesn't load the charter the way a host does. It checks only `paths:` patterns.
- A sensor eval doesn't use the host bridge, the event payload, or the 50-second limit.
- An `agent` eval doesn't run. Its result is `skip`.

## Failure modes

- **"no primitive" for a rule id.** The id is wrong, or the file isn't a rule. Use the `id` from the rule's frontmatter.
- **A pathless rule fails `rules_fired`.** Static evals count only `paths:` matches. Assert scoped rules only.
- **A sensor eval says "has no shell command block".** The hook keeps its command only in `run:`. Add a fenced block to the body.
- **Every result is `skip`.** The eval declares no assertions, or its level is `agent`.

## Authoring

```
keystone new eval <id>
```

This writes `EVAL.md` with `level: static` and an `expected.json` with empty assertion lists. The `sensors` key in `expected.json` keeps its 4.x name, and it lists hook ids. Don't rename it, because the runner reads only that key.
