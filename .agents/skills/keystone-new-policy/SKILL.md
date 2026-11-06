---
kind: skill
id: keystone:new-policy
description: Scaffold a new policy repo skeleton outside the current project.
triggers:
  - keystone new policy
  - keystone:new-policy
  - /keystone:new-policy
  - bootstrap a policy repo
  - scaffold a keystone policy
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

# keystone:new-policy — scaffold a policy repo

A **policy** is a separately-versioned charter fragment that consumer
projects pin in their `keystone.json` (e.g. an org-wide policy that
ships strict guides). This generator writes a fresh policy repo
skeleton — manifest, charter directory layout, README — that you then
populate and publish.

Unlike the other generators, `keystone new policy` writes to
**`./<name>/`** in the current working directory, not into an existing
charter install.

## Run

From the directory where the policy repo should land:

```
keystone new policy <name>
```

Example:

```
keystone new policy acme-policies
# writes ./acme-policies/{keystone-policy.json, .charter/, README.md, ...}
```

## After scaffolding

1. Fill in `keystone-policy.json` — name, version, required port
   declarations, strict locks.
2. Author the policy's `.charter/` content — guides, sensors, etc.
3. Initialize git and push to a host so consumers can pin the source.
4. Consumers add the policy via `keystone policy add <owner>/<repo>@<version>`.

See the [policy authoring guide](../../../../docs/adr/) for the full
policy lifecycle.
