# Port: Sensor

**Activation:** Invoked inside an action — sensors are computational checks the agent runs to gather facts (lint, type-check, test, build, coverage, drift, debt).
**Purpose:** Automated checks that produce a machine-parseable result the calling action interprets.

## Path convention

```
harness/sensors/<name>.md                                     # project-owned
harness/plugins/<plugin>/sensors/<name>.md                    # plugin-owned (read-only)
```

Flat — no topic directory. Sensors are global by name across the cascade.

## Required shape

```markdown
---
kind: <computational | drift | coverage | ...>
---

# Sensor: <name>

<one-sentence description>

## Command
<the shell invocation; may be templated with project variables>

## Interpretation
<how to parse output; what constitutes pass / fail / warning>

## Remediation
<what the agent should do on failure>
```

- **`kind:` frontmatter** — required. Open set; documented values in `docs/conventions.md`.
- **H1 title** — required. Format: `# Sensor: <name>`.
- **Command / Interpretation / Remediation** — required.

## Cascade behavior

Same as other ports: project wins, then pre-order over plugin tree, `strict.sensors: [<name>]` locks downward. Exactly one file loads per `<name>`.

**Depth limit.** Sensors declare a `max_depth` in the port contract (default: 2 — i.e., only the project layer and one plugin layer may ship sensors). Replaces the 0.x rule that sensors are a team-tier kind. Plugins deeper than `max_depth` that ship a sensor fail `keystone verify`.

## Example

```markdown
---
kind: computational
---

# Sensor: build

Runs the project's build to surface compile errors.

## Command
`go build ./...`

## Interpretation
Exit 0 = pass. Non-zero exit = fail; capture stderr.

## Remediation
On fail, hand the captured stderr to the agent and request a fix before
proceeding past the current action.
```

## Authoring

```
keystone new sensor <name>
```
