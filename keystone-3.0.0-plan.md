# keystone 3.0.0 — execution plan (build to domain.md)

`keystone-3.0.0-domain.md` is the **frozen spec** (the *what*). This is the
**build plan** (the *how*). The earlier vocabulary-collapse plan is dead: it
flattened keystone's 1-to-many mapping. We **revert it forward** (new commits,
no history rewrite) on `feat/3.0.0-vocabulary`.

## Net change from current `HEAD` (`f6f94c4`)

Current committed taxonomy = the wrong collapse: `KnownKinds = rule, hook,
command, skill, agent, document, corpus, eval, source, concern`. Restore the
keystone vocabulary and add the missing kinds.

| current (wrong) | 3.0 corrected | role |
| --- | --- | --- |
| `rule` (was guide) | **`guide`** | keystone kind; projects rule\|hook by `mode:` |
| `hook` (was sensor) | **`sensor`** | keystone kind; projects hook\|agent by `mode:` |
| `agent` (was persona/subagent) | **`agent`** | keep the name; absorbs persona+subagent; dir `agents/` |
| — (folded into skill) | **`playbook`** | keystone kind; projects skill orchestrator + `gates:` |
| `command` | `command` | keep (was action+command) |
| `skill` | `skill` | keep |
| `document` `corpus` `concern` `eval` | unchanged | keep |
| `source` | **dropped as a kind** | external-system access is a `tool` (cli=curl / mcp). The context.json query subsystem stays transitionally; removed in a later slice |
| `rule` | **dropped as a kind** | projection-target name only; author a `guide` (`mode: inferential`) |
| `hook` | **kept — as a framework abstraction** | keystone's hook *layer*: binds host phase \| framework event → action. Earns its kind (framework events + fire dispatcher); raw passthrough still banned |
| — (new) | **`pattern`** | keystone prose recipe |
| — (new) | **`posture`** | → settings.json permissions |
| — (new) | **`tool`** | author-defined callable; `transport:` cli \| mcp \| plugin. Absorbs external-system access (former `source`) |

**3.0 `KnownKinds`:** `guide, sensor, hook, agent, command, skill, playbook,
pattern, corpus, document, concern, posture, tool, eval`. (`source` dropped —
external-system access is a `tool`.) **No escape hatches / no raw passthroughs** — the kinds + `mode:` produce every host
primitive. **`rule` is not a kind** — it's a projection-target name; author a
`guide`. **`hook` IS a kind** — keystone's framework hook layer (host phase or
framework event → action), not a raw host passthrough. **New field:** `mode:
computational | inferential`.

## Framework hooks (extensible workflow-event layer)

Host hooks fire only on host events (PreToolUse/PostToolUse/Stop/
UserPromptSubmit/SessionStart/PreCompact) — they see tool calls, not
keystone's workflow. The host system **cannot** fire on keystone lifecycle
points, so the **`hook` kind is keystone's own hook layer** — an authorable
framework abstraction. A `hook` binds an `event:` to an action.

**All hooks live at the framework layer; the adapter does not map them.** Host
hook configs differ too much across hosts to gain from per-host projection
(unlike rules, which map cleanly) — so keystone fires every hook itself and
writes **no per-hook entries** into a host's `settings.json`. Host events reach
the framework through a **single generic bridge** (one settings.json entry per
host phase → `keystone hook fire <phase>`), installed once — not a per-hook
adapter mapping. You don't have to use host events at all; the framework-event
layer stands on its own.

| `hook` event namespace | fired by | reaches keystone via |
| --- | --- | --- |
| **host phase** (`PreToolUse`…) | the host | the single generic bridge → `keystone hook fire <phase>` |
| **framework event** (enum below) | keystone runtime | direct dispatch from the firing subcommand |

**Framework-event registry (closed enum, the extension points):**
`pre-command` · `post-command` · `pre-playbook` · `post-playbook` ·
`on-gate` · `pre-verify` · `post-verify` · `on-phase-enter` ·
`on-phase-exit`. Optional matcher narrows: `phase:` / `command:` / `type:`.

**Fire mechanism (deterministic, no daemon):**
- Auto-fire from real subcommands: `keystone document promote` → `on-gate`;
  `keystone verify` → `pre-verify` / `post-verify`. No agent discretion.
