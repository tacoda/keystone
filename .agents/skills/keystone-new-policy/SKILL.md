---
name: keystone-new-policy
description: 'Scaffold a new policy repo skeleton outside the current project. Triggers: keystone new policy; keystone:new-policy; /keystone:new-policy; bootstrap a policy repo; scaffold a keystone policy.'
allowed-tools: Read Write Edit Glob Bash
metadata:
  model: sonnet
  tags: scaffold,composition
---

# keystone:new-policy — scaffold a policy repo

A **policy** is a separately-versioned charter fragment that consumer
projects pin in their `keystone.json` (e.g. an org-wide policy that
ships strict rules). This generator writes a fresh policy repo
skeleton (manifest, one directory per kind, README) that you then
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
# writes ./acme-policies/{keystone-policy.json, README.md, rules/, hooks/, ...}
```

## After scaffolding

1. Fill in `keystone-policy.json` — name, version, required port
   declarations, strict locks.
2. Author the policy's content: rules, hooks, and so on.
3. Initialize git and push to a host so consumers can pin the source.
4. Consumers add the policy via `keystone policy add <owner>/<repo>@<version>`.

The policy lifecycle is in keystone's ADRs:
https://github.com/tacoda/keystone/tree/main/docs/adr
