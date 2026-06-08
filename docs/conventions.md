# Keystone Conventions

The canonical convention table for 1.0. Every port lists its on-disk
location, frontmatter, activation, and cascade behavior. Generators emit
conformant files; `keystone doctor` enforces these rules against existing
installs.

`<harness-root>` is the harness folder name configured in `keystone.json`
(default `harness`). `<repo-root>` is the project root — the directory
that contains `keystone.json`.

## Ports

### Guide

- **Project path:** `<harness-root>/guides/<topic>/<name>.md`
- **Plugin path:** `<harness-root>/plugins/<plugin>/guides/<topic>/<name>.md`
- **Activation:** Ambient (always loaded).
- **Frontmatter:** none required.
- **Required shape:** H1 title `# <Name> — rules`, body of rules, optional forward-link to paired corpus.
- **Cascade:** project beats plugin; among plugins, first occurrence in `keystone.json` pre-order wins.
- **Strict-able:** yes (per-port `strict.guides: [...]` in keystone.json).
- **Generator:** `keystone new guide <topic>/<name>` (scaffolds guide + paired corpus stub).

### Corpus

- **Project path:** `<harness-root>/corpus/<topic>/<name>.md`
- **Plugin path:** `<harness-root>/plugins/<plugin>/corpus/<topic>/<name>.md`
- **Activation:** On-demand (loaded when a guide forward-links to it).
- **Frontmatter:** none required.
- **Required shape:** H1 title `# <Name> — reasoning`, long-form explanation, back-link to paired guide.
- **Cascade:** project beats plugin; pre-order over plugin tree.
- **Strict-able:** no (corpus is reference, not policy).

### Sensor

- **Project path:** `<harness-root>/sensors/<name>.md`
- **Plugin path:** `<harness-root>/plugins/<plugin>/sensors/<name>.md`
- **Activation:** Invoked inside an action.
- **Frontmatter:** `kind: <computational | drift | coverage | ...>` (required).
- **Required shape:** H1 title `# Sensor: <name>`, sections `## Command`, `## Interpretation`, `## Remediation`.
- **Cascade:** project beats plugin; pre-order; per-port depth limit (`max_depth: 2` by default).
- **Strict-able:** yes.
- **Generator:** `keystone new sensor <name>`.

### Action

- **Project path:** `<harness-root>/actions/<name>.md`
- **Plugin path:** `<harness-root>/plugins/<plugin>/actions/<name>.md`
- **Activation:** Invoked by name (playbook, another action, agent menu, or `keystone <action>`).
- **Frontmatter:** none required.
- **Required shape:** H1 title `# Action: <name>`, sections `## Entry condition`, `## Activities`, `## Exit condition`.
- **Cascade:** project beats plugin; pre-order.
- **Strict-able:** yes.
- **Generator:** `keystone new action <name>`.

### Playbook

- **Project path:** `<harness-root>/playbooks/<name>.md`
- **Plugin path:** `<harness-root>/plugins/<plugin>/playbooks/<name>.md`
- **Activation:** Invoked by name; runs an ordered sequence of actions.
- **Frontmatter:** none required.
- **Required shape:** H1 title `# Playbook: <name>`, sections `## Sequence` (numbered list of action names), `## Halt conditions`.
- **Cascade:** project beats plugin; pre-order.
- **Strict-able:** yes.
- **Generator:** `keystone new playbook <name>`.

### Adapter (per-agent)

- **Project path:** `<harness-root>/adapters/<agent>/{lifecycle,sensors,activation}.md`
- **Plugin path:** `<harness-root>/plugins/<plugin>/adapters/<agent>/...`
- **Activation:** Loaded at session start by the agent.
- **Frontmatter:** none required.
- **Required shape:** three files per agent — `lifecycle.md` (how the agent invokes playbooks/actions), `sensors.md` (sensor invocation), `activation.md` (the agent's menu file content).
- **Cascade:** project beats plugin; pre-order.
- **Strict-able:** yes (per-agent).
- **Generator:** `keystone new adapter <agent>`.

### Flywheel sink

- **Project path:** `<harness-root>/learning/inbox/<date>-<slug>.md`, `<harness-root>/learning/triage.md`, `<harness-root>/archive/<date>-<slug>.md`
- **Activation:** Write-only — written by `learn` and `audit` actions.
- **Frontmatter:** `status: new | reviewed | accepted | rejected` (inbox); `archived_from: <path>, reason: <one line>` (archive).
- **Not part of the cascade.** These are project-owned state, not policy.

## Path conventions

Two rules govern how files refer to each other.

### Rule 1: Inter-harness references are harness-root-relative

When a markdown file inside the harness links to another harness file,
the link is **relative to the harness root**, not to the source file.
Example:

```markdown
<!-- inside <harness-root>/guides/process/spec.md -->

For reasoning, see [`corpus/process/spec.md`](corpus/process/spec.md).
```

The link resolves against the harness root: `<harness-root>/corpus/process/spec.md`.

**Forbidden:** any `../` or `./` segments in inter-harness links.

```markdown
<!-- DO NOT WRITE -->
[reasoning](../../corpus/process/spec.md)
[reasoning](./paired.md)
```

**Why:** harness files are loaded by an agent that knows the harness
root. Relative-to-source paths force the agent to reconstruct file
locations from imperfect path math; harness-root-relative paths are
unambiguous regardless of nesting depth.

### Rule 2: Code references are repo-root-relative

When a markdown file references **code or config files outside the
harness** (the project's source tree, `keystone.json`, the lockfile,
build scripts), the path is **relative to the repo root** — the
directory holding `keystone.json`.

```markdown
<!-- inside <harness-root>/actions/release.md -->

See [`keystone.json`](keystone.json) for the plugin pin list.
See [`src/main.go`](src/main.go) for the entrypoint.
```

The harness root is itself reached via `<harness-root>/...` from this
form (e.g. `harness/guides/process/spec.md` from a code reference).

**Why:** code referenced from a harness file is being read by the agent
to act on the project's source — the agent already operates from the
repo root for those operations.

### Enforcement

- **Generators** emit harness-root-relative inter-harness links and
  repo-root-relative code links by default.
- **`keystone doctor`** flags any inter-harness link containing `../`
  or `./` segments, plus broken forward-links.
- **CI usage:** `keystone doctor` exits non-zero on any path
  violation, so projects can gate merges on it.

## Authoring

Every port has a one-line scaffold command. Generators write a skeleton
with conformant frontmatter, paths, and headings; the author fills in
the body.

```
keystone new guide <topic>/<name>
keystone new sensor <name>
keystone new action <name>
keystone new playbook <name>
keystone new adapter <agent>
keystone new plugin <name>           # scaffolds a new plugin repo
```

The skeletons live in `internal/framework/scaffold/templates/` and are
embedded into the binary. Updating a generator's template changes what
new files look like without affecting already-scaffolded content.

## Sensor depth limit

The `sensors` port has a default `max_depth: 2` — that is, sensors are
permitted at the project layer and at the immediate plugin level, but
not at deeper-nested plugins. Adjustable via `sensors.max_depth` in
`keystone.json` (when added in Phase 5's budgets/limits section).

## Reference

- Port contracts (long form, one per file): [`docs/ports/`](ports/).
- JSON Schemas for config files: [`docs/schemas/`](schemas/).
- Architecture decisions: [`docs/adr/`](adr/).