- `keystone hook fire <event> [--phase|--command|--type|--task]` — the one
  dispatch entry point, called by both the host bridge and prose boundaries (a
  playbook step). Dispatches every `hook` matching event + matcher; each runs
  its `run:` shell command/script (computational) or dispatches an `agent`
  (inferential). Non-zero exit blocks. (`run:` is a shell script, **not** a
  keystone `command` — that kind is an agent-driven unit of work.)

**Extensible** = add an enum value + one fire call in core; authors bind hooks
without touching keystone. Decision: **hybrid firing (a)** — auto-fire where a
subcommand boundary exists, explicit `keystone hook fire` in playbook prose
elsewhere. No workflow runtime/daemon (the agent drives the lifecycle by
reading playbook prose; keystone has no orchestrator and gains none here).

## Keep (already built, correct — do not touch)

`.harness/` root · kind inference (CoC) · frontmatter associations
(`corpus`/`includes`/`produces`/`consumes`/`gates`/`gate`/`type`/`supersedes`)
· `document` primitive + `keystone document` subcommand · `init` mkdir · the
migration *machinery* (v3_0 plumbing, `.keystone→.harness` relocation).

## Slices

Each slice ends green (`go test ./...` + `go vet ./...`) and is one commit.

**Slice 1 — taxonomy revert + `mode:`** (`primitive.go`, `lint.go`, `infer`)
- Restore kind consts: `KindGuide` `KindSensor` `KindPlaybook`; add `KindHook` `KindPattern` `KindPosture` `KindTool`. Keep `KindAgent` (name unchanged; absorbs persona+subagent), `KindCommand`/`KindSkill`/`KindDocument`/`KindCorpus`/`KindConcern`/`KindEval`/`KindSource`. **Drop `KindRule`** entirely — `rule` is a projection-target name, not a kind.
- `KnownKinds` + `canonicalDirKind`: `guides/ sensors/ hooks/ commands/ skills/ agents/ playbooks/ patterns/ corpus/ documents/ concerns/ posture/ tools/ evals/ sources/`. **No `rules/` dir.**
- Add to `Frontmatter`: `Mode` (`mode`, default guide→inferential, sensor→computational), `Event`/matcher (`hook` binding), `Run` (`run:` shell action), `Agent` (`agent:` dispatch target), `Returns` (`returns:` structured-result schema for inferential sensors/hooks).
- `lint.go` (added): an inferential `sensor`/`hook` **must** declare `agent:` + `returns:`; a computational one **must** declare `run:` — mutually exclusive. A `guide`'s `tier:` ∈ {iron-law, golden-rule, preference, ""} (empty = preference).
- `lint.go`: validate `mode` ∈ {computational, inferential, ""}; reject `kind: rule` with a "rule is not a kind — author a guide" hint; reject dropped kinds (`action`/`persona`/`subagent`→`agent`) with a migration hint; per-kind required fields.
- → verify: unit tests for `InferKind` on every new dir; lint rejects `kind: rule` and unknown kinds.

**Slice 2 — type-aware projection** (`project.go`, adapters)
- `ProjectionRelPath` / `Project` keyed on (kind, mode). **Adapters project rules + host files only — never hooks** (hooks are framework-layer, slice 5):
  - `guide` + inferential + globs → `.claude/rules/<slug>.md` shim (existing `writeRuleShim`). (`guide` + computational = author a `hook` instead — no rule shim.)
  - `sensor` + inferential → `.claude/agents/<id>.md` (agent dispatch / review fleet); the dispatched agent returns a `returns:`-schema'd structured result, validated before it's surfaced as feedback.
  - `sensor` + computational → **no adapter projection** (`run:` script executed via the hook layer / `keystone verify`).
  - `agent` → `.claude/agents/`; `command` → `.claude/commands/`; `skill`/`playbook` → `.claude/skills/<id>/SKILL.md`.
  - `hook`/`pattern`/`corpus`/`document`/`concern`/`posture`/`eval`/`source` → no adapter file projection (posture handled in slice 3; hooks in slice 5).
- The existing per-sensor `ProjectHooks` mapping + `host_triggers` stay untouched here (still keyed on `KindHook`) — the hook layer replaces them in slice 5. Slice 2 is projection-path correctness only.
- → verify: projection table test — one primitive per (kind,mode) lands at the expected path.

