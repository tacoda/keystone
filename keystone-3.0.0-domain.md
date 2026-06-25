# keystone 3.0 — the domain (FROZEN)

Keystone keeps its own authoring vocabulary and **maps** each abstraction to
host concept(s). The mapping is deliberately **1-to-many**: one keystone
abstraction can project to different host mechanisms depending on its nature.
That is the framework's value — not a 1:1 rename.

## Governing principle

> Use the **host's name** when keystone's concept is identical to the host
> primitive. Keep a **keystone name** only when the abstraction adds structure
> or maps to more than one host mechanism.

So `command`/`skill`/`agent` take the host names (identical concepts);
`guide`/`sensor`/`playbook` keep keystone names (they add structure or fan out).

## The kinds

| keystone kind | projects to (host) | why the name / notes |
| --- | --- | --- |
| `guide` | **rule** *(inferential)* \| **hook** *(computational, e.g. LSP/formatter)* | keystone name — **1-to-many** by `mode:`; richer than a bare rule (paired `corpus`, `tier`). An inferential guide carries a `tier:` — **iron-law \| golden-rule \| preference** (default) |
| `sensor` | **hook** *(computational)* \| **agent** *(inferential)* | keystone name — **1-to-many** by `mode:` |
| `hook` | settings.json *(host phase)* \| keystone-fired *(framework event)* | keystone's **framework hook layer** — binds an event to an action. Earns its kind via framework events + the fire dispatcher the bare host hook lacks (not a raw passthrough) |
| `command` | command | host name — identical concept (a unit of work / lifecycle step) |
| `skill` | skill | host name — identical concept (a single capability) |
| `agent` | subagent | a role spawned as a subagent (was `persona`/`subagent`); dir `agents/` |
| `playbook` *(alias **workflow**)* | skill (orchestrator) | keystone name — a composed sequence of commands with human `gates:`. Subagent spawning / parallelism is prose in its SKILL.md (the host agent executes it); deterministic parallel fan-out lives in the hook layer, not a playbook field |
| `pattern` | — | keystone-native **prose** — a reusable **documentation pattern** (the Diátaxis modes: tutorial, how-to, reference, explanation). On-demand, glob-scoped for discovery; prescribes how to structure a kind of doc. Referenced by guides/authors as "write this as pattern X" |
| `corpus` | — | keystone-native prose — the *reasoning / why*; on-demand |
| `document` | — | keystone artifact — governed output with `gates:` (plan/review/adr/retro/feature) |
| `concern` | — | composition mixin (inlined frontmatter + corpus routing) |
| `posture` | settings.json permissions | tool/permission posture (allow/ask/deny) — governs *tool access* |
| `tool` | MCP server \| plugin \| CLI | author-defined callable the agent invokes on demand (Claude's generic "tool") — `transport:` picks how it reaches the agent; `run:` + typed `args:` |
| `eval` | — | keystone-native eval harness |
| `source` | — | an **external system** referenced for context via a fetch (curl / HTTP, or a connector) — Slack, Jira, Linear, Confluence, GitHub. Read-side. Not files, not a tool (an invocable callable) |

**corpus vs pattern:** corpus = the *why* behind a specific primitive
(reasoning, tied via `corpus:`); pattern = a reusable *documentation pattern*
in prose (a Diátaxis doc shape — tutorial / how-to / reference / explanation).
Both prose, on-demand; corpus is primitive-specific reasoning, pattern is a
standalone doc-structure prescription.

**No escape hatches; no raw passthroughs.** The keystone kinds + `mode:` cover
every host primitive — `command`/`skill`/`agent` *are* the host primitive.
**`rule` is not a kind** — it's a projection-target name (the output of an
inferential `guide`). Author a `guide`; you never author a `rule`.

**`hook` is a kind — but as a framework abstraction, not an escape hatch.**
keystone owns a hook *layer*: a `hook` binds an **event** to an **action**.
The events span host phases (`PreToolUse`…, projected to settings.json) **and**
keystone's own workflow events (`pre-command`, `on-gate`, `pre-verify`…, fired
by keystone). It earns its kind via the framework events + the fire dispatcher
+ INDEX discovery a bare host hook lacks — a raw host-passthrough hook (no
keystone model over it) stays banned.

