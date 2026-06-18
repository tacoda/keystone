---
kind: corpus
id: corpus/state/INSTALL_PROFILE
description: 'Selections captured by `keystone init`; read by the bootstrap action.'
---
# Install Profile

Selections captured by `keystone init`. Read by the **bootstrap** action; safe to edit by hand. Machine state (keystone version, agents, policies) lives in [`.keystone/lockfile.json`](.keystone/lockfile.json) at the repo root.

## Selections

| Category | Value(s) |
|---|---|
| agent | _(set by `keystone init`)_ |
| app-type | _(unset)_ |
| architecture | _(unset)_ |
| testing | _(unset)_ |
| compliance | _(unset)_ |
| starter | _(unset)_ |