**Slice 3 — `posture` → settings.json** (`posture` projection)
- `posture` primitive frontmatter → permissions (allow/ask/deny) merged into `.claude/settings.json` (same merge path hooks use).
- → verify: posture fixture projects to settings permissions block; idempotent re-merge.

**Slice 4 — `pattern` + `posture` + `tool` authoring** (scaffold, `keystone new`, MCP)
- `keystone new pattern|posture|tool`; MCP `keystone_new_pattern|posture|tool`; INDEX + `keystone new` usage text pick them up.
- `pattern` is prose (a reusable software-design pattern) — scaffold a frontmatter + body stub at the canonical path; no generator, no projection.
- → verify: each `keystone new <kind>` scaffolds at the canonical path with valid frontmatter; index lists the new kinds; lint clean.

**Slice 5 — framework-hook layer** (`hook` kind, event registry, fire)
- `hook` frontmatter: `event:` (host phase | framework event) + matcher + action — `run:` (shell command/script, computational) **or** `agent:` (dispatch an agent, inferential). `run:` is a shell script, not the `command` kind.
- Framework-event enum + `keystone hook fire <event> [--phase|--command|--type|--task]` — the single dispatcher (reads INDEX, runs `run:` / dispatches `agent:`, non-zero blocks). Inferential dispatch validates the agent's output against `returns:`; non-conforming output is an error, not silent passthrough.
- **Parallel fan-out lives here, not in playbooks.** When an event matches multiple inferential sensors/agents, the dispatcher fans them out in parallel and collects their `returns:`-structured findings — keystone runs this, so the structure is real. Playbooks orchestrate in prose (their projected SKILL.md); spawning subagents is author-written prose the host agent executes, no playbook schema.
- Single host bridge: `keystone project` installs one generic `.claude/settings.json` entry per host phase → `keystone hook fire <phase>`. Not per-hook.
- Auto-fire wiring: `keystone document promote` → `on-gate`; `keystone verify` → `pre-verify`/`post-verify`.
- **Inferential guides + sensors are framework-fired here**: the hook layer dispatches an inferential sensor's agent at its event (e.g. `pre-verify`/`on-review`) and surfaces an inferential guide's directive at its event (e.g. `pre-command`/post-edit) on a glob match. This is keystone-controlled activation, replacing reliance on host-ambient loading. The slice-2 `agents/`/`rules/` projections are the artifacts dispatched.
- → verify: a fixture framework-event hook fires on `keystone verify`; non-zero exit blocks; a host-phase hook fires through the bridge; `project` writes the bridge entries but no per-hook entries; an inferential sensor dispatches at `pre-verify`.

**Slice 5b — `tool` transports** (`mcp/`)
- `tool` frontmatter (scaffold landed in slice 4): `transport:` (cli | mcp | plugin) + `run:` (handler) + `args:` schema + `tools:`-style scoping.
- Bind per transport: `mcp` → keystone MCP server reads `kind: tool`/`transport: mcp` from INDEX at startup and registers each (handler shells out to `run:` with validated args); `cli` → the `run:` script is the callable, surfaced for direct invocation; `plugin` → host plugin descriptor. Lint: `transport` ∈ {cli, mcp, plugin, ""}.
- → verify: an `mcp` tool registers + invokes via the MCP server; bad args rejected by schema; `transport` lint rejects an unknown value.

**Slice 5c — remove the `source` subsystem** (`mcp/source.go`, `web/`, `context.json`) — DONE
- The `source` *kind* is already dropped (done — KnownKinds/dir/new/MCP-new). This slice removes the read-side subsystem that backed it: `mcp/source.go` (`keystone_source_list|query|health` + resources), `web/sources_actions.go` + the 5 `source*`/`_sources_*` templates and their refs in `web/cache.go`/`topics.go`/`insights.go`, and the `context.json` document (read loosely as `map[string]any`; nothing else uses it — verify, then kill the file + its handling). Drop the `context.json` relocation line from `migrations/v3_0.go`. Resolution flow collapses to rules → corpus (external context now via a `tool`).
- → verify: `go build ./...` + `go test ./...` green with the source packages gone; no dangling `context.json` / `keystone_source_*` refs; web dashboard renders without the sources pages.

