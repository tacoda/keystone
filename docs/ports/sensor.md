# Port: Sensor

**Activation:** Invoked inside an action — sensors are computational checks the agent runs to gather facts (lint, type-check, test, build, coverage, drift, debt).
**Purpose:** Automated checks that produce a machine-parseable result the calling action interprets.

## Path convention

```
.charter/sensors/<name>.md                                   # project-owned
.charter/policies/<policy>/sensors/<name>.md                  # policy-owned (read-only)
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

Same as other ports: project wins by default; among policies, policies nested deeper in `keystone.json` refine the outer policies they're nested in; a policy may declare `strict.sensors: [<name>]` to lock the item absolutely so nothing else (project or any other policy) can override it. Exactly one file loads per `<name>`.

**Depth limit.** Sensors are only allowed at the project layer and at top-level policies in `keystone.json`. Policies nested under another policy that declare `strict.sensors` or ship vendored sensor files fail `keystone verify` with a `DepthViolation`.

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