## Mode — computational vs inferential (cross-cutting)

`computational` vs `inferential` is a **general dimension** (a `mode:` field),
not sensor-only. It picks the host mechanism; the kind picks intent/activation:

|              | `mode: computational` (`run:` a shell command/script) | `mode: inferential` (LLM / prose / dispatch an `agent`) |
| ------------ | --- | --- |
| **`guide`** (ambient standard, glob-activated) | **hook** — LSP / formatter (gofmt, gopls), e.g. PostToolUse | **rule** — prose directive shim |
| **`sensor`** (phase-gated check) | **hook** — script at a gate (test, build, lint) | **agent** — review, returns structured results |

So `mode` chooses computational→**hook** vs inferential→**rule** (guide) /
**agent** (sensor). Default: `guide` → inferential, `sensor` → computational.

A computational `guide`/`sensor` *projects to* a host hook (the `hook` row in
the table above). The `hook` **kind** is the authorable layer on top of that
same target — you reach for it directly when you want a framework-event binding
(`on-gate`, `pre-verify`) that no glob-scoped guide/sensor expresses.

**Structural (not primitive kinds):**
- `policy` *(alias **plugin**)* — vendored shared content pulled in via
  keystone.json; refined by the project layer through the cascade. The
  engine/gem mechanism.
- `adapter` — the per-host shim triple (`activation`/`lifecycle`/`sensors`).

## Workbook reconciliation (aliases)

| Workbook / Claude term | keystone |
| --- | --- |
| Workflow (named sequence + human gates) | `playbook` |
| plugin | `policy` |
| Document (Plan/Review/ADR/retro) | `document` |
| command / skill / subagent | `command` / `skill` / `agent` |
| tool posture / permissions | `posture` |
| CI guard / hook | `sensor` → hook, or a `hook` (framework layer) |
| review fleet | `sensor` → agent / `agent` |

## Tools — mostly cross-cutting, one real kind

"Tool" splits four ways. Three are cross-cutting concerns, never authored as a
primitive; the fourth earns a kind.

- permissions (allow/ask/deny) → `posture` (cross-cutting)
- per-role scoping → `tools:` frontmatter on `agent` (cross-cutting)
- built-in / MCP-server tools → host/server-provided; referenced, never defined
- **author-defined callable → `tool` kind.** A callable the agent invokes *on
  demand* with a typed input — keystone's generic sense of "tool" (as Claude
  uses the word). Its `transport:` is one of **cli | mcp | plugin**: a plain
  CLI, an MCP-server registration, or a plugin. Distinct from a `sensor` (fires
  automatically at a gate) and a `skill` (prose the agent reads): a tool is a
  programmatic, schema'd callable. Model over row: keystone adds the input
  schema, agent scoping, posture integration, INDEX discovery, and the
  transport binding the bare callable lacks.

**`source` vs `tool` vs MCP** — easy to conflate, kept distinct:
- `source` — an **external system referenced for context via a fetch**
  (curl / HTTP, or the configured connector). Read-side; the EXTERNAL
  escalation layer of the resolution flow (rules → corpus → source). Configured
  in `context.json`; reached via `keystone source query`. Explicitly **not**
  local files, **not** a tool.
- `tool` — **wraps an invocable callable**: an `mcp` server, a `cli`, or a
  `plugin` (`transport:`). Action-side, with a typed input.
- **MCP** — one possible tool **transport**, never a kind. The same external
  system can appear two ways: as a `source` (fetched read-side, e.g. curl its
  API for context) **and** as a `tool` wrapping its MCP/CLI (invoked to act).
  Two separate primitives, different intent — read context vs invoke action.

## Type-aware projection (the compiler)

