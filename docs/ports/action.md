# Port: Action

**Activation:** Invoked by name — by a playbook, by another action, by the agent's menu file, or directly by the user.
**Purpose:** One atomic unit of lifecycle work. Spec, orient, verify, review, learn, audit, bootstrap, mode, synthesize.

## Path convention

```
harness/actions/<name>.md                                     # project-owned
harness/plugins/<plugin>/actions/<name>.md                    # plugin-owned (read-only)
```

Flat — no topic directory. Actions are global by name across the cascade.

## Required shape

```markdown
# Action: <name>

<one-sentence description>

## Entry condition
<what must be true before this action runs>

## Activities
1. <verb + artifact>
2. <verb + artifact>
3. ...

## Exit condition
<what must be true when this action completes>
```

- **H1 title** — required. Format: `# Action: <name>`.
- **Frontmatter** — none required.
- **Entry condition / Activities / Exit condition** — required.

## Cascade behavior

Same as other ports: project wins by default; among plugins, plugins nested deeper in `keystone.json` refine the outer plugins they're nested in; a `strict.actions: [<name>]` declaration locks the item absolutely — nothing else (project or any other plugin) can override it. Exactly one file loads per `<name>`.

## Example

See `harness/actions/spec.md` (scaffolded by `keystone init`) for the canonical example — sourcing a spec from a tracker card or authoring inline, writing acceptance criteria, anchoring downstream phases.

## Authoring

```
keystone new action <name>
```
