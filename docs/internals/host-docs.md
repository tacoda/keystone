# Host docs

<!-- voice-register: informational -->
<!-- voice-english: ste -->

Keystone 5.0 ships no per-host docs. In 4.x, `keystone init` installed three
docs for each host (`activation.md`, `lifecycle.md`, `sensors.md`). They
drifted. 5.0 replaces them with what the code knows about each host:

- `AGENTS.md`, in its keystone block, has one line per enabled host. The line
  says what the host loads by itself and what the agent opens or runs by hand.
- `keystone project` names each gap: hooks a host can't run, and permission
  rules it can't enforce.
- `docs/hosts.md` is the full matrix, with its sources and open items.

`keystone init` installs only `.keystone/hosts/README.md`, which points at
these. To record how your team handles a gap on one host, write a reference
under `.agents/references/hosts/`. To project to a host keystone has no
profile for, run `keystone new adapter <name>` (see `docs/hosts.md`).

A 4.x install keeps its old per-host docs after `keystone migrate up` moves
them to `.keystone/hosts/<host>/`. Keystone doesn't read them. Delete them
when you no longer need them.
