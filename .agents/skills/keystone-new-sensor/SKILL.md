---
kind: skill
id: keystone:new-sensor
description: Scaffold a new sensor (automated check) at the canonical path.
triggers:
  - keystone new sensor
  - keystone:new-sensor
  - /keystone:new-sensor
  - add a sensor
  - scaffold a new check
model: sonnet
tools:
  - Read
  - Write
  - Edit
  - Glob
  - Bash
includes:
  - scaffolds-primitive
tags:
  - scaffold
---

# keystone:new-sensor — scaffold a sensor

A **sensor** is an automated check the charter fires inside an action.
Two kinds exist:

- **computational** — deterministic execution (binary, script, library).
  Examples: unit-test runner, type checker, linter, build command.
- **inferential** — produced by agent reasoning. Examples: functional
  review, security review, spec-acceptance walk.

Sensors live at `.charter/sensors/<name>.md` and declare
`kind:` (`computational` or `inferential`) in frontmatter.

## Run

```
keystone new sensor <name> [--kind <kind>]
```

Examples:

```
keystone new sensor lint                          # default kind: computational
keystone new sensor security --kind inferential   # explicit inferential
```

## After scaffolding

1. Fill in `## Command` — for computational sensors, the exact shell
   invocation. For inferential, the prompt the agent runs.
2. Fill in `## Interpretation` — what pass / fail look like.
3. Fill in `## Remediation` — what the agent does on failure.
4. Wire the sensor into the relevant action (`verify`, `check-drift`,
   etc.) via its `deps:` frontmatter or its `## Activities` list.
5. Run `keystone index` to refresh the descriptor surface.

Full port contract:
[`docs/ports/sensor.md`](../../../../docs/ports/sensor.md).