**Slice 6 — migration v3_0 rewrite + symmetric sensor model** (`v3_0.go`, lint, projection, hook layer) — DONE
- Sensor model (decision A→symmetric): guide & sensor each keep `mode:`. Computational guide/sensor + hook all fire deterministically via `event:`+`run:` (unified `primitive.HookFire`); inferential guide → rule shim, inferential sensor → agent dispatch. Closed the "computational sensor fires nowhere" gap.
- Migration: guide→guide, playbook→playbook (collapse-fixes), action→command, persona/subagent→agent, computational sensor (host_triggers) → hook (event/run/mode), review sensor → agent, traces→corpus. Lint-clean test un-skipped + green.
- Source is always a **2.4 install** (`.keystone/`, old vocab). Fix the target map: `guide→guide` (stays), `sensor→sensor` (stays), `action→command`, `playbook→playbook` (stays), `persona→agent` + `subagent→agent`. Migrate 2.4 `kind: rule` → `guide` (`mode: inferential`); 2.4 sensor `host_triggers:` → `hook` primitives (the framework hook layer). Keep `.keystone→.harness` relocation, `traces→corpus`, seed documents, `.harness/work/`.
- Infer `mode:` on migrate: sensor with `host_triggers:` → computational; `review-*`/`code-debt` → inferential; guide → inferential default.
- Down: reverse to 2.4 kinds/paths (best-effort).
- → verify: full 2.4-install fixture upgrades; every primitive at corrected path/kind + `mode` set; re-indexes clean.

**Slice 7 — embedded scaffold templates → corrected vocab** (`scaffold/templates/`) — DONE
- actions→commands, personas→agents, traces→corpus; dropped sources/ + rules/ source dirs. Sensors classified: computational dev-checks (build/test/lint/coverage/type-check/sast/secret-scan/vuln-scan/stack-drift/commit-message) → `hooks/` (mode:computational, event, placeholder run); inferential reviews/snapshots → `sensors/` (mode:inferential, returns). `keystone-new-action` skill → `keystone-new-command`. Fresh `keystone init` is lint-clean; golden test updated + green.
- Restore `guides/ sensors/ playbooks/ agents/` template dirs + `command/`; add `mode:`; one `pattern/` + `posture/` + `tool/` seed each + one framework-event `sensor/` seed.
- → verify: `TestInit_FreshScaffoldGoldenFiles` green against `.harness/` + corrected vocab.

**Slice 8 — dogfood migration + reinstall** (run, don't code)
- `go build`, run `keystone migrate up` on this repo's `.keystone/` → `.harness/` corrected vocab. `keystone index && keystone project`. Then `go install ./cmd/keystone` (hooks repoint).
- → verify: no `actions/playbooks/personas` dirs left under `.harness/`; `keystone verify` clean; hooks fire with new binary.

**Slice 9 — docs** (CLAUDE.md, MCP help, README, menu template) — DONE
- Speak corrected vocab + `mode`/`pattern`/`posture`/framework hooks. Reconcile aliases (workflow=playbook, plugin=policy).
- → verify: no stale `rule`-as-guide / `agent`-as-subagent language; MCP help matches `KnownKinds`.

**Slice 10 — frontmatter lowering** — DONE
- Projection emits ONLY host-native keys (name/description/allowed-tools/model/globs/alwaysApply); strips framework fields.
- → verify: projected host files carry no keystone-only frontmatter (`corpus`/`includes`/`event`/`mode`/`gates`…).

## Sequencing / gotchas
- Code slices (1–6) before dogfood migration (7). The **installed binary backs the hooks** — keep it 2.4-consistent until slice 7 migrates the dogfood, else hooks + `.keystone/` content disagree.
- `keystone.json` stays at repo root. INDEX + lockfile + work + primitives under `.harness/`.
- Migration legacy-detection (`.keystone/harness`, `.keystone/lockfile.json`) stays `.keystone` — it identifies pre-3.0 installs.
- Temp prefixes (`.keystone-project.*`) are tool-internal, not the harness location — leave them.

## Acceptance (from domain.md)
- `KnownKinds` = corrected set; `mode:` field live; type-aware projection working.
- `keystone migrate up` upgrades a 2.4 install to corrected 3.0 vocab; `verify` clean.
- `go test ./...` + `go vet ./...` green.
- Dogfood fully migrated — no `action/playbook→skill`-collapse or `persona`/`agent` artifacts.
- Docs speak the corrected vocabulary.