Projection reads a primitive's nature, not just its kind:
- `sensor` + `mode: computational` → runs its `run:` script; verdict from exit code + stdout
- `sensor` + `mode: inferential` → **agent** dispatch; the agent **must** return a **structured result** (a `returns:`-schema'd object — findings + verdict), never free prose. The dispatcher validates against the schema, rejects non-conforming output, and surfaces it as feedback
- `guide` + `mode: inferential` (with globs) → `.claude/rules/` shim
- `guide` + `mode: computational` → host **hook** (LSP / formatter command, e.g. gofmt PostToolUse)
- `hook` + host-phase event → settings.json entry; `hook` + framework event → no projection (keystone-fired via `keystone hook fire`)
- `command`/`skill`/`agent` → their host files
- `tool` → bound per `transport:` — `cli` (the `run:` script is the callable), `mcp` (registered on keystone's MCP server at startup), or `plugin`; no host file
- `posture` → settings.json permissions block
- `pattern`/`corpus`/`document`/`concern`/`eval`/`source` → no projection (prose / read via INDEX / on-demand)

**Projected naming.** Every projected host artifact is **kebab-case and
`keystone-`-prefixed** — `command review` → `.claude/commands/keystone-review.md`,
invoked `/keystone-review`; skills/agents/rule-shims likewise. The harness owns
a clear namespace; ids already in the keystone namespace are not
double-prefixed (`keystone:index` → `keystone-index`).

**Inferential activation is framework-fired.** Inferential guides and
inferential sensors fire via the **framework hook layer**, not host-ambient
loading — keystone controls *when* they activate (host-agnostic). The
projections above are the artifacts the hook layer fires: an inferential
sensor's `agents/` file is the subagent a `pre-verify`/`on-review` hook
dispatches; an inferential guide's rule shim is the directive a
`pre-command`/post-edit hook surfaces on a glob match.

## Canonical directories (under `.harness/`)

```
guides/  sensors/  hooks/  commands/  skills/  agents/  playbooks/
patterns/  corpus/  documents/  concerns/  posture/  tools/  evals/  sources/
adapters/  policies/(vendored)
```

## Frontmatter associations (unchanged from prior 3.0 work)

`corpus:` (cite reasoning) · `includes:` (compose concern) · `produces:` /
`consumes:` (document graph) · `gates:` + instance `gate:` · `supersedes:` ·
`type:` (subtype, e.g. document `type: feature`) · `mode:` (new —
computational | inferential) · `event:` (new — `hook` binding: host phase or
framework event) · `run:` (new — shell command/script for a computational
hook/sensor/tool; **not** the `command` kind) · `agent:` (new — the agent an
inferential hook/sensor dispatches) · `returns:` (new — the structured-result
schema an inferential sensor/hook's agent must emit; the dispatcher validates
and surfaces it as feedback) · `tier:` (inferential guide authority —
iron-law | golden-rule | preference, default preference) · `transport:` (new —
a `tool`'s binding: cli | mcp | plugin) · `allow:`/`ask:`/`deny:` (a `posture`'s
permission lists).

## What this means for the in-flight rewrite

The earlier 3.0 work renamed everything to community names (guide→rule,
sensor→hook, action→command, playbook→skill, persona→agent) and **removed** the
keystone vocabulary. That was wrong (it flattened the 1-to-many mapping). The
rewrite:

**Keep (already built, correct):** `.harness/` root, kind inference (CoC),
frontmatter associations, `document` primitive + `keystone document`
subcommand, `init` mkdir, the migration *machinery*.

**Revert:** the vocabulary collapse — restore `guide`/`sensor`/`playbook`;
rename `action`→`command`; `persona`/`subagent`→`agent` (keep the `agent`
name); `skill` stays. **Drop `rule` as a kind** (projection target only).

**Add:** `mode:` + type-aware `sensor → hook|agent` projection · `hook` as a
first-class **framework hook layer** (host phase | framework event → `run:`
shell script or `agent:` dispatch; all hooks framework-layer, no per-hook
adapter mapping, single host bridge) · `pattern` (prose kind) · `posture` (+
settings.json projection) · `tool` (callable → keystone MCP server) · tightened
`source` (external systems only) · playbook `gates:` · workflow-retro document
type · `policy`/`playbook` aliases in docs.
