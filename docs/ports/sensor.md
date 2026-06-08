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

Same as other ports: project wins by default; among plugins, plugins nested deeper in `keystone.json` refine the outer plugins they're nested in; a plugin may declare `strict.sensors: [<name>]` to lock the item absolutely so nothing else (project or any other plugin) can override it. Exactly one file loads per `<name>`.

**Depth limit.** Sensors are only allowed at the project layer and at top-level plugins in `keystone.json`. Plugins nested under another plugin that declare `strict.sensors` or ship vendored sensor files fail `keystone verify` with a `DepthViolation`.

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
