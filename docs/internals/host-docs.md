# Host docs

<!-- voice-register: informational -->
<!-- voice-english: ste -->

Host docs are Markdown notes that tell an agent how the charter works on its host. They are prose for the agent and the user to read. No keystone command reads them, and they don't change what `keystone project` writes. For the projection itself, see [`docs/hosts.md`](../hosts.md). That page gives the strategy per host and kind.

## Where they live

The templates live in the scaffold package, under `templates/keystone/hosts/`. `keystone init` copies them to `.keystone/hosts/`, one folder per host plus a `README.md`:

```
.keystone/hosts/<host>/activation.md
.keystone/hosts/<host>/lifecycle.md
.keystone/hosts/<host>/sensors.md
```

Each of the ten hosts gets a folder, such as `claude-code` or `aider`. If your host has no folder, read `_generic`.

## The three files

| File | Answers |
|---|---|
| `activation.md` | What the host reads at session start, how it finds rules and references, and how to reset context. |
| `lifecycle.md` | How to run each lifecycle command (`orient`, `verify`, `learn`, and the rest) on this host. |
| `sensors.md` | How hooks fire here: by themselves, as a command the user runs, or as a checklist. |

After `init` finishes, it prints the path to `lifecycle.md` for each host you chose. Open that file first.

## Add a host doc

Run `keystone new adapter <host>`. It writes the three files under `.keystone/hosts/<host>/` with a heading and a short prompt in each. Fill in each file for your host.

A host doc doesn't add projection support. For that, add a host profile to `internal/framework/projection/hosts.go`.
