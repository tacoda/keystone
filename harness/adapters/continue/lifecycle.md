# Continue — Lifecycle binding

> **Stub.** Fill in the bindings as you wire Continue up.

## Action → invocation

Continue reads `.continuerules` (and `config.json` / `config.yaml` for richer customization). Lifecycle actions are invoked by asking the agent or by configuring custom slash commands in `config.json`.

| Action | Invocation |
|---|---|
| **spec** | Custom slash command or freeform ask. |
| **orient** | Custom slash command or freeform ask. |
| **check-drift** | Freeform ask. |
| **verify** | Freeform ask. Continue can run shell commands via its `cmd` step type. |
| **review** | Freeform ask. Sequential. |
| **learn** | Freeform ask. |
| **bootstrap** | Freeform ask. |
| **audit** | Freeform ask. |
| **synthesize** | Freeform ask. |
| **mode** | Edit `harness/guides/process/modes.md` directly. |

## Capability matrix

| Capability | Supported? |
|---|---|
| Autonomous shell execution | ✓ (`cmd` step) |
| Sub-agent parallelism | ✗ |
| Autonomy levels | ✗ |
| Lazy-by-region | partial (via context providers in `config.json`) |
| Context-reset primitive | new chat |
