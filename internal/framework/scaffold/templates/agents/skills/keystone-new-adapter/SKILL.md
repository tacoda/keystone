---
name: keystone-new-adapter
description: 'Define a host keystone has no built-in profile for (.agents/hosts/<name>.json): pick native, link, copy, or index per kind, then project to it. Triggers: keystone new adapter; keystone:new-adapter; /keystone:new-adapter; add a host; support a new coding agent; project to another agent.'
---

# keystone:new-adapter — define a host keystone has no profile for

Keystone projects to ten hosts out of the box. For any other coding agent,
define the host yourself in `.agents/hosts/<name>.json`, and `keystone
project` delivers the charter to it.

## Run

1. Run `keystone new adapter <name>`. Use lowercase letters, digits, and
   hyphens. A built-in host name is refused.
2. Open `.agents/hosts/<name>.json`. For each kind (`skill`, `rule`,
   `command`, `subagent`), set how the host gets it:
   - `native`: the host already reads that kind from `.agents/`.
   - `link` or `copy`: add `"dir"`, the folder the host reads. Keystone
     links (or copies) each skill folder, each path-scoped rule, and each
     command or subagent file there. `link` follows `link_mode`.
   - `index`: `AGENTS.md` lists it, and the agent opens it from `.agents/`.
3. Run `keystone target add <name>`. It adds the host to `adapters` and
   projects.
4. Read the `keystone project` output. A custom host gets no hook bridge and
   no permission translation, so keystone reports both as gaps, and
   `AGENTS.md` tells the agent to run the checks itself.

A custom host gets no format conversion: files land in the same format as
`.agents/`. If the host needs another format, it needs a built-in profile.

`dir` must stay inside the repo and outside `.agents/` and `.keystone/`.
Keystone records each file it writes there and never replaces a file it
didn't write.
