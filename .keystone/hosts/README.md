# Hosts

Keystone delivers `.agents/` to each host listed in `keystone.json`
`adapters` when you run `keystone project`. Nothing here needs editing.

- `AGENTS.md`, in its keystone block, says for each enabled host what it
  loads by itself and what the agent has to open or run by hand.
- `keystone project` names each gap: hooks a host can't run, and
  permission rules it can't enforce.

To record how your team handles a gap on one host, write a reference under
`.agents/references/hosts/`, so agents open it when it applies.
